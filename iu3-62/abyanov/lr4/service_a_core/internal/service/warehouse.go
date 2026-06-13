package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"warehouse_system/internal/client"
	"warehouse_system/internal/domain"
	"warehouse_system/internal/observer"
	"warehouse_system/internal/repository"
	customerrors "warehouse_system/pkg/errors"
)

type ProductValidator interface {
	Validate(p domain.Product) error
}

type Statistics struct {
	TotalItems int            `json:"total_items"`
	TotalValue float64        `json:"total_value"`
	Categories map[string]int `json:"categories"`
}

type WarehouseService struct {
	repo         repository.Repository
	compatClient *client.CompatibilityClient
	validator    ProductValidator
	notifier     *observer.WarehouseNotifier
	luckyLottery *LuckyLottery
}

func NewWarehouseService(
	repo repository.Repository,
	compatClient *client.CompatibilityClient,
	validator ProductValidator,
) *WarehouseService {
	return &WarehouseService{
		repo:         repo,
		compatClient: compatClient,
		validator:    validator,
		notifier:     observer.NewWarehouseNotifier(),
		luckyLottery: NewLuckyLottery(),
	}
}

func (s *WarehouseService) AttachObserver(obs observer.Observer) {
	s.notifier.Attach(obs)
}

func (s *WarehouseService) DetachObserver(obs observer.Observer) {
	s.notifier.Detach(obs)
}

func (s *WarehouseService) IsServiceBAvailable() bool {
	if s.compatClient == nil {
		return false
	}
	return s.compatClient.IsServiceBAvailable()
}

func (s *WarehouseService) GetCategoriesFromServiceB(traceID string) []string {
	if s.compatClient == nil {
		return []string{}
	}
	ctx := context.Background()
	categories, err := s.compatClient.GetCategories(ctx, traceID)
	if err != nil {
		log.Printf("[TraceID: %s] Ошибка получения категорий: %v", traceID, err)
		return []string{}
	}
	return categories
}

func (s *WarehouseService) CheckCompatibilityWithServiceB(traceID, category string, existingCategories []string) (bool, string, error) {
	if s.compatClient == nil {
		return true, "Service B недоступен, проверка пропущена", nil
	}
	ctx := context.Background()
	return s.compatClient.CheckCompatibilityWithFallback(ctx, traceID, category, existingCategories)
}

func (s *WarehouseService) AddProduct(p domain.Product, traceID string) error {
	// Валидация товара
	if err := s.validator.Validate(p); err != nil {
		log.Printf("[TraceID: %s] ❌ Ошибка валидации: %v", traceID, err)
		return err
	}

	// Получение существующих категорий на складе
	existing := s.repo.FindAll()
	existingCategories := make([]string, 0, len(existing))
	catMap := make(map[string]bool)
	for _, item := range existing {
		if !catMap[item.GetCategory()] {
			catMap[item.GetCategory()] = true
			existingCategories = append(existingCategories, item.GetCategory())
		}
	}

	log.Printf("[TraceID: %s] Проверка совместимости для категории '%s' с существующими: %v",
		traceID, p.GetCategory(), existingCategories)

	// Проверка совместимости через Service B
	compatible, compatMsg, err := s.CheckCompatibilityWithServiceB(traceID, p.GetCategory(), existingCategories)

	if err != nil {
		// Обработка ошибок gRPC с различными кодами статусов
		log.Printf("[TraceID: %s] ⚠️ Ошибка при вызове Service B: %v", traceID, err)

		// В случае фатальных ошибок (например, PermissionDenied) - блокируем добавление
		errStr := err.Error()
		if containsAny(errStr, []string{"PermissionDenied", "InvalidArgument", "code = 7", "code = 3"}) {
			log.Printf("[TraceID: %s] 🔒 Критическая ошибка Service B, добавление товара заблокировано", traceID)
			return customerrors.NewCompatibilityError(
				fmt.Sprintf("Сервис проверки совместимости вернул критическую ошибку: %v", err),
				p.GetID(), p.GetName())
		}

		// Для остальных ошибок (Timeout, Unavailable, Internal) - используем fallback
		log.Printf("[TraceID: %s] Используем fallback режим из-за ошибки Service B", traceID)
		// В fallback режиме считаем товар совместимым
		compatible = true
		compatMsg = fmt.Sprintf("Сервис проверки недоступен (ошибка: %v), товар добавлен без проверки", err)
	}

	// Если несовместим - возвращаем ошибку
	if !compatible {
		log.Printf("[TraceID: %s] ❌ Нарушение совместимости: %s", traceID, compatMsg)
		return customerrors.NewCompatibilityError(compatMsg, p.GetID(), p.GetName())
	}

	if compatMsg != "" {
		log.Printf("[TraceID: %s] ⚠️ Предупреждение совместимости: %s", traceID, compatMsg)
	}

	// Применяем лотерейную магию (если условия выполнены)
	p = s.luckyLottery.ApplyLotteryMagic(p)

	// Сохраняем товар в репозиторий
	if err := s.repo.Save(p); err != nil {
		log.Printf("[TraceID: %s] ❌ Ошибка сохранения товара: %v", traceID, err)
		return fmt.Errorf("ошибка сохранения товара: %w", err)
	}

	log.Printf("[TraceID: %s] ✅ Товар '%s' (ID: %s) успешно добавлен, количество: %d",
		traceID, p.GetName(), p.GetID(), p.GetQuantity())

	// Уведомляем наблюдателей о добавлении товара
	s.notifier.Notify(observer.Event{
		Type:      observer.ProductAdded,
		Product:   p,
		Quantity:  p.GetQuantity(),
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Товар добавлен. TraceID: %s, совместимость: %t", traceID, compatible),
	})

	// Дополнительная проверка для продуктов питания (срок годности)
	if food, ok := p.(*domain.Food); ok {
		expiryDate, err := time.Parse("2006-01-02", food.ExpiryDate)
		if err == nil {
			daysUntilExpiry := int(time.Until(expiryDate).Hours() / 24)
			if daysUntilExpiry >= 0 && daysUntilExpiry <= 7 {
				log.Printf("[TraceID: %s] ⚠️ Товар '%s' истекает через %d дней", traceID, p.GetName(), daysUntilExpiry)
				s.notifier.Notify(observer.Event{
					Type:      observer.ExpiringSoon,
					Product:   p,
					Timestamp: time.Now(),
					Message:   food.ExpiryDate,
				})
			}
		}
	}

	return nil
}

// containsAny проверяет, содержит ли строка хотя бы одну из подстрок
func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

func (s *WarehouseService) RemoveProduct(id string, qty int) error {
	if qty <= 0 {
		err := customerrors.NewValidationError("количество должно быть положительным", "quantity")
		log.Printf("❌ Ошибка удаления: %v", err)
		return err
	}

	product, err := s.repo.FindByID(id)
	if err != nil {
		log.Printf("❌ Товар с ID '%s' не найден: %v", id, err)
		return err
	}

	if product.GetQuantity() < qty {
		err := customerrors.NewStockError("недостаточное количество", product.GetQuantity(), qty)
		log.Printf("❌ Ошибка удаления: %v (доступно: %d, запрошено: %d)", err, product.GetQuantity(), qty)
		return err
	}

	oldQuantity := product.GetQuantity()
	product.SetQuantity(product.GetQuantity() - qty)

	if product.GetQuantity() == 0 {
		if err := s.repo.Delete(id); err != nil {
			log.Printf("❌ Ошибка удаления товара из репозитория: %v", err)
			return err
		}
		log.Printf("✅ Товар '%s' (ID: %s) полностью удален со склада (было: %d)", product.GetName(), id, oldQuantity)
	} else {
		if err := s.repo.Update(product); err != nil {
			log.Printf("❌ Ошибка обновления товара: %v", err)
			return err
		}
		log.Printf("✅ Товар '%s' (ID: %s): количество уменьшено с %d до %d", product.GetName(), id, oldQuantity, product.GetQuantity())
	}

	// Уведомляем наблюдателей об удалении
	s.notifier.Notify(observer.Event{
		Type:      observer.ProductRemoved,
		Product:   product,
		Quantity:  qty,
		Timestamp: time.Now(),
		Message:   fmt.Sprintf("Удалено %d единиц товара", qty),
	})

	// Проверка низкого остатка (LowStock)
	if product.GetQuantity() > 0 && product.GetQuantity() <= 5 {
		log.Printf("⚠️ Низкий остаток товара '%s': осталось %d шт.", product.GetName(), product.GetQuantity())
		s.notifier.Notify(observer.Event{
			Type:      observer.LowStock,
			Product:   product,
			Quantity:  product.GetQuantity(),
			Timestamp: time.Now(),
			Message:   "Остаток на складе ниже порогового значения",
		})
	}

	return nil
}

func (s *WarehouseService) GetAllProducts() []domain.Product {
	products := s.repo.FindAll()
	log.Printf("📦 Запрошен список всех товаров: найдено %d позиций", len(products))
	return products
}

func (s *WarehouseService) GetProductsByCategory(category string) []domain.Product {
	products := s.repo.FindByCategory(category)
	log.Printf("📂 Запрошены товары категории '%s': найдено %d позиций", category, len(products))
	return products
}

func (s *WarehouseService) GetStatistics() Statistics {
	items := s.repo.FindAll()
	stats := Statistics{
		Categories: make(map[string]int),
	}

	for _, item := range items {
		stats.TotalItems += item.GetQuantity()
		stats.TotalValue += float64(item.GetQuantity()) * item.GetPrice()
		stats.Categories[item.GetCategory()] += item.GetQuantity()
	}

	log.Printf("📊 Сформирована статистика склада: всего %d товаров на сумму %.2f руб.",
		stats.TotalItems, stats.TotalValue)

	return stats
}

func (s *WarehouseService) GetLuckyLottery() *LuckyLottery {
	return s.luckyLottery
}

// GetProductByID возвращает товар по ID (удобный метод для внешнего использования)
func (s *WarehouseService) GetProductByID(id string) (domain.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		log.Printf("❌ Товар с ID '%s' не найден", id)
		return nil, err
	}
	return product, nil
}

// UpdateProduct обновляет информацию о товаре
func (s *WarehouseService) UpdateProduct(product domain.Product) error {
	if err := s.validator.Validate(product); err != nil {
		log.Printf("❌ Ошибка валидации при обновлении: %v", err)
		return err
	}

	if err := s.repo.Update(product); err != nil {
		log.Printf("❌ Ошибка обновления товара: %v", err)
		return err
	}

	log.Printf("✅ Товар '%s' (ID: %s) успешно обновлен", product.GetName(), product.GetID())
	return nil
}
