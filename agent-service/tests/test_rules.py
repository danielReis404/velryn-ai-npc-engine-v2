from app.rules import is_in_danger
from app.schemas.npc import AdjacentMonsterAlert, Perception, PerceptionAlerts


def test_is_in_danger_true_when_monster_adjacent_outside_safe_zone(
    base_perception_kwargs,
):
    perception = Perception(
        **base_perception_kwargs,
        alerts=PerceptionAlerts(
            adjacent_monster=AdjacentMonsterAlert(
                species_or_name="Stone Boar", in_safe_zone=False
            )
        ),
    )
    assert is_in_danger(perception) is True


def test_is_in_danger_false_inside_safe_zone_even_with_a_monster_alert(
    base_perception_kwargs,
):
    base_perception_kwargs["is_safe_zone"] = True
    perception = Perception(
        **base_perception_kwargs,
        alerts=PerceptionAlerts(
            adjacent_monster=AdjacentMonsterAlert(
                species_or_name="Stone Boar", in_safe_zone=True
            )
        ),
    )
    assert is_in_danger(perception) is False


def test_is_in_danger_false_with_no_alerts_object_at_all(base_perception_kwargs):
    perception = Perception(**base_perception_kwargs, alerts=None)
    assert is_in_danger(perception) is False


def test_is_in_danger_false_when_alerts_present_but_no_monster(base_perception_kwargs):
    perception = Perception(**base_perception_kwargs, alerts=PerceptionAlerts())
    assert is_in_danger(perception) is False
