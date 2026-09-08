# Velryn — Agent Service

The **agent-service** is the decision-making brain for Velryn's NPCs. It's a
small FastAPI app that the Go simulation (`../engine`) calls over HTTP every
time an NPC (or the world itself) needs to decide what happens next.

It exists as a separate service — not a Go package — because this is where
all the LLM/agent complexity actually lives: LangGraph state machines, a
multi-provider fallback chain, and MCP tool calling. Keeping that in Python
means using LangChain/LangGraph's ecosystem directly instead of
reimplementing it in Go.

---

## How it fits into Velryn

```text
engine (Go)  ──HTTP──▶  agent-service (this)  ──stdio/MCP──▶  mcp/
   world state              LangGraph +              action-formatting
   perception                 LLM router                 tools only
```

The engine never talks to an LLM provider for NPC decisions directly — it
sends a `Perception` (what the NPC can currently see/feel/remember) and gets
back a structured `AgentAction` (`MOVE`, `TALK`, `ATTACK`, ...). See the
[root README](../README.md) for the full three-service picture.

## Endpoints

| Route             | Called by                              | Purpose                                                        |
|--------------------|-----------------------------------------|------------------------------------------------------------------|
| `GET /health`       | anyone                                  | liveness check                                                   |
| `POST /decide`       | engine, one NPC at a time                | single NPC decision                                               |
| `POST /decide-batch`  | engine, for NPCs sharing a scene          | decides a whole group concurrently, aware of what others just said |
| `POST /world-decide`   | engine, periodically (the "Game Director") | decides world-level events: weather shifts, monster invasions, etc. |

## Project layout

```text
agent-service/
├── app/
│   ├── main.py               # FastAPI app + the 4 routes above, nothing else
│   ├── rules.py                # tiny pure-Python rules (e.g. is_in_danger)
│   ├── schemas/
│   │   ├── npc.py               # Perception / AgentAction — the NPC decision contract
│   │   └── world.py             # WorldSnapshot / WorldEvent — the world-agent contract
│   ├── prompts/
│   │   └── npc_prompt.py         # world lore, action guide, per-NPC prompt builder
│   ├── llm/
│   │   └── router.py              # provider instances + fallback/cooldown logic
│   ├── graphs/
│   │   ├── npc_graph.py            # LangGraph: one NPC's decide → (maybe) MCP tool call → action
│   │   ├── group_graph.py           # runs npc_graph concurrently for a whole scene
│   │   └── world_graph.py            # the Game Director's own (simpler, tool-less) chain
│   └── tools/
│       └── mcp_client.py             # spawns and talks to the mcp/ subprocess
├── pyproject.toml
├── uv.lock
└── .env.example
```

Every subfolder is one layer of the same pipeline: `schemas` (what goes in
and out), `prompts` (what the LLM reads), `llm` (which provider actually
answers), `graphs` (the LangGraph wiring that ties those together), `tools`
(the MCP bridge). `main.py` only exposes that pipeline over HTTP.

### Why LangGraph instead of a single LLM call

Each NPC decision goes through a tiny 3-node graph (`app/graphs/npc_graph.py`):

```text
call_model ──(tool call?)──▶ run_tool ──▶ extract_action
     │
     └──(no tool call)──────────────────▶ extract_action
```

`call_model` gives the LLM the NPC's situation *and* the MCP action tools
(`move`, `talk`, `attack`, ...). If it calls one, `run_tool` executes it
(formatting only — see `mcp/`) and `extract_action` turns the result into a
validated `AgentAction`. If the model answers with plain text instead of
calling a tool, that's treated as "didn't decide anything actionable" and
the NPC just waits — the engine never receives an invalid or missing action.

### The provider fallback chain

`app/llm/router.py` tries, in order: **Groq → Gemini → Cerebras → Mistral →
local Ollama** (last resort — much slower with no GPU, so it only gets
consulted once every cloud provider has already failed). A failing provider
enters a cooldown (from a parsed `retry-after` when the error gives one,
otherwise a heuristic based on the error type) instead of being retried
every single NPC decision. All calls share one global concurrency cap
(`MAX_CONCURRENT_LLM_CALLS`) so a busy tick can't fan out into dozens of
simultaneous provider requests.

## Running locally

Requires [`uv`](https://docs.astral.sh/uv/) and Python 3.13+.

```bash
cd agent-service
cp .env.example .env        # fill in at least one provider key
uv sync
uv run uvicorn app.main:app --reload --port 8000
```

The engine expects this service at `AGENT_SERVICE_URL` (see
`../engine/.env.example`), which defaults to `http://localhost:8000`.

### Observability (LangSmith)

Every LangGraph run is auto-traced with **zero code changes** — LangChain
picks up tracing purely from environment variables. Set these in `.env` and
every `/decide`, `/decide-batch` and `/world-decide` call shows up as a full
trace (including which MCP tool got called and why) in your LangSmith
project:

```env
LANGSMITH_TRACING=true
LANGSMITH_API_KEY=...
LANGSMITH_PROJECT=velryn
```

Leave `LANGSMITH_TRACING` unset (or `false`) to run with no tracing at all.

## Testing

```bash
uv sync --extra dev
uv run pytest
```

47 tests total across this project and `mcp/` (see `mcp/README.md`).
`tests/test_rules.py`, `tests/test_schemas_npc.py` and
`tests/test_schemas_world.py` cover the pure logic and the Pydantic
contracts in isolation. `tests/test_llm_router.py` covers the
retry-after-parsing and cooldown-classification logic in
`app/llm/router.py` without making any real provider calls. `tests/test_main_api.py`
exercises the actual FastAPI routes with `TestClient`, including
confirming that `/decide` degrades to a `WAIT` action (HTTP 200, not 500)
when every provider in the fallback chain is unreachable.

## MCP tool subprocess

`app/tools/mcp_client.py` launches `../mcp/src/velryn_mcp/server.py` as a
subprocess (stdio transport) the first time `/decide` is called, and reuses
that connection afterward. See [`mcp/README.md`](../mcp/README.md) for what
that server actually exposes and why it's a separate project instead of
being inlined here.
