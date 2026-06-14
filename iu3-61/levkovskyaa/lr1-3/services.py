from __future__ import annotations

from datetime import date, datetime, timedelta
from typing import Dict, List, Optional

from exceptions import BusinessRuleError, ValidationError
from exporters import ReportExporterFactory
from models import Client, Membership, MembershipType, Visit, Workout
from observer import EventManager
from repositories import Repository
from strategies import SortByDateStrategy, WorkoutSortStrategy


class IdGenerator:
    def __init__(self, start: int = 1) -> None:
        self._current = start

    def next_id(self) -> int:
        current = self._current
        self._current += 1
        return current


class ClientService:
    def __init__(self, client_repo: Repository[Client], client_id_generator: IdGenerator) -> None:
        self._client_repo = client_repo
        self._client_id_generator = client_id_generator

    def register_client(self, full_name: str, phone: str, email: str) -> Client:
        if not full_name.strip():
            raise ValidationError("ФИО клиента не может быть пустым")
        if "@" not in email:
            raise ValidationError("Некорректный email")
        if len(phone.strip()) < 5:
            raise ValidationError("Некорректный номер телефона")

        client = Client(
            client_id=self._client_id_generator.next_id(),
            full_name=full_name.strip(),
            phone=phone.strip(),
            email=email.strip(),
        )
        self._client_repo.add(client)
        return client

    def get_all_clients(self) -> List[Client]:
        return self._client_repo.get_all()

    def get_client(self, client_id: int) -> Client:
        return self._client_repo.get_by_id(client_id)


class MembershipService:
    def __init__(
        self,
        membership_repo: Repository[Membership],
        client_repo: Repository[Client],
        membership_id_generator: IdGenerator,
        event_manager: EventManager,
    ) -> None:
        self._membership_repo = membership_repo
        self._client_repo = client_repo
        self._membership_id_generator = membership_id_generator
        self._event_manager = event_manager

    def create_membership(
        self,
        client_id: int,
        membership_type: MembershipType,
        start_date: date,
        duration_days: int,
        visit_limit: Optional[int] = None,
    ) -> Membership:
        client = self._client_repo.get_by_id(client_id)

        if duration_days <= 0:
            raise ValidationError("Длительность абонемента должна быть положительной")
        if visit_limit is not None and visit_limit <= 0:
            raise ValidationError("Лимит посещений должен быть положительным")

        membership = Membership(
            membership_id=self._membership_id_generator.next_id(),
            client_id=client_id,
            membership_type=membership_type,
            start_date=start_date,
            end_date=start_date + timedelta(days=duration_days),
            visit_limit=visit_limit,
        )
        self._membership_repo.add(membership)

        self._event_manager.notify(
            "membership_created",
            {"client": client, "membership": membership},
        )
        return membership

    def find_active_membership(self, client_id: int, target_date: date) -> Membership:
        memberships = [m for m in self._membership_repo.get_all() if m.client_id == client_id]
        for membership in memberships:
            if membership.is_active_on(target_date):
                return membership
        raise BusinessRuleError("У клиента нет активного абонемента на указанную дату")

    def register_visit_usage(self, membership_id: int) -> Membership:
        membership = self._membership_repo.get_by_id(membership_id)
        if membership.visit_limit is not None and membership.used_visits >= membership.visit_limit:
            raise BusinessRuleError("Лимит посещений по абонементу исчерпан")

        updated = Membership(
            membership_id=membership.membership_id,
            client_id=membership.client_id,
            membership_type=membership.membership_type,
            start_date=membership.start_date,
            end_date=membership.end_date,
            visit_limit=membership.visit_limit,
            used_visits=membership.used_visits + 1,
        )
        self._membership_repo.update(updated)
        return updated

    def get_memberships(self) -> List[Membership]:
        return self._membership_repo.get_all()

    def get_expiring_memberships(self, today: date, days_before: int = 7) -> List[Membership]:
        return [
            membership
            for membership in self._membership_repo.get_all()
            if membership.is_expiring_soon(today, days_before)
        ]


class WorkoutService:
    def __init__(
        self,
        workout_repo: Repository[Workout],
        workout_id_generator: IdGenerator,
        sort_strategy: Optional[WorkoutSortStrategy] = None,
    ) -> None:
        self._workout_repo = workout_repo
        self._workout_id_generator = workout_id_generator
        self._sort_strategy = sort_strategy or SortByDateStrategy()

    def set_sort_strategy(self, strategy: WorkoutSortStrategy) -> None:
        self._sort_strategy = strategy

    def schedule_workout(
        self,
        title: str,
        trainer: str,
        workout_datetime: datetime,
        capacity: int,
    ) -> Workout:
        if not title.strip():
            raise ValidationError("Название тренировки не может быть пустым")
        if not trainer.strip():
            raise ValidationError("Имя тренера не может быть пустым")
        if capacity <= 0:
            raise ValidationError("Вместимость тренировки должна быть положительной")

        workout = Workout(
            workout_id=self._workout_id_generator.next_id(),
            title=title.strip(),
            trainer=trainer.strip(),
            workout_datetime=workout_datetime,
            capacity=capacity,
        )
        self._workout_repo.add(workout)
        return workout

    def get_workout(self, workout_id: int) -> Workout:
        return self._workout_repo.get_by_id(workout_id)

    def get_all_workouts(self) -> List[Workout]:
        return self._sort_strategy.sort(self._workout_repo.get_all())


class VisitService:
    def __init__(
        self,
        visit_repo: Repository[Visit],
        workout_repo: Repository[Workout],
        membership_service: MembershipService,
        client_service: ClientService,
        visit_id_generator: IdGenerator,
        event_manager: EventManager,
    ) -> None:
        self._visit_repo = visit_repo
        self._workout_repo = workout_repo
        self._membership_service = membership_service
        self._client_service = client_service
        self._visit_id_generator = visit_id_generator
        self._event_manager = event_manager

    def register_visit(self, client_id: int, workout_id: int, visited_at: datetime) -> Visit:
        workout = self._workout_repo.get_by_id(workout_id)
        client = self._client_service.get_client(client_id)

        existing_visits = [visit for visit in self._visit_repo.get_all() if visit.workout_id == workout_id]
        if len(existing_visits) >= workout.capacity:
            raise BusinessRuleError("На тренировку больше нет свободных мест")

        duplicated = [
            visit
            for visit in self._visit_repo.get_all()
            if visit.client_id == client_id and visit.workout_id == workout_id
        ]
        if duplicated:
            raise BusinessRuleError("Клиент уже отмечен на этой тренировке")

        membership = self._membership_service.find_active_membership(client_id, visited_at.date())
        self._membership_service.register_visit_usage(membership.membership_id)

        visit = Visit(
            visit_id=self._visit_id_generator.next_id(),
            client_id=client_id,
            workout_id=workout_id,
            visited_at=visited_at,
        )
        self._visit_repo.add(visit)

        self._event_manager.notify(
            "visit_registered",
            {"client": client, "workout": workout, "visit": visit},
        )
        return visit

    def get_all_visits(self) -> List[Visit]:
        return sorted(self._visit_repo.get_all(), key=lambda item: item.visited_at)


class ReminderService:
    def __init__(
        self,
        membership_service: MembershipService,
        client_service: ClientService,
        event_manager: EventManager,
    ) -> None:
        self._membership_service = membership_service
        self._client_service = client_service
        self._event_manager = event_manager

    def send_renewal_reminders(self, today: date, days_before: int = 7) -> None:
        memberships = self._membership_service.get_expiring_memberships(today, days_before)
        for membership in memberships:
            client = self._client_service.get_client(membership.client_id)
            self._event_manager.notify(
                "membership_expiring",
                {"client": client, "membership": membership},
            )


class ReportService:
    def __init__(
        self,
        client_service: ClientService,
        membership_service: MembershipService,
        workout_service: WorkoutService,
        visit_service: VisitService,
    ) -> None:
        self._client_service = client_service
        self._membership_service = membership_service
        self._workout_service = workout_service
        self._visit_service = visit_service

    def build_summary_data(self) -> Dict[str, int]:
        return {
            "clients_count": len(self._client_service.get_all_clients()),
            "memberships_count": len(self._membership_service.get_memberships()),
            "workouts_count": len(self._workout_service.get_all_workouts()),
            "visits_count": len(self._visit_service.get_all_visits()),
        }

    def export_summary(self, format_name: str) -> str:
        exporter = ReportExporterFactory.create(format_name)
        data = self.build_summary_data()
        return exporter.export(data)