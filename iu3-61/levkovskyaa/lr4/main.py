from __future__ import annotations

import subprocess
import sys
import time
from pathlib import Path


def main() -> None:
    project_root = Path(__file__).resolve().parent

    reference_cmd = [sys.executable, str(project_root / "reference_service" / "server.py")]
    core_cmd = [sys.executable, str(project_root / "core_service" / "server.py")]

    print("Запуск Reference Service...")
    reference_process = subprocess.Popen(reference_cmd, cwd=project_root)

    time.sleep(1)

    print("Запуск Core Service...")
    core_process = subprocess.Popen(core_cmd, cwd=project_root)

    print("\nОба сервиса запущены.")
    print("Reference Service -> port 50052")
    print("Core Service      -> port 50051")
    print("Для остановки нажми Ctrl+C\n")

    try:
        while True:
            if reference_process.poll() is not None:
                print("Reference Service завершился")
                break
            if core_process.poll() is not None:
                print("Core Service завершился")
                break
            time.sleep(1)
    except KeyboardInterrupt:
        print("\nОстановка сервисов...")

    for process in (core_process, reference_process):
        if process.poll() is None:
            process.terminate()

    for process in (core_process, reference_process):
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()

    print("Сервисы остановлены.")


if __name__ == "__main__":
    main()