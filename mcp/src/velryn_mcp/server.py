"""
Velryn MCP server: exposes the actions a character can take as tools
discoverable over the MCP protocol (JSON-RPC). It never talks to the Go
engine directly — each tool only formats the intention; the Go engine is
always the one that actually applies it, on the other side of agent-service.
"""

from typing import Optional
from mcp.server.mcpserver import MCPServer

mcp = MCPServer("velryn-actions")


@mcp.tool()
def move(x: int, y: int) -> dict:
    """Set [x, y] as the next short-term destination -- a single step or a
    continuous goal, depending on how far away it is."""
    return {"action": "MOVE", "target_coordinates": [x, y]}


@mcp.tool()
def set_goal(x: int, y: int) -> dict:
    """Set [x, y] as a long-term destination -- the Go engine walks there on
    its own over the next turns, without needing to think again every tick."""
    return {"action": "MOVE", "goal_coordinates": [x, y]}


@mcp.tool()
def talk(target: str, dialogue: str, trust_shift: int = 0) -> dict:
    """Talk to a nearby character. trust_shift (-15 to 15, optional) declares
    how THIS line changes how much you trust target: positive if it brought
    you closer, negative if it annoyed/raised suspicion, 0 (default) if it
    was neutral."""
    payload = {"action": "TALK", "target": target, "dialogue": dialogue}
    if trust_shift:
        payload["trust_shift"] = max(-15, min(15, trust_shift))
    return payload


@mcp.tool()
def attack(target: str, skill: Optional[str] = None) -> dict:
    """Attack an adjacent, hostile target. skill is optional -- the name of
    a technique from your own ability list."""
    payload = {"action": "ATTACK", "target": target}
    if skill:
        payload["skill"] = skill
    return payload


@mcp.tool()
def defend() -> dict:
    """Brace yourself, reducing the next hit you take."""
    return {"action": "DEFEND"}


@mcp.tool()
def flee() -> dict:
    """Flee from an adjacent threat."""
    return {"action": "FLEE"}


@mcp.tool()
def buy(merchant: str, item: str) -> dict:
    """Buy an item from an adjacent merchant."""
    return {"action": "BUY", "target": merchant, "item": item}


@mcp.tool()
def sell(merchant: str, item: str) -> dict:
    """Sell one of your own items to an adjacent merchant, for half its
    price."""
    return {"action": "SELL", "target": merchant, "item": item}


@mcp.tool()
def trade(target: str, item: Optional[str] = None, gold: Optional[int] = None) -> dict:
    """Close a deal already agreed on in conversation with another
    character, in a single call. If you're handing over an item, provide
    both item and gold (the gold comes out of the OTHER person's balance).
    If it's just a payment with no item, provide only gold (comes out of
    your own balance)."""
    payload = {"action": "TRADE", "target": target}
    if item:
        payload["item"] = item
    if gold:
        payload["gold"] = gold
    return payload


@mcp.tool()
def use_item(item: str) -> dict:
    """Consume an item from your own inventory (e.g. a potion)."""
    return {"action": "USE", "item": item}


@mcp.tool()
def wait() -> dict:
    """Do nothing this turn."""
    return {"action": "WAIT"}


def main() -> None:
    mcp.run()


if __name__ == "__main__":
    main()
