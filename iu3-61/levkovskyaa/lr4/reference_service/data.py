from __future__ import annotations


def get_clients() -> dict[int, dict[str, str]]:
    return {
        1: {"full_name": "Иван Петров"},
        2: {"full_name": "Анна Соколова"},
        3: {"full_name": "Дмитрий Волков"},
        4: {"full_name": "Елена Миронова"},
    }


def get_workouts() -> dict[int, dict[str, object]]:
    return {
        1: {"title": "Йога", "capacity": 10},
        2: {"title": "Силовая тренировка", "capacity": 8},
        3: {"title": "Бокс", "capacity": 12},
        4: {"title": "Пилатес", "capacity": 9},
        5: {"title": "Стретчинг", "capacity": 15},
        6: {"title": "Кардио", "capacity": 14},
    }