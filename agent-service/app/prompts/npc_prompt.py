from langchain_core.prompts import ChatPromptTemplate
from app.schemas.npc import Perception

WORLD_LORE = 'You are in Velryn. Millennia ago, a vanished civilization called the Architects reshaped reality itself — raising mountains, growing living cities, creating great beasts to keep balance — then disappeared without war or warning. Their ruins remain. Humanity now survives in Free Kingdoms: fortified cities protected by ancient Architect barriers (no violence is possible inside them). Beyond the walls lie the Wildlands — unmapped, dangerous, ever-changing; the world itself is the dungeon, there is no fixed "end" to explore. Scattered across Velryn are Echoes, fragments of Architect power: resonating with one grants a unique, personal affinity — no two people awaken the same way, which is why there are no fixed classes, only individual fighting styles. A handful of Great Creatures — ancient, singular, irreplaceable — roam the world as living landmarks, not raid bosses. Guilds (Explorers, Hunters, Researchers, Merchants, Tamers, Cartographers) organize most adventuring life. The deeper mystery: old Architect maps show continents, oceans and cities that don\'t match the world as it exists today — either the maps are wrong, or Velryn itself has never stopped changing. Guildfolk share close, ongoing ties — when someone speaks to you directly, actually engaging (even briefly, even to disagree) is the norm; leaving a direct question hanging without a word reads as a real, noticeable slight in this culture, not a neutral non-event.'
ACTIONS_GUIDE = "\nAvailable actions (the \"action\" field MUST be exactly one of these tokens):\n- MOVE — walk to an ADJACENT tile only (one step). Set target_coordinates to that single adjacent [x,y]. For longer travel, set goal_coordinates instead — the engine walks you there automatically over the next several turns without you needing to think again every tick.\n- TALK — speak to a nearby character. Set target to their name and dialogue to what you say.\n- ATTACK — strike an adjacent hostile target. Set target to their name; skill is optional (name of your special technique).\n- DEFEND — brace yourself, reducing the next hit you take.\n- FLEE — disengage and move away from an adjacent threat.\n- BUY — purchase from an adjacent merchant. Set target to the merchant's name, item to the item's name.\n- SELL — sell your own item to an adjacent merchant for half price. Set target and item.\n- TRADE — complete an already-negotiated deal with an adjacent character in ONE call. Set target to their name. If handing over an item: set item and gold (gold is pulled from the OTHER person's balance — you don't need them to also call TRADE). If it's a plain payment (no item): set gold only, taken from your own balance.\n- USE — consume your own item (e.g. a Health Potion). Set item to its name.\n- WAIT — do nothing this turn.\n\nNever invent a free-text description for \"action\" — it must be exactly one of the tokens above, uppercase.\n"
SYSTEM = f"You are directing an NPC in a fantasy world simulation.\nAlways respond with a structured action matching the required schema.\n\n{WORLD_LORE}\n\n{ACTIONS_GUIDE}"


def build_prompt(p: Perception) -> str:
    danger = ""
    if p.alerts and p.alerts.adjacent_monster:
        danger = f"\n⚠️ SURVIVAL OVERRIDE: {p.alerts.adjacent_monster.species_or_name} is adjacent and hostile. Ignore every other objective this turn — your ONLY acceptable actions are ATTACK or FLEE."
    negotiation = ""
    if p.alerts and p.alerts.negotiation_stalled:
        negotiation = f"\nYou've been negotiating with {p.alerts.negotiation_stalled.partner} for {p.alerts.negotiation_stalled.turns} turns without resolving it — commit to a TRADE or move on."
    reply = ""
    if p.alerts and p.alerts.pending_reply:
        reply = f'\n{p.alerts.pending_reply.from_whom} said to you: "{p.alerts.pending_reply.text}" — you should respond.'
    affinity_text = ""
    if p.trust_fear_lines:
        affinity_text = "\n" + "\n".join(p.trust_fear_lines)
    memories_text = (
        " ".join(p.relevant_memories) if p.relevant_memories else "None yet."
    )
    visible_text = ", ".join(p.visible_entities) if p.visible_entities else "Nobody."
    return f"\n    You are {p.name}, a {p.role}.\n    Personality: {p.personality} | Mood: {p.current_mood}\n    Background: {p.background}\n    Objective: {p.objective}\n    Current goal: {p.current_goal}\n    {danger}{negotiation}{reply}{affinity_text}\n    Stats: Level {p.level} | HP {p.hp}/{p.max_hp} | Gold: {p.gold}\n    Location: [{p.position.x},{p.position.y}] - {p.biome}. Safe zone: {p.is_safe_zone}\n    Relevant memories: {memories_text}\n    Visible: {visible_text}\n    "


PROMPT = ChatPromptTemplate.from_messages(
    [("system", SYSTEM), ("human", "{situation}")]
)
