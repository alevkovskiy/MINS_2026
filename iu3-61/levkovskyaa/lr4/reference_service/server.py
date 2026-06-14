from __future__ import annotations

import logging
import os
import sys
from concurrent import futures

import grpc

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(CURRENT_DIR)

if PROJECT_ROOT not in sys.path:
    sys.path.insert(0, PROJECT_ROOT)

from generated import reference_service_pb2
from generated import reference_service_pb2_grpc
from reference_service.data import get_clients, get_workouts


def get_trace_id(context: grpc.ServicerContext) -> str:
    metadata = dict(context.invocation_metadata())
    return metadata.get("trace-id", "no-trace-id")


class ReferenceService(reference_service_pb2_grpc.ReferenceServiceServicer):
    def __init__(self) -> None:
        self.clients = get_clients()
        self.workouts = get_workouts()

    def ValidateVisit(self, request, context):
        trace_id = get_trace_id(context)
        client = self.clients.get(request.client_id)
        workout = self.workouts.get(request.workout_id)

        logging.info(
            "[ReferenceService] trace_id=%s ValidateVisit client_id=%s workout_id=%s",
            trace_id,
            request.client_id,
            request.workout_id,
        )

        client_exists = client is not None
        workout_exists = workout is not None

        if not client_exists and not workout_exists:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details("Клиент и тренировка не найдены")
            return reference_service_pb2.ValidateVisitResponse()

        if not client_exists:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details("Клиент не найден")
            return reference_service_pb2.ValidateVisitResponse()

        if not workout_exists:
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details("Тренировка не найдена")
            return reference_service_pb2.ValidateVisitResponse()

        response = reference_service_pb2.ValidateVisitResponse()

        response.ok = True
        response.message = "Валидация прошла успешно"
        response.client_exists = True
        response.workout_exists = True
        response.client_name = str(client["full_name"])
        response.workout_title = str(workout["title"])
        response.workout_capacity = int(workout["capacity"])

        return response


def serve() -> None:
    logging.basicConfig(
        level=logging.INFO,
        format="%(asctime)s %(levelname)s %(message)s",
    )

    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    reference_service_pb2_grpc.add_ReferenceServiceServicer_to_server(ReferenceService(), server)
    server.add_insecure_port("[::]:50052")

    logging.info("ReferenceService started on port 50052")
    server.start()
    server.wait_for_termination()


if __name__ == "__main__":
    serve()