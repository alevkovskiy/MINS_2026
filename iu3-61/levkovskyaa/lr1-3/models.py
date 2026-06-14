from __future__ import annotations

from dataclasses import dataclass
from datetime import date, datetime, timedelta
from enum import Enum
from typing import Optional


class MembershipType(Enum):
    MORNING = "Утренний"
    FULL_DAY = "Полный день"
    PREMIUM = "Премиум"


@dataclass(frozen=True)
class Client:
    client_id: int
    full_name: str
    phone: str
    email: str


@dataclass(frozen=True)
class Membership:
    membership_id: int
    client_id: int
    membership_type: MembershipType
    start_date: date
    end_date: date
    visit_limit: Optional[int] = None
    used_visits: int = 0

    def is_active_on(self, target_date: date) -> bool:
        return self.start_date <= target_date <= self.end_date

    def visits_left(self) -> Optional[int]:
        if self.visit_limit is None:
            return None
        return self.visit_limit - self.used_visits

    def is_expiring_soon(self, today: date, days_before: int = 7) -> bool:
        return today <= self.end_date <= today + timedelta(days=days_before)


@dataclass(frozen=True)
class Workout:
    workout_id: int
    title: str
    trainer: str
    workout_datetime: datetime
    capacity: int


@dataclass(frozen=True)
class Visit:
    visit_id: int
    client_id: int
    workout_id: int
    visited_at: datetime