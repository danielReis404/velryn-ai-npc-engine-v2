from typing import Literal, Optional
from pydantic import BaseModel

WorldEventKind = Literal[
    "monster_invasion",
    "weather_shift",
    "resource_scarcity",
    "forced_calm",
    "great_creature_sighting",
    "none",
]
Weather = Literal["clear", "rain", "storm", "fog"]


class WorldSnapshot(BaseModel):
    turn: int
    npcs_alive: int
    npcs_low_hp: list[str] = []
    recent_monster_kills: list[str] = []
    ticks_since_last_event: int
    active_regions: list[str] = []
    current_weather: Weather
    weather_ticks_remaining: int
    day_count: int
    time_of_day: Literal["day", "night"]


class WorldEvent(BaseModel):
    kind: WorldEventKind
    description: str = ""
    region: Optional[str] = None
    spawn_monster_species: Optional[str] = None
    spawn_count: Optional[int] = None
    new_weather: Optional[Weather] = None
    weather_duration_ticks: Optional[int] = None
