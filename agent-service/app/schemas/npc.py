from typing import Literal
from pydantic import BaseModel, Field, ConfigDict

ActionVerb = Literal[
    "MOVE", "TALK", "ATTACK", "DEFEND", "FLEE", "BUY", "SELL", "TRADE", "USE", "WAIT"
]


class Coordinate(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    x: int = Field(alias="X")
    y: int = Field(alias="Y")


class WeaponInfo(BaseModel):
    name: str
    power: int
    durability: int
    max_durability: int


class MerchantInfo(BaseModel):
    name: str
    stock: list[str] | None = None


class AdjacentMonsterAlert(BaseModel):
    species_or_name: str
    in_safe_zone: bool


class NegotiationAlert(BaseModel):
    partner: str
    turns: int


class PendingReplyAlert(BaseModel):
    model_config = ConfigDict(populate_by_name=True)
    from_whom: str = Field(alias="from")
    text: str


class PerceptionAlerts(BaseModel):
    adjacent_monster: AdjacentMonsterAlert | None = None
    negotiation_stalled: NegotiationAlert | None = None
    pending_reply: PendingReplyAlert | None = None


class Perception(BaseModel):
    name: str
    role: str
    personality: str
    faction: str
    current_mood: str
    background: str
    objective: str
    current_goal: str | None = ""
    relationships: list[str] | None = None
    likes: list[str] | None = None
    dislikes: list[str] | None = None
    known_secrets: list[str] | None = None
    level: int
    hp: int
    max_hp: int
    base_attack: int
    weapon_info: WeaponInfo | None = None
    combat_style: str | None = None
    skills: list[str] | None = None
    gold: int
    inventory_names: list[str] | None = None
    nearby_merchant: MerchantInfo | None = None
    position: Coordinate
    biome: str
    tile_description: str
    is_safe_zone: bool
    nearby_pois: list[str] | None = None
    trust_fear_lines: list[str] | None = None
    visible_entities: list[str] | None = None
    relevant_memories: list[str] | None = None
    alerts: PerceptionAlerts | None = None


class AgentAction(BaseModel):
    internal_thought: str
    action: ActionVerb
    target_coordinates: list[int] | None = None
    target: str | None = None
    dialogue: str | None = None
    skill: str | None = None
    goal_coordinates: list[int] | None = None
    item: str | None = None
    gold: int | None = None
    trust_shift: int | None = None


class GroupDecideRequest(BaseModel):
    members: list[Perception]


class GroupDecideResponse(BaseModel):
    actions: dict[str, AgentAction]
