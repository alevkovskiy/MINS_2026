from __future__ import annotations

import os
import sys
import uuid

import grpc

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(CURRENT_DIR)

if PROJECT_ROOT not in sys.path:
    sys.path.insert(0, PROJECT_ROOT)

from generated import core_service_pb2
from generated import core_service_pb2_grpc


def print_grpc_error(exc: grpc.RpcError) -> None:
    print("ok: False")
    print(f"error_code: {exc.code().name}")
    print(f"message: {exc.details()}")


def main() -> None:
    trace_id = str(uuid.uuid4())

    with grpc.insecure_channel("localhost:50051") as channel:
        stub = core_service_pb2_grpc.CoreServiceStub(channel)

        print("=== RegisterVisit ===")
        try:
            response = stub.RegisterVisit(
                core_service_pb2.RegisterVisitRequest(client_id=1, workout_id=1),
                metadata=(("trace-id", trace_id),),
            )
            print("ok:", response.ok)
            print("message:", response.message)
            print("trace_id:", response.trace_id)
        except grpc.RpcError as exc:
            print_grpc_error(exc)

        print("\n=== GetAllVisits ===")
        try:
            visits_response = stub.GetAllVisits(
                core_service_pb2.GetAllVisitsRequest(),
                metadata=(("trace-id", trace_id),),
            )

            if not visits_response.visits:
                print("Список посещений пуст")
            else:
                for visit in visits_response.visits:
                    print(
                        f"visit_id={visit.visit_id}, "
                        f"client_id={visit.client_id}, "
                        f"workout_id={visit.workout_id}, "
                        f"visited_at={visit.visited_at}"
                    )
        except grpc.RpcError as exc:
            print_grpc_error(exc)


if __name__ == "__main__":
    main()