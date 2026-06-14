from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Dict, Generic, List, TypeVar

from exceptions import DuplicateEntityError, EntityNotFoundError

T = TypeVar("T")


class Repository(ABC, Generic[T]):
    @abstractmethod
    def add(self, entity: T) -> None:
        pass

    @abstractmethod
    def get_by_id(self, entity_id: int) -> T:
        pass

    @abstractmethod
    def get_all(self) -> List[T]:
        pass

    @abstractmethod
    def update(self, entity: T) -> None:
        pass


class InMemoryRepository(Repository[T], Generic[T]):
    def __init__(self, id_field: str) -> None:
        self._storage: Dict[int, T] = {}
        self._id_field = id_field

    def add(self, entity: T) -> None:
        entity_id = getattr(entity, self._id_field)
        if entity_id in self._storage:
            raise DuplicateEntityError(f"Сущность с id={entity_id} уже существует")
        self._storage[entity_id] = entity

    def get_by_id(self, entity_id: int) -> T:
        entity = self._storage.get(entity_id)
        if entity is None:
            raise EntityNotFoundError(f"Сущность с id={entity_id} не найдена")
        return entity

    def get_all(self) -> List[T]:
        return list(self._storage.values())

    def update(self, entity: T) -> None:
        entity_id = getattr(entity, self._id_field)
        if entity_id not in self._storage:
            raise EntityNotFoundError(f"Сущность с id={entity_id} не найдена")
        self._storage[entity_id] = entity