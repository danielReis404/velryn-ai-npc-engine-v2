import pytest


@pytest.fixture
def base_perception_kwargs():
    return dict(
        name="Selene",
        role="Relic Hunter",
        personality="cautious",
        faction="none",
        current_mood="alert",
        background="",
        objective="find the White Stag",
        level=3,
        hp=20,
        max_hp=20,
        base_attack=5,
        gold=10,
        position={"X": 5, "Y": 5},
        biome="Wildlands",
        tile_description="dense forest",
        is_safe_zone=False,
    )
