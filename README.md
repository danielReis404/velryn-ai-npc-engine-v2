# Velryn

**Velryn** is an experimental persistent-world simulation where NPCs are
driven by LLM agents instead of scripted dialogue trees. It's a learning
project exploring how far you can push autonomous, tool-using AI agents
inside a deterministic game loop — combat, movement, and economy stay
predictable and instant; only genuinely ambiguous decisions ("what do I do
right now, given what I see and remember?") go through an LLM.

> The world was never meant to be understood.

This repository is a monorepo with **three independent services**:

```text
velryn/
├── engine/          Go — the game world itself: state, ticks, combat,
│                    movement, memory, HTTP/SSE server + frontend
│
├── agent-service/   Python — the decision-making brain: FastAPI + LangGraph,
│                    calls the LLM providers, calls the MCP tools below
│
└── mcp/             Python — an MCP server exposing the NPC action set
                     (move, talk, attack, buy, ...) as discoverable tools
```

```text
 ┌────────────┐   HTTP    ┌──────────────────┐   MCP (stdio)   ┌────────┐
 │   engine   │ ────────▶ │  agent-service    │ ──────────────▶ │  mcp   │
 │    (Go)    │ ◀──────── │ FastAPI+LangGraph │ ◀────────────── │(Python)│
 └────────────┘  action   └──────────────────┘   tool result   └────────┘
       │                            │
       ▼                            ▼
  world state,                 LLM providers:
  SSE to browser              Groq → Gemini → Cerebras
                                  → Mistral → Ollama
```

Each service has its own README with the actual detail:

- **[`engine/README.md`](engine/README.md)** — the simulation architecture, NPC decision flow, deterministic vs. AI-driven mechanics, how memory persistence works.
- **[`agent-service/README.md`](agent-service/README.md)** — the LangGraph state machines, the provider fallback chain, LangSmith tracing.
- **[`mcp/README.md`](mcp/README.md)** — the action tool set and why it's exposed over MCP instead of being inline LangChain tools.

## Why three services instead of one

Each one is doing a fundamentally different job, in the language best suited
to it:

- `engine` needs to be fast, deterministic, and stateful — a tight Go loop
  ticking a live world is a better fit than a Python process doing the same.
- `agent-service` needs LangChain/LangGraph's ecosystem (structured output,
  multi-provider fallback, tool calling) — reimplementing that in Go would
  mean reinventing most of LangGraph by hand.
- `mcp` is deliberately separate from `agent-service` even though nothing
  forces it to be — it exists specifically to demonstrate the Model Context
  Protocol as its own reusable server, decoupled from any one agent
  framework.

The trade-off is an HTTP hop (engine → agent-service) and a stdio hop
(agent-service → mcp) per decision, instead of one in-process call. For a
simulation ticking every several seconds with a handful of NPCs, that's a
non-issue; it wouldn't be the right call for something needing sub-millisecond
decisions.

## Running locally (all three services)

You need three terminals. Start them in this order — each one is usable
standalone, but `engine` won't get AI-driven decisions without
`agent-service`, and `agent-service` won't get past its first tool call
without `mcp` (which it launches automatically, so you don't run this one
by hand under normal use).

```bash
# 1. agent-service — the decision-making brain
cd agent-service
cp .env.example .env   # fill in at least one LLM provider key
uv sync
uv run uvicorn app.main:app --reload --port 8000

# 2. engine — the game world (new terminal)
cd engine
cp .env.example .env   # AGENT_SERVICE_URL defaults to localhost:8000, fine as-is
go mod tidy
go run main.go
```

Then open **http://localhost:8080** to watch the world — NPCs will start
thinking as soon as a viewer connects (see "Real-time observation" in the
engine README for why).

The `mcp/` server does not need to be started manually: `agent-service`
spawns it as a subprocess on first use.

## Testing

Each service tests independently — there's no end-to-end test harness
across all three (that would need a live LLM provider or a lot of mocking
for little extra confidence over testing each boundary in isolation).

```bash
cd engine && go test ./...          # Go: geometry, combat, perception, resolveIntent
cd agent-service && uv sync --extra dev && uv run pytest   # 30 tests
cd mcp && uv sync --extra dev && uv run pytest             # 17 tests
```

See each service's README for what's actually covered and, just as
important, what isn't (things like the live SSE stream or an actual round
trip through a real LLM provider are integration-level concerns, not unit
tests).

## Project status

This is a learning project, not a production game. Expect narrative
inconsistencies, occasional bad LLM decisions, and code that prioritizes
"can I learn something building this" over "is this the most efficient
architecture possible." See each service's own README for its specific
disclaimers and known rough edges.

## Before you fork/clone this to run with your own keys

Never commit `.env` files (all three services' `.gitignore` — actually just
the one root `.gitignore` — already excludes them). Copy the matching
`.env.example` in each service folder and fill in your own credentials.
