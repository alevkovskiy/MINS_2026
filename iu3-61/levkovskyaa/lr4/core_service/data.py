from __future__ import annotations


def get_memberships() -> dict[int, dict[str, object]]:
    return {
        1: {"visit_limit": 12, "used_visits": 0, "active": True},
        2: {"visit_limit": None, "used_visits": 0, "active": True},
        3: {"visit_limit": 8, "used_visits": 0, "active": True},
        4: {"visit_limit": 10, "used_visits": 0, "active": True},
    }