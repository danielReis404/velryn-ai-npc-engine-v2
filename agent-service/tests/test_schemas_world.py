import pytest
from pydantic import ValidationError

from app.schemas.world import WorldEvent, WorldSnapshot


def _base_snapshot(**overrides) -> dict:
    defaults = dict(
        turn=42,
        npcs_alive=8,
        ticks_since_last_event=5,
        current_weather="clear",
        weather_ticks_remaining=0,
        day_count=1,
        time_of_day="day",
    )
    defaults.update(overrides)
    return defaults


def test_world_snapshot_defaults_empty_lists():
    snapshot = WorldSnapshot(**_base_snapshot())
    assert snapshot.npcs_low_hp == []
    assert snapshot.recent_monster_kills == []
    assert snapshot.active_regions == []


def test_world_snapshot_rejects_an_unknown_weather():
    with pytest.raises(ValidationError):
        WorldSnapshot(**_base_snapshot(current_weather="tornado"))


def test_world_event_defaults_to_no_optional_fields():
    event = WorldEvent(kind="none")
    assert event.description == ""
    assert event.region is None
    assert event.spawn_monster_species is None
    assert event.new_weather is None


def test_world_event_monster_invasion_carries_spawn_details():
    event = WorldEvent(
        kind="monster_invasion",
        description="A pack of wolves crosses into the Wildlands.",
        region="Wildlands",
        spawn_monster_species="Shadow Wolf",
        spawn_count=3,
    )
    assert event.kind == "monster_invasion"
    assert event.spawn_count == 3


def test_world_event_rejects_an_unknown_kind():
    with pytest.raises(ValidationError):
        WorldEvent(kind="dragon_attack")
