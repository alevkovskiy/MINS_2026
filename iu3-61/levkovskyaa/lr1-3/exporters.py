from __future__ import annotations

import json
from abc import ABC, abstractmethod
from typing import Any, Dict


class ReportExporter(ABC):
    @abstractmethod
    def export(self, report_data: Dict[str, Any]) -> str:
        pass


class TextReportExporter(ReportExporter):
    def export(self, report_data: Dict[str, Any]) -> str:
        return (
            "--- Сводный отчет фитнес-клуба ---\n"
            f"Клиентов: {report_data['clients_count']}\n"
            f"Абонементов: {report_data['memberships_count']}\n"
            f"Тренировок в расписании: {report_data['workouts_count']}\n"
            f"Зафиксированных посещений: {report_data['visits_count']}"
        )


class JsonReportExporter(ReportExporter):
    def export(self, report_data: Dict[str, Any]) -> str:
        return json.dumps(report_data, ensure_ascii=False, indent=4)


class ReportExporterFactory:
    @staticmethod
    def create(format_name: str) -> ReportExporter:
        format_name = format_name.lower()
        if format_name == "txt" or "":
            return TextReportExporter()
        if format_name == "json":
            return JsonReportExporter()
        raise ValueError(f"Неизвестный формат отчета: {format_name}")