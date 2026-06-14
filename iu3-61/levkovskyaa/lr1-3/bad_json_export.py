from __future__ import annotations

import json
from datetime import datetime


class BadClientJsonPrinter:
    """
    Антипаттерн: Hardcoded Dependencies.

    Почему это плохой код:
    1. Класс сам зависит от конкретной реализации client_service.
    2. Формат JSON собран вручную через словари прямо здесь.
    3. Логика вывода захардкожена и не интегрирована в основной модуль отчетов.
    4. Имя файла и часть поведения зашиты прямо в коде.
    """

    def __init__(self, app) -> None:
        # ПЛОХО:
        # вместо передачи абстракции или нужного сервиса напрямую
        # мы принимаем весь app и лезем внутрь его полей
        self._app = app

    def print_clients_json(self) -> str:
        # жесткая зависимость от внутреннего устройства FitnessClubApp
        clients = self._app.client_service.get_all_clients()

        data = {
            "exported_at": datetime.now().isoformat(),
            "clients_count": len(clients),
            "clients": [],
        }

        for client in clients:
            # ручная сборка структуры прямо тут,
            # в обход существующей системы отчетов/экспортеров
            data["clients"].append(
                {
                    "id": client.client_id,
                    "name": client.full_name,
                    "phone": client.phone,
                    "email": client.email,
                }
            )

        result = json.dumps(data, ensure_ascii=False, indent=4)

        # захардкоженное имя файла
        with open("clients_dump.json", "w", encoding="utf-8") as file:
            file.write(result)

        return result 