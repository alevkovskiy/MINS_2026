from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Any, Dict, List


class Observer(ABC):
    @abstractmethod
    def update(self, event_name: str, payload: Dict[str, Any]) -> None:
        pass


class EventManager:
    def __init__(self) -> None:
        self._subscribers: Dict[str, List[Observer]] = {}

    def subscribe(self, event_name: str, observer: Observer) -> None:
        self._subscribers.setdefault(event_name, []).append(observer)

    def notify(self, event_name: str, payload: Dict[str, Any]) -> None:
        for observer in self._subscribers.get(event_name, []):
            observer.update(event_name, payload)


class NotificationObserver(Observer):
    def update(self, event_name: str, payload: Dict[str, Any]) -> None:
        if event_name == "membership_created":
            client = payload["client"]
            membership = payload["membership"]
            print(
                f"[УВЕДОМЛЕНИЕ] Клиенту {client.email}: "
                f"оформлен абонемент '{membership.membership_type.value}' "
                f"до {membership.end_date}"
            )

        elif event_name == "visit_registered":
            client = payload["client"]
            workout = payload["workout"]
            print(
                f"[УВЕДОМЛЕНИЕ] Клиенту {client.email}: "
                f"посещение тренировки '{workout.title}' зафиксировано"
            )

        elif event_name == "membership_expiring":
            client = payload["client"]
            membership = payload["membership"]
            print(
                f"[УВЕДОМЛЕНИЕ] Клиенту {client.email}: "
                f"абонемент '{membership.membership_type.value}' истекает {membership.end_date}"
            )


class AuditObserver(Observer):
    def update(self, event_name: str, payload: Dict[str, Any]) -> None:
        print(f"[AUDIT] Событие: {event_name} | Данные: {payload}")