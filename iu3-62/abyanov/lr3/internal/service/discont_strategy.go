package service

import (
	"fmt"
	"warehouse_system/internal/domain"
)

type DiscountStrategy interface {
	Calculate(product domain.Product, quantity int) float64
	GetName() string
}

type NoDiscountStrategy struct{}

func NewNoDiscountStrategy() *NoDiscountStrategy {
	return &NoDiscountStrategy{}
}

func (s *NoDiscountStrategy) Calculate(product domain.Product, quantity int) float64 {
	return 0
}

func (s *NoDiscountStrategy) GetName() string {
	return "Без скидки"
}

type BulkDiscountStrategy struct {
	MinQuantity     int
	DiscountPercent float64
}

func NewBulkDiscountStrategy(minQuantity int, discountPercent float64) *BulkDiscountStrategy {
	return &BulkDiscountStrategy{
		MinQuantity:     minQuantity,
		DiscountPercent: discountPercent,
	}
}

func (s *BulkDiscountStrategy) Calculate(product domain.Product, quantity int) float64 {
	if quantity >= s.MinQuantity {
		return product.GetPrice() * float64(quantity) * s.DiscountPercent / 100
	}
	return 0
}

func (s *BulkDiscountStrategy) GetName() string {
	return fmt.Sprintf("Оптовая скидка %d%% (от %d шт.)", int(s.DiscountPercent), s.MinQuantity)
}

type CategoryDiscountStrategy struct {
	Category        string
	DiscountPercent float64
}

func NewCategoryDiscountStrategy(category string, discountPercent float64) *CategoryDiscountStrategy {
	return &CategoryDiscountStrategy{
		Category:        category,
		DiscountPercent: discountPercent,
	}
}

func (s *CategoryDiscountStrategy) Calculate(product domain.Product, quantity int) float64 {
	if product.GetCategory() == s.Category {
		return product.GetPrice() * float64(quantity) * s.DiscountPercent / 100
	}
	return 0
}

func (s *CategoryDiscountStrategy) GetName() string {
	return fmt.Sprintf("Категорийная скидка %d%% (%s)", int(s.DiscountPercent), s.Category)
}

type SeasonDiscountStrategy struct {
	DiscountPercent float64
}

func NewSeasonDiscountStrategy(discountPercent float64) *SeasonDiscountStrategy {
	return &SeasonDiscountStrategy{
		DiscountPercent: discountPercent,
	}
}

func (s *SeasonDiscountStrategy) Calculate(product domain.Product, quantity int) float64 {
	return product.GetPrice() * float64(quantity) * s.DiscountPercent / 100
}

func (s *SeasonDiscountStrategy) GetName() string {
	return fmt.Sprintf("Сезонная скидка %d%%", int(s.DiscountPercent))
}

type DiscountCalculator struct {
	strategy DiscountStrategy
}

func NewDiscountCalculator(strategy DiscountStrategy) *DiscountCalculator {
	return &DiscountCalculator{
		strategy: strategy,
	}
}

func (c *DiscountCalculator) SetStrategy(strategy DiscountStrategy) {
	c.strategy = strategy
}

func (c *DiscountCalculator) GetStrategy() DiscountStrategy {
	return c.strategy
}

func (c *DiscountCalculator) CalculateDiscount(product domain.Product, quantity int) float64 {
	if c.strategy == nil {
		return 0
	}
	return c.strategy.Calculate(product, quantity)
}
