# Velryn — MCP Action Server

A small [MCP](https://modelcontextprotocol.io) server that exposes the set
of actions an NPC can take (`move`, `talk`, `attack`, `buy`, ...) as
discoverable tools, over the standard MCP stdio transport.

## Why this is a separate service

`agent-service` could have defined these as plain LangChain tools inline.
They're an MCP server instead specifically to *demonstrate* MCP — the tool
schema (name, args, docstring) is exactly the same either way, but exposing
it over MCP means any MCP-speaking client (not just this project's
LangGraph agent) could reuse the same action set.

## What it does — and does not — do

Each tool only **formats** an intention as a dict, e.g.:

```python
@mcp.tool()
def attack(target: str, skill: Optional[str] = None) -> dict:
    payload = {"action": "ATTACK", "target": target}
    if skill:
        payload["skill"] = skill
    return payload
```

No tool here touches the game world, the database, or the Go engine. The
only thing that ever actually *applies* an action to the world is
`engine/engine/actions.go`'s `resolveIntent` — this server's job ends at
producing a well-formed `AgentAction`-shaped dict that `agent-service` reads
back and returns to the engine over HTTP.

## Tools

| Tool | Maps to `AgentAction.action` |
|---|---|
| `move(x, y)` | `MOVE` (single adjacent step) |
| `set_goal(x, y)` | `MOVE` with `goal_coordinates` (multi-turn pathing) |
| `talk(target, dialogue, trust_shift=0)` | `TALK` |
| `attack(target, skill=None)` | `ATTACK` |
| `defend()` | `DEFEND` |
| `flee()` | `FLEE` |
| `buy(merchant, item)` | `BUY` |
| `sell(merchant, item)` | `SELL` |
| `trade(target, item=None, gold=None)` | `TRADE` |
| `use_item(item)` | `USE` |
| `wait()` | `WAIT` |

This list must stay in sync with `models.ActionVerb` in
`engine/models/action.go` and the `ActionVerb` Literal in
`agent-service/app/schemas/npc.py` — there's no codegen tying the three
together, so a new action verb has to be added by hand in all three places.

## Running locally

Requires [`uv`](https://docs.astral.sh/uv/) and Python 3.13+.

```bash
cd mcp
uv sync
```

You normally don't run this server by hand — `agent-service` spawns it
automatically as a subprocess the first time it needs the action tools (see
`agent-service/app/tools/mcp_client.py`). To run and inspect it standalone
(useful for debugging tool schemas):

```bash
uv run velryn-mcp
```

or, with the [MCP inspector](https://modelcontextprotocol.io/legacy/tools/inspector):

```bash
npx @modelcontextprotocol/inspector uv run velryn-mcp
```

## Testing

```bash
uv sync --extra dev
uv run pytest
```

17 tests in `tests/test_server.py`, one per tool plus a couple of edge
cases (trust_shift clamping, trade's item-vs-gold-only shapes). Since
`@mcp.tool()` leaves the underlying function directly callable, the tests
just import and call `move`, `attack`, `talk`, etc. like any other Python
function — no MCP client, no subprocess, no server startup needed.
