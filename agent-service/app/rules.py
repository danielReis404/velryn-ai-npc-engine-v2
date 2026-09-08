from app.schemas.npc import Perception


def is_in_danger(p: Perception) -> bool:
    return bool(p.alerts and p.alerts.adjacent_monster) and (not p.is_safe_zone)
