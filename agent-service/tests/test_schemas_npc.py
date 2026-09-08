import pytest
from pydantic import ValidationError

from app.schemas.npc import AgentAction, Coordinate, Perception


def test_coordinate_accepts_the_go_engine_field_casing():
    coord = Coordinate.model_validate({"X": 3, "Y": 7})
    assert coord.x == 3
    assert coord.y == 7


def test_perception_round_trips_from_go_style_json(base_perception_kwargs):
    perception = Perception(**base_perception_kwargs)
    assert perception.position.x == 5
    assert perception.position.y == 5
    assert perception.alerts is None


def test_agent_action_rejects_an_unknown_verb():
    with pytest.raises(ValidationError):
        AgentAction(internal_thought="do a barrel roll", action="BARREL_ROLL")


def test_agent_action_accepts_every_known_verb():
    for verb in (
        "MOVE",
        "TALK",
        "ATTACK",
        "DEFEND",
        "FLEE",
        "BUY",
        "SELL",
        "TRADE",
        "USE",
        "WAIT",
    ):
        action = AgentAction(internal_thought="...", action=verb)
        assert action.action == verb


def test_agent_action_optional_fields_default_to_none():
    action = AgentAction(internal_thought="waiting quietly", action="WAIT")
    assert action.target is None
    assert action.target_coordinates is None
    assert action.trust_shift is None
