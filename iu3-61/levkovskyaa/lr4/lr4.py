from __future__ import annotations

from dataclasses import dataclass
from typing import Iterable, List

try:
    import matplotlib.pyplot as plt
except Exception:  # если matplotlib не установлен
    plt = None


Q15_SCALE = 1 << 15
Q15_MIN = -32768
Q15_MAX = 32767
FS = 48_000

# Квантованные коэффициенты КИХ-фильтра из Filter Designer, формат 1.15
# Получены из коэффициентов:
# -0.0104128, -0.00910359, 0.00518372, ...
COEFFS_Q15: List[int] = [
    -341, -298, 170, 1438,
    3549, 6151, 8560, 10014,
    10014, 8560, 6151, 3549,
    1438, 170, -298, -341,
]


@dataclass
class FixedQ15:
    value: int

    def __post_init__(self) -> None:
        self.value = saturate_q15(self.value)

    @staticmethod
    def from_float(x: float) -> "FixedQ15":
        return FixedQ15(float_to_q15(x))

    def to_float(self) -> float:
        return q15_to_float(self.value)

    def __repr__(self) -> str:
        return f"{self.to_float():.8f}"


def saturate_q15(x: int) -> int:
    """
    Ограничение значения диапазоном формата 1.15.
    """
    if x > Q15_MAX:
        return Q15_MAX
    if x < Q15_MIN:
        return Q15_MIN
    return x


def float_to_q15(x: float) -> int:
    """
    Перевод вещественного числа из диапазона [-1, 1) в формат 1.15.
    Число +1.0 в формате 1.15 непредставимо, поэтому заменяется на максимум.
    """
    if x >= 1.0:
        x = Q15_MAX / Q15_SCALE
    if x < -1.0:
        x = -1.0

    return saturate_q15(int(round(x * Q15_SCALE)))


def q15_to_float(x: int) -> float:
    """
    Перевод числа из формата 1.15 в вещественный вид.
    """
    return x / Q15_SCALE


def q15_mul(a: int, b: int) -> int:
    """
    Умножение двух чисел Q15.
    Промежуточный результат имеет формат Q30.
    """
    product = a * b
    product += 1 << 14  # округление перед сдвигом
    return saturate_q15(product >> 15)


class FIRFilterQ15:
    """
    КИХ-фильтр с фиксированной запятой в формате 1.15.
    """

    def __init__(self, coeffs_q15: Iterable[int]) -> None:
        self.coeffs = [saturate_q15(c) for c in coeffs_q15]
        self.n_taps = len(self.coeffs)
        self.state = [0] * self.n_taps

    def reset(self) -> None:
        self.state = [0] * self.n_taps

    def process_sample(self, x_q15: int) -> int:
        """
        Обработка одного отсчёта входного сигнала.
        """
        self.state.pop(0)
        self.state.append(saturate_q15(x_q15))

        # Аккумулятор расширенной разрядности.
        # При умножении Q15 * Q15 получается Q30.
        acc = 0

        for k in range(self.n_taps):
            acc += self.coeffs[k] * self.state[self.n_taps - 1 - k]

        # Округление и перевод из Q30 обратно в Q15
        acc += 1 << 14
        y_q15 = acc >> 15

        return saturate_q15(y_q15)

    def process(self, signal_q15: Iterable[int]) -> List[int]:
        return [self.process_sample(x) for x in signal_q15]


def print_signal(title: str, signal_q15: List[int]) -> None:
    print(f"\n{title}")
    print("Q15:")
    print(signal_q15)
    print("float:")
    print([round(q15_to_float(x), 8) for x in signal_q15])


def save_plot(y_q15: List[int], title: str, filename: str) -> None:
    """
    Сохранение графика сигнала.
    """
    if plt is None:
        print(f"matplotlib не установлен, график {filename} не сохранён")
        return

    x = list(range(len(y_q15)))
    y = [q15_to_float(v) for v in y_q15]

    plt.figure(figsize=(10, 4.5))
    plt.stem(x, y)
    plt.title(title)
    plt.xlabel("Номер отсчёта n")
    plt.ylabel("Амплитуда")
    plt.grid(True)
    plt.tight_layout()
    plt.savefig(filename, dpi=150)
    plt.close()


def main() -> None:
    coeffs_float = [q15_to_float(c) for c in COEFFS_Q15]

    print_signal("Коэффициенты фильтра", COEFFS_Q15)

    fir = FIRFilterQ15(COEFFS_Q15)

    # ---------------------------------------------------------
    # 1. Реакция на единичный импульс
    # ---------------------------------------------------------
    # Число +1 в Q15 непредставимо, поэтому используем Q15_MAX.
    impulse = [Q15_MAX] + [0] * 63

    impulse_response = fir.process(impulse)

    print_signal("Импульсная характеристика", impulse_response)

    save_plot(
        impulse_response,
        "Импульсная характеристика КИХ-фильтра в формате 1.15",
        "impulse_response.png",
    )

    # ---------------------------------------------------------
    # 2. Реакция на ступенчатый сигнал
    # ---------------------------------------------------------
    fir.reset()

    # Если использовать 0.95, из-за суммы коэффициентов больше 1
    # выход может насыщаться около максимума Q15.
    step_amplitude = 0.95

    step = [float_to_q15(step_amplitude)] * 64

    step_response = fir.process(step)

    print_signal("Переходная характеристика", step_response)

    save_plot(
        step_response,
        "Переходная характеристика КИХ-фильтра в формате 1.15",
        "step_response.png",
    )

    print("\nСправочно:")
    print(f"Сумма коэффициентов = {sum(coeffs_float):.8f}")
    print(f"LSB для формата 1.15 = {2 ** -15:.10f}")
    print(f"Частота дискретизации = {FS} Гц")


if __name__ == "__main__":
    main()