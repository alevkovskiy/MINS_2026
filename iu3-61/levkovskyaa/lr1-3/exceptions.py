class FitnessClubError(Exception):
    """Базовое исключение системы фитнес-клуба."""


class ValidationError(FitnessClubError):
    """Ошибка валидации входных данных."""


class EntityNotFoundError(FitnessClubError):
    """Сущность не найдена."""


class DuplicateEntityError(FitnessClubError):
    """Попытка добавить уже существующую сущность."""


class BusinessRuleError(FitnessClubError):
    """Нарушение бизнес-правил."""