from __future__ import annotations

import logging
import os
import sys
import uuid
from concurrent import futures
from datetime import datetime

import grpc

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(CURRENT_DIR)

if PROJECT_ROOT not in sys.path:
    sys.path.insert(0, PROJECT_ROOT)

from core_service.data import get_memberships
from generated import core_service_pb2
from generated import core_service_pb2_grpc
from generated import reference_service_pb2
from generated import reference_service_pb2_grpc


class CoreService(core_service_pb2_grpc.CoreServiceServicer):
    def __init__(self) -> None:
        self.visits: list[dict] = []
        self.next_visit_id = 1
        self.memberships = get_memberships()

    def _extract_trace_id(self, context: grpc.ServicerContext) -> str:
        metadata = dict(context.invocation_metadata())
        return metadata.get("trace-id", str(uuid.uuid4()))

    def _validate_membership(self, client_id: int) -> tuple[bool, str]:
        membership = self.memberships.get(client_id)
        if membership is None:
            return False, "У клиента нет абонемента"

        if not membership["active"]:
            return False, "Абонемент клиента неактивен"

        limit = membership["visit_limit"]
        used = membership["used_visits"]

        if limit is not None and used >= limit:
            return False, "Лимит посещений по абонементу исчерпан"

        return True, "OK"

    def _client_already_registered(self, client_id: int, workout_id: int) -> bool:
        return any(
            visit["client_id"] == client_id and visit["workout_id"] == workout_id
            for visit in self.visits
        )

    def _count_visits_for_workout(self, workout_id: int) -> int:
        return sum(1 for visit in self.visits if visit["workout_id"] == workout_id)

    def _increment_membership_usage(self, client_id: int) -> None:
        membership = self.memberships[client_id]
        if membership["visit_limit"] is not None:
            membership["used_visits"] += 1

    def RegisterVisit(self, request, context):
        trace_id = self._extract_trace_id(context)

        logging.info(
            "[CoreService] trace_id=%s RegisterVisit client_id=%s workout_id=%s",
            trace_id,
            request.client_id,
            request.workout_id,
        )

        membership_ok, membership_message = self._validate_membership(request.client_id)
        if not membership_ok:
            logging.warning(
                "[CoreService] trace_id=%s membership validation failed: %s",
                trace_id,
                membership_message,
            )

            if membership_message == "Лимит посещений по абонементу исчерпан":
                context.set_code(grpc.StatusCode.RESOURCE_EXHAUSTED)
            else:
                context.set_code(grpc.StatusCode.FAILED_PRECONDITION)

            context.set_details(membership_message)
            return core_service_pb2.RegisterVisitResponse(trace_id=trace_id)

        try:
            with grpc.insecure_channel("localhost:50052") as channel:
                stub = reference_service_pb2_grpc.ReferenceServiceStub(channel)

                reference_response = stub.ValidateVisit(
                    reference_service_pb2.ValidateVisitRequest(
                        client_id=request.client_id,
                        workout_id=request.workout_id,
                    ),
                    metadata=(("trace-id", trace_id),),
                    timeout=2.0,
                )
        except grpc.RpcError as exc:
            logging.error(
                "[CoreService] trace_id=%s ReferenceService unavailable: %s",
                trace_id,
                exc,
            )
            context.set_code(grpc.StatusCode.UNAVAILABLE)
            context.set_details("Справочный сервис временно недоступен. Попробуйте позже.")
            return core_service_pb2.RegisterVisitResponse(trace_id=trace_id)

        if self._client_already_registered(request.client_id, request.workout_id):
            context.set_code(grpc.StatusCode.FAILED_PRECONDITION)
            context.set_details("Клиент уже отмечен на этой тренировке")
            return core_service_pb2.RegisterVisitResponse(trace_id=trace_id)

        current_count = self._count_visits_for_workout(request.workout_id)
        if current_count >= reference_response.workout_capacity:
            context.set_code(grpc.StatusCode.RESOURCE_EXHAUSTED)
            context.set_details("На тренировку больше нет свободных мест")
            return core_service_pb2.RegisterVisitResponse(trace_id=trace_id)

        visit = {
            "visit_id": self.next_visit_id,
            "client_id": request.client_id,
            "workout_id": request.workout_id,
            "visited_at": datetime.now().isoformat(timespec="seconds"),
        }
        self.visits.append(visit)
        self.next_visit_id += 1

        self._increment_membership_usage(request.client_id)

        logging.info(
            "[CoreService] trace_id=%s visit registered successfully visit_id=%s",
            trace_id,
            visit["visit_id"],
        )

        return core_service_pb2.RegisterVisitResponse(
            ok=True,
            message=(
                f"Посещение зарегистрировано: "
                f"{reference_response.client_name} -> {reference_response.workout_title}"
            ),
            trace_id=trace_id,
        )

    def GetAllVisits(self, request, context):
        trace_id = self._extract_trace_id(context)
        logging.info("[CoreService] trace_id=%s GetAllVisits", trace_id)

        items = [
            core_service_pb2.VisitItem(
                visit_id=visit["visit_id"],
                client_id=visit["client_id"],
                workout_id=visit["workout_id"],
                visited_at=visit["visited_at"],
            )
            for visit in self.visits
        ]
        return core_service_pb2.GetAllVisitsResponse(visits=items)


def serve() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
    )

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    core_service_pb2_grpc.add_CoreServiceServicer_to_server(CoreService(), server)
    server.add_insecure_port("[::]:50051")

    logging.info("CoreService started on port 50051")
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()