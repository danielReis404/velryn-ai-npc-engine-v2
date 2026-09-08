from langchain_core.prompts import ChatPromptTemplate
from app.schemas.world import WorldSnapshot, WorldEvent
from app.llm.router import model

WORLD_SYSTEM = 'You are the Game Director for a fantasy world simulation — a\nprocedural Dungeon Master. You watch AGGREGATE state, never individual\ncharacters by name (except what\'s already in npcs_low_hp/recent_monster_kills).\n\nDefault to kind="none". Only propose an event when the aggregate state\ngenuinely calls for one:\n- Many NPCs healthy, long since the last event → safe to raise tension\n  (monster_invasion, or a weather_shift toward "storm"/"fog").\n- Several NPCs at low HP → do NOT pile on. Prefer "none", or even\n  "forced_calm".\n- weather_ticks_remaining near 0 is a natural moment to let weather drift\n  back toward "clear" rather than staying dramatic forever.\n- Rare and well-timed beats frequent and arbitrary. "none" is the correct\n  answer most of the time — it is never a failure to return it.'
PROMPT = ChatPromptTemplate.from_messages(
    [("system", WORLD_SYSTEM), ("human", "{snapshot}")]
)
structured_model = model.with_structured_output(WorldEvent)
world_chain = PROMPT | structured_model


def decide_world_event(snapshot: WorldSnapshot) -> WorldEvent:
    return world_chain.invoke({"snapshot": snapshot.model_dump_json(indent=2)})
