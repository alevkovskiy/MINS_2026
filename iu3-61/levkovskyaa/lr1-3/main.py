from __future__ import annotations

from datetime import date, datetime, timedelta

from exceptions import FitnessClubError, ValidationError
from models import Client, Membership, MembershipType, Visit, Workout
from observer import AuditObserver, EventManager, NotificationObserver
from repositories import InMemoryRepository
from services import (
    ClientService,
    IdGenerator,
    MembershipService,
    ReminderService,
    ReportService,
    VisitService,
    WorkoutService,
)
from bad_json_export import BadClientJsonPrinter
from strategies import SortByDateStrategy, SortByTitleStrategy, SortByTrainerStrategy


class FitnessClubApp:
    def __init__(self) -> None:
        self.event_manager = EventManager()

        notification_observer = NotificationObserver()
        audit_observer = AuditObserver()

        self.event_manager.subscribe("membership_created", notification_observer)
        self.event_manager.subscribe("membership_created", audit_observer)

        self.event_manager.subscribe("visit_registered", notification_observer)
        self.event_manager.subscribe("visit_registered", audit_observer)

        self.event_manager.subscribe("membership_expiring", notification_observer)
        self.event_manager.subscribe("membership_expiring", audit_observer)

        client_repo = InMemoryRepository[Client]("client_id")
        membership_repo = InMemoryRepository[Membership]("membership_id")
        workout_repo = InMemoryRepository[Workout]("workout_id")
        visit_repo = InMemoryRepository[Visit]("visit_id")

        self.bad_json_printer = BadClientJsonPrinter(self)

        self.client_service = ClientService(client_repo, IdGenerator())
        self.membership_service = MembershipService(
            membership_repo,
            client_repo,
            IdGenerator(),
            self.event_manager,
        )
        self.workout_service = WorkoutService(
            workout_repo,
            IdGenerator(),
            SortByDateStrategy(),
        )
        self.visit_service = VisitService(
            visit_repo,
            workout_repo,
            self.membership_service,
            self.client_service,
            IdGenerator(),
            self.event_manager,
        )
        self.reminder_service = ReminderService(
            self.membership_service,
            self.client_service,
            self.event_manager,
        )
        self.report_service = ReportService(
            self.client_service,
            self.membership_service,
            self.workout_service,
            self.visit_service,
        )

    def seed_demo_data(self) -> None:
        client_1 = self.client_service.register_client("Иван Петров", "+79990001122", "ivan@mail.ru")
        client_2 = self.client_service.register_client("Анна Соколова", "+79990003344", "anna@mail.ru")
        client_3 = self.client_service.register_client("Дмитрий Волков", "+79990005566", "dmitry@mail.ru")
        client_4 = self.client_service.register_client("Елена Миронова", "+79990007788", "elena@mail.ru")

        self.membership_service.create_membership(
            client_1.client_id,
            MembershipType.FULL_DAY,
            date.today() - timedelta(days=10),
            duration_days=30,
            visit_limit=12,
        )
        self.membership_service.create_membership(
            client_2.client_id,
            MembershipType.PREMIUM,
            date.today() - timedelta(days=25),
            duration_days=30,
            visit_limit=None,
        )
        self.membership_service.create_membership(
            client_3.client_id,
            MembershipType.MORNING,
            date.today() - timedelta(days=5),
            duration_days=20,
            visit_limit=8,
        )
        self.membership_service.create_membership(
            client_4.client_id,
            MembershipType.FULL_DAY,
            date.today() - timedelta(days=2),
            duration_days=15,
            visit_limit=10,
        )

        now = datetime.now()

        workout_1 = self.workout_service.schedule_workout(
            "Йога",
            "Мария Орлова",
            now + timedelta(hours=2),
            capacity=10,
        )
        workout_2 = self.workout_service.schedule_workout(
            "Силовая тренировка",
            "Олег Миронов",
            now + timedelta(days=1, hours=1),
            capacity=8,
        )
        workout_3 = self.workout_service.schedule_workout(
            "Бокс",
            "Алексей Смирнов",
            now + timedelta(hours=5),
            capacity=12,
        )
        workout_4 = self.workout_service.schedule_workout(
            "Пилатес",
            "Виктор Лебедев",
            now + timedelta(days=2),
            capacity=9,
        )
        workout_5 = self.workout_service.schedule_workout(
            "Стретчинг",
            "Мария Орлова",
            now + timedelta(days=1, hours=3),
            capacity=15,
        )
        workout_6 = self.workout_service.schedule_workout(
            "Кардио",
            "Екатерина Белова",
            now + timedelta(hours=8),
            capacity=14,
        )
        workout_7 = self.workout_service.schedule_workout(
            "Функциональный тренинг",
            "Олег Миронов",
            now + timedelta(days=3),
            capacity=10,
        )
        workout_8 = self.workout_service.schedule_workout(
            "Zumba",
            "Анна Крылова",
            now + timedelta(days=2, hours=4),
            capacity=16,
        )
        workout_9 = self.workout_service.schedule_workout(
            "Кроссфит",
            "Денис Павлов",
            now + timedelta(hours=12),
            capacity=10,
        )
        workout_10 = self.workout_service.schedule_workout(
            "Танцы",
            "Анна Крылова",
            now + timedelta(days=4),
            capacity=20,
        )
        workout_11 = self.workout_service.schedule_workout(
            "TRX",
            "Сергей Романов",
            now + timedelta(days=1, hours=6),
            capacity=8,
        )
        workout_12 = self.workout_service.schedule_workout(
            "Аэробика",
            "Екатерина Белова",
            now + timedelta(days=3, hours=2),
            capacity=18,
        )

        self.visit_service.register_visit(client_1.client_id, workout_1.workout_id, now)
        self.visit_service.register_visit(client_2.client_id, workout_3.workout_id, now + timedelta(minutes=10))
        self.visit_service.register_visit(client_3.client_id, workout_6.workout_id, now + timedelta(minutes=20))
        self.visit_service.register_visit(client_4.client_id, workout_2.workout_id, now + timedelta(minutes=30))
    
    def _print_clients_json_bad(self) -> None:
        result = self.bad_json_printer.print_clients_json()
        print("\n=== Быстрый JSON клиентов (ЛР3) ===")
        print(result)
        print("\nФайл также сохранен как clients_dump.json")

    def run(self) -> None:
        self.seed_demo_data()

        while True:
            print(
                "\n=== Фитнес-клуб (ЛР2) ===\n"
                "1. Показать клиентов\n"
                "2. Показать абонементы\n"
                "3. Показать расписание тренировок\n"
                "4. Показать посещения\n"
                "5. Добавить клиента\n"
                "6. Оформить абонемент\n"
                "7. Запланировать тренировку\n"
                "8. Отметить посещение\n"
                "9. Отправить напоминания о продлении\n"
                "10. Экспорт отчета\n"
                "11. Сменить стратегию сортировки тренировок\n"
                "12. Быстрый JSON-список клиентов\n" 
                "0. Выход"
            )
            choice = input("Выберите действие: ").strip()

            try:
                if choice == "1":
                    self._print_clients()
                elif choice == "2":
                    self._print_memberships()
                elif choice == "3":
                    self._print_workouts()
                elif choice == "4":
                    self._print_visits()
                elif choice == "5":
                    self._create_client_from_input()
                elif choice == "6":
                    self._create_membership_from_input()
                elif choice == "7":
                    self._create_workout_from_input()
                elif choice == "8":
                    self._register_visit_from_input()
                elif choice == "9":
                    self.reminder_service.send_renewal_reminders(date.today())
                elif choice == "10":
                    self._export_report()
                elif choice == "11":
                    self._change_workout_sorting()
                elif choice == "12":
                    self._print_clients_json_bad()
                elif choice == "0":
                    print("Выход из программы")
                    break
                else:
                    print("Неизвестная команда")
            except FitnessClubError as exc:
                print(f"Ошибка: {exc}")
            except ValueError:
                print("Ошибка: введены данные неверного формата")

    def _print_clients(self) -> None:
        for client in self.client_service.get_all_clients():
            print(f"ID={client.client_id} | {client.full_name} | {client.phone} | {client.email}")

    def _print_memberships(self) -> None:
        for membership in self.membership_service.get_memberships():
            visits_left = membership.visits_left()
            visits_text = "безлимит" if visits_left is None else str(visits_left)
            print(
                f"ID={membership.membership_id} | client_id={membership.client_id} | "
                f"{membership.membership_type.value} | {membership.start_date} - {membership.end_date} | "
                f"осталось посещений: {visits_text}"
            )

    def _print_workouts(self) -> None:
        for workout in self.workout_service.get_all_workouts():
            print(
                f"ID={workout.workout_id} | {workout.title} | тренер: {workout.trainer} | "
                f"{workout.workout_datetime.strftime('%Y-%m-%d %H:%M')} | мест: {workout.capacity}"
            )

    def _print_visits(self) -> None:
        for visit in self.visit_service.get_all_visits():
            print(
                f"ID={visit.visit_id} | client_id={visit.client_id} | workout_id={visit.workout_id} | "
                f"{visit.visited_at.strftime('%Y-%m-%d %H:%M')}"
            )

    def _create_client_from_input(self) -> None:
        full_name = input("ФИО: ")
        phone = input("Телефон: ")
        email = input("Email: ")
        client = self.client_service.register_client(full_name, phone, email)
        print(f"Клиент добавлен: ID={client.client_id}")

    def _create_membership_from_input(self) -> None:
        client_id = int(input("ID клиента: "))
        print("Типы абонемента: 1 - Утренний, 2 - Полный день, 3 - Премиум")
        type_choice = input("Выберите тип: ").strip()
        mapping = {"1": MembershipType.MORNING, "2": MembershipType.FULL_DAY, "3": MembershipType.PREMIUM}
        membership_type = mapping.get(type_choice)
        if membership_type is None:
            raise ValidationError("Неверный тип абонемента")

        start_date = datetime.strptime(input("Дата начала (YYYY-MM-DD): "), "%Y-%m-%d").date()
        duration_days = int(input("Срок действия в днях: "))
        visit_limit_text = input("Лимит посещений (Enter для безлимита): ").strip()
        visit_limit = int(visit_limit_text) if visit_limit_text else None

        membership = self.membership_service.create_membership(
            client_id,
            membership_type,
            start_date,
            duration_days,
            visit_limit,
        )
        print(f"Абонемент оформлен: ID={membership.membership_id}")

    def _create_workout_from_input(self) -> None:
        title = input("Название тренировки: ")
        trainer = input("Тренер: ")
        workout_datetime = datetime.strptime(
            input("Дата и время (YYYY-MM-DD HH:MM): "),
            "%Y-%m-%d %H:%M",
        )
        capacity = int(input("Количество мест: "))
        workout = self.workout_service.schedule_workout(title, trainer, workout_datetime, capacity)
        print(f"Тренировка добавлена: ID={workout.workout_id}")

    def _register_visit_from_input(self) -> None:
        client_id = int(input("ID клиента: "))
        workout_id = int(input("ID тренировки: "))
        visit = self.visit_service.register_visit(client_id, workout_id, datetime.now())
        print(f"Посещение зафиксировано: ID={visit.visit_id}")

    def _export_report(self) -> None:
        format_name = input("Формат отчета txt/json (default txt): ").strip().lower()
        report = self.report_service.export_summary(format_name)
        print("\n" + report)

    def _change_workout_sorting(self) -> None:
        print("1 - по дате, 2 - по тренеру, 3 - по названию")
        choice = input("Выберите стратегию сортировки: ").strip()

        if choice == "1":
            self.workout_service.set_sort_strategy(SortByDateStrategy())
        elif choice == "2":
            self.workout_service.set_sort_strategy(SortByTrainerStrategy())
        elif choice == "3":
            self.workout_service.set_sort_strategy(SortByTitleStrategy())
        else:
            raise ValidationError("Неверный вариант стратегии")

        print("Стратегия сортировки обновлена")


if __name__ == "__main__":
    app = FitnessClubApp()
    app.run()