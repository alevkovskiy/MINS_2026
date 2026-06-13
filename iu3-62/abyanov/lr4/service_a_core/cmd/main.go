package main

import (
	"log"
	"os"
	"time"

	"warehouse_system/internal/client"
	"warehouse_system/internal/handler"
	"warehouse_system/internal/observer"
	"warehouse_system/internal/repository"
	"warehouse_system/internal/service"
)

func main() {
	log.Println("========================================")
	log.Println("Service A (Core Service) запущен")
	log.Println("========================================")

	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatalf("Ошибка создания директории: %v", err)
	}

	var compatClient *client.CompatibilityClient
	var err error

	for i := 0; i < 5; i++ {
		compatClient, err = client.NewCompatibilityClient("localhost:50051")
		if err == nil {
			log.Println("✅ Успешно подключен к Service B (Reference Service)")
			break
		}
		log.Printf("⚠️  Попытка %d/5: не удалось подключиться к Service B: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if compatClient == nil {
		log.Println("⚠️  ВНИМАНИЕ: Service B недоступен, проверка совместимости будет использовать fallback режим")
	} else {
		defer compatClient.Close()
	}

	repo := repository.NewJSONRepository("data/warehouse.json")
	if err := repo.Load(); err != nil {
		log.Printf("Предупреждение при загрузке данных: %v", err)
	}

	validator := service.NewDefaultProductValidator()
	warehouseService := service.NewWarehouseService(repo, compatClient, validator)

	consoleLogger := observer.NewConsoleLogger()
	lowStockAlert := observer.NewLowStockAlert(5)
	expiryTracker := observer.NewExpiryTracker(7)

	warehouseService.AttachObserver(consoleLogger)
	warehouseService.AttachObserver(lowStockAlert)
	warehouseService.AttachObserver(expiryTracker)

	consoleHandler := handler.NewConsoleHandler(warehouseService, repo)

	if err := consoleHandler.Run(); err != nil {
		log.Fatalf("Ошибка выполнения: %v", err)
	}
}
