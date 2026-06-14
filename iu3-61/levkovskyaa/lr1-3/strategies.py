from __future__ import annotations

from abc import ABC, abstractmethod
from typing import List

from models import Workout


class WorkoutSortStrategy(ABC):
    @abstractmethod
    def sort(self, workouts: List[Workout]) -> List[Workout]:
        pass


class SortByDateStrategy(WorkoutSortStrategy):
    def sort(self, workouts: List[Workout]) -> List[Workout]:
        return sorted(workouts, key=lambda workout: workout.workout_datetime)


class SortByTrainerStrategy(WorkoutSortStrategy):
    def sort(self, workouts: List[Workout]) -> List[Workout]:
        return sorted(workouts, key=lambda workout: workout.trainer.lower())


class SortByTitleStrategy(WorkoutSortStrategy):
    def sort(self, workouts: List[Workout]) -> List[Workout]:
        return sorted(workouts, key=lambda workout: workout.title.lower())