package main

import (
	"log"
	"os"
	"warehouse_system/internal/handler"
	"warehouse_system/internal/observer"
	"warehouse_system/internal/repository"
	"warehouse_system/internal/service"
)

func main() {
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Ошибка создания директории: %v", err)
	}

	repo := repository.NewJSONRepository("data/warehouse.json")

	if err := repo.Load(); err != nil {
		log.Printf("Предупреждение при загрузке: %v", err)
	}

	checker := service.NewSimpleCompatibilityChecker()

	validator := service.NewDefaultProductValidator()

	warehouseService := service.NewWarehouseService(repo, checker, validator)

	consoleLogger := observer.NewConsoleLogger()
	lowStockAlert := observer.NewLowStockAlert(5)
	expiryTracker := observer.NewExpiryTracker(7)

	warehouseService.AttachObserver(consoleLogger)
	warehouseService.AttachObserver(lowStockAlert)
	warehouseService.AttachObserver(expiryTracker)

	consoleHandler := handler.NewConsoleHandler(warehouseService, repo, checker)

	if err := consoleHandler.Run(); err != nil {
		log.Fatalf("Ошибка выполнения: %v", err)
	}
}
