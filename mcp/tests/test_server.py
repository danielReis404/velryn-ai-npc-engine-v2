from velryn_mcp.server import (
    attack,
    buy,
    defend,
    flee,
    move,
    sell,
    set_goal,
    talk,
    trade,
    use_item,
    wait,
)


def test_move_sets_target_coordinates():
    assert move(3, 4) == {"action": "MOVE", "target_coordinates": [3, 4]}


def test_set_goal_sets_goal_coordinates_not_target_coordinates():

    result = set_goal(10, 2)
    assert result == {"action": "MOVE", "goal_coordinates": [10, 2]}
    assert "target_coordinates" not in result


def test_talk_without_trust_shift_omits_the_field():
    result = talk("Cedric", "Greetings.")
    assert result == {"action": "TALK", "target": "Cedric", "dialogue": "Greetings."}
    assert "trust_shift" not in result


def test_talk_with_trust_shift_included_when_nonzero():
    result = talk("Cedric", "You lied to me.", trust_shift=-10)
    assert result["trust_shift"] == -10


def test_talk_clamps_trust_shift_to_valid_range():
    assert talk("Cedric", "hi", trust_shift=999)["trust_shift"] == 15
    assert talk("Cedric", "hi", trust_shift=-999)["trust_shift"] == -15


def test_attack_without_skill():
    result = attack("Shadow Wolf")
    assert result == {"action": "ATTACK", "target": "Shadow Wolf"}
    assert "skill" not in result


def test_attack_with_skill():
    result = attack("Shadow Wolf", skill="Silent Execution")
    assert result == {
        "action": "ATTACK",
        "target": "Shadow Wolf",
        "skill": "Silent Execution",
    }


def test_defend_takes_no_arguments_and_returns_bare_action():
    assert defend() == {"action": "DEFEND"}


def test_flee_takes_no_arguments_and_returns_bare_action():
    assert flee() == {"action": "FLEE"}


def test_wait_takes_no_arguments_and_returns_bare_action():
    assert wait() == {"action": "WAIT"}


def test_buy_includes_merchant_and_item():
    assert buy("Cedric", "Iron Sword") == {
        "action": "BUY",
        "target": "Cedric",
        "item": "Iron Sword",
    }


def test_sell_includes_merchant_and_item():
    assert sell("Cedric", "Rusty Dagger") == {
        "action": "SELL",
        "target": "Cedric",
        "item": "Rusty Dagger",
    }


def test_trade_with_item_and_gold():
    result = trade("Rowan", item="Health Potion", gold=20)
    assert result == {
        "action": "TRADE",
        "target": "Rowan",
        "item": "Health Potion",
        "gold": 20,
    }


def test_trade_gold_only_payment_omits_item():
    result = trade("Rowan", gold=15)
    assert result == {"action": "TRADE", "target": "Rowan", "gold": 15}
    assert "item" not in result


def test_trade_with_neither_item_nor_gold_still_names_the_target():
    result = trade("Rowan")
    assert result == {"action": "TRADE", "target": "Rowan"}


def test_use_item_consumes_the_named_item():
    assert use_item("Health Potion") == {"action": "USE", "item": "Health Potion"}


def test_every_tool_returns_an_action_field_matching_models_actionverb():

    known_verbs = {
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
    }
    results = [
        move(0, 0),
        set_goal(0, 0),
        talk("x", "y"),
        attack("x"),
        defend(),
        flee(),
        buy("x", "y"),
        sell("x", "y"),
        trade("x"),
        use_item("x"),
        wait(),
    ]
    for result in results:
        assert result["action"] in known_verbs
