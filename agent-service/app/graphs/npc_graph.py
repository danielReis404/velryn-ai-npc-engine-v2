import json
from typing import TypedDict, Annotated
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langchain_core.messages import SystemMessage, HumanMessage, AIMessage
from app.schemas.npc import Perception, AgentAction
from app.prompts.npc_prompt import build_prompt, SYSTEM
from app.rules import is_in_danger
from app.tools.mcp_client import get_action_tools
from app.llm.router import invoke_agent_model


class AgentState(TypedDict):
    perception: Perception
    messages: Annotated[list, add_messages]
    action: AgentAction | None


async def call_model(state: AgentState) -> dict:
    if not state["messages"]:
        situation = build_prompt(state["perception"])
        danger_prefix = ""
        if is_in_danger(state["perception"]):
            danger_prefix = (
                "\n⚠️ You are in immediate danger — you MUST call attack or flee now.\n"
            )
        messages = [SystemMessage(SYSTEM), HumanMessage(situation + danger_prefix)]
    else:
        messages = state["messages"]
    tools = await get_action_tools()
    response = await invoke_agent_model(messages, tools=tools)
    return {"messages": [response]}


async def run_tool(state: AgentState) -> dict:
    last = state["messages"][-1]
    tools = await get_action_tools()
    tools_by_name = {t.name: t for t in tools}
    call = last.tool_calls[0]
    tool = tools_by_name[call["name"]]
    result = await tool.ainvoke(call["args"])
    if (
        isinstance(result, list)
        and result
        and isinstance(result[0], dict)
        and ("text" in result[0])
    ):
        content_text = result[0]["text"]
    elif isinstance(result, str):
        content_text = result
    else:
        content_text = json.dumps(result)
    from langchain_core.messages import ToolMessage

    return {"messages": [ToolMessage(content=content_text, tool_call_id=call["id"])]}


def extract_action(state: AgentState) -> dict:
    last = state["messages"][-1]
    if isinstance(last, AIMessage) and (not last.tool_calls):
        return {
            "action": AgentAction(
                internal_thought=last.content or "hesitated without calling an action",
                action="WAIT",
            )
        }
    payload = json.loads(last.content)
    ai_message = state["messages"][-2]
    thought = (
        ai_message.content or f"Performed the action: {payload.get('action', 'WAIT')}"
    )
    return {"action": AgentAction(internal_thought=thought, **payload)}


def has_tool_call(state: AgentState) -> str:
    last = state["messages"][-1]
    if isinstance(last, AIMessage) and last.tool_calls:
        return "run_tool"
    return "extract_action"


graph = StateGraph(AgentState)
graph.add_node("call_model", call_model)
graph.add_node("run_tool", run_tool)
graph.add_node("extract_action", extract_action)
graph.set_entry_point("call_model")
graph.add_conditional_edges(
    "call_model",
    has_tool_call,
    {"run_tool": "run_tool", "extract_action": "extract_action"},
)
graph.add_edge("run_tool", "extract_action")
graph.add_edge("extract_action", END)
compiled_graph = graph.compile()
