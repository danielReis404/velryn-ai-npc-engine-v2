import asyncio
from typing import TypedDict
from langgraph.graph import END, StateGraph
from langchain_core.messages import HumanMessage, SystemMessage
from app.prompts.npc_prompt import SYSTEM, build_prompt
from app.rules import is_in_danger
from app.schemas.npc import AgentAction, Perception
from app.graphs.npc_graph import compiled_graph


class GroupState(TypedDict):
    queue: list[Perception]
    scene_log: list[str]
    actions: dict[str, AgentAction]


async def _decide_member(perception: Perception, scene_context: str) -> AgentAction:
    danger_prefix = ""
    if is_in_danger(perception):
        danger_prefix = (
            "\n⚠️ You are in immediate danger — you MUST call attack or flee now.\n"
        )
    prompt_text = (
        build_prompt(perception)
        + danger_prefix
        + f"\n\nWhat just happened in this exact scene, moments ago:\n{scene_context}"
        + f"\n\nReminder: no matter what you just read above, YOU are {perception.name}, a {perception.role} — not any of the other characters mentioned. Never copy another character's words, claim their name, or repeat their dialogue as your own. React as {perception.name} would, in {perception.name}'s own voice."
    )
    messages = [SystemMessage(content=SYSTEM), HumanMessage(content=prompt_text)]
    result = await compiled_graph.ainvoke(
        {"perception": perception, "messages": messages, "action": None}
    )
    return result["action"]


async def decide_group(state: GroupState) -> GroupState:
    if not state["queue"]:
        return state
    scene_context = (
        "\n".join(state["scene_log"]) or "Nothing has happened in this exchange yet."
    )
    tasks = [
        asyncio.create_task(_decide_member(perception, scene_context))
        for perception in state["queue"]
    ]
    results = await asyncio.gather(*tasks, return_exceptions=True)
    actions = dict(state["actions"])
    for perception, result in zip(state["queue"], results):
        if isinstance(result, Exception):
            action = AgentAction(
                internal_thought="My usual decision process was unavailable, so I will wait.",
                action="WAIT",
            )
            print(f"[AGENT FALLBACK] {perception.name}: {result}")
        else:
            action = result
        actions[perception.name] = action
        log_line = f"{perception.name}: {action.internal_thought}"
        if action.dialogue:
            log_line += f' — says: "{action.dialogue}"'
        print(f"[CENA] {log_line}")
    return {"queue": [], "scene_log": state["scene_log"], "actions": actions}


group_graph = StateGraph(GroupState)
group_graph.add_node("decide_group", decide_group)
group_graph.set_entry_point("decide_group")
group_graph.add_edge("decide_group", END)
compiled_group_graph = group_graph.compile()
