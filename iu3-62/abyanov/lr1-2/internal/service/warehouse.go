package service

import (
	"time"
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
	repo      repository.Repository
	checker   CompatibilityChecker
	validator ProductValidator
	notifier  *observer.WarehouseNotifier
}

func NewWarehouseService(
	repo repository.Repository,
	checker CompatibilityChecker,
	validator ProductValidator,
) *WarehouseService {
	return &WarehouseService{
		repo:      repo,
		checker:   checker,
		validator: validator,
		notifier:  observer.NewWarehouseNotifier(),
	}
}

func (s *WarehouseService) AttachObserver(obs observer.Observer) {
	s.notifier.Attach(obs)
}

func (s *WarehouseService) DetachObserver(obs observer.Observer) {
	s.notifier.Detach(obs)
}

func (s *WarehouseService) AddProduct(p domain.Product) error {
	if err := s.validator.Validate(p); err != nil {
		return err
	}

	allItems := s.repo.FindAll()
	if err := s.checker.Check(p, allItems); err != nil {
		return err
	}

	if err := s.repo.Save(p); err != nil {
		return err
	}

	s.notifier.Notify(observer.Event{
		Type:      observer.ProductAdded,
		Product:   p,
		Quantity:  p.GetQuantity(),
		Timestamp: time.Now(),
		Message:   "Товар добавлен на склад",
	})

	return nil
}

func (s *WarehouseService) RemoveProduct(id string, qty int) error {
	if qty <= 0 {
		return customerrors.NewValidationError("количество должно быть положительным", "quantity")
	}

	product, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	if product.GetQuantity() < qty {
		return customerrors.NewStockError("недостаточное количество", product.GetQuantity(), qty)
	}

	product.SetQuantity(product.GetQuantity() - qty)

	if product.GetQuantity() == 0 {
		if err := s.repo.Delete(id); err != nil {
			return err
		}
	} else {
		if err := s.repo.Update(product); err != nil {
			return err
		}
	}

	s.notifier.Notify(observer.Event{
		Type:      observer.ProductRemoved,
		Product:   product,
		Quantity:  qty,
		Timestamp: time.Now(),
		Message:   "Товар удален со склада",
	})

	if product.GetQuantity() > 0 && product.GetQuantity() <= 5 {
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
	return s.repo.FindAll()
}

func (s *WarehouseService) GetProductsByCategory(category string) []domain.Product {
	return s.repo.FindByCategory(category)
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

	return stats
}

func (s *WarehouseService) GetCompatibilityChecker() CompatibilityChecker {
	return s.checker
}

func (s *WarehouseService) CheckProductCompatibility(p domain.Product) error {
	allItems := s.repo.FindAll()
	return s.checker.Check(p, allItems)
}
