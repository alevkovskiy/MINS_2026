package domain

import (
	"fmt"
	customerrors "warehouse_system/pkg/errors"
)

type ProductFactory interface {
	CreateProduct(id, name string, quantity int, price float64, params map[string]interface{}) (Product, error)
	GetCategory() string
}

type FoodFactory struct{}

func NewFoodFactory() *FoodFactory {
	return &FoodFactory{}
}

func (f *FoodFactory) CreateProduct(id, name string, quantity int, price float64, params map[string]interface{}) (Product, error) {
	expiryDate, ok := params["expiry_date"].(string)
	if !ok || expiryDate == "" {
		return nil, customerrors.NewValidationError("не указан срок годности", "expiry_date")
	}

	return &Food{
		BaseProduct: BaseProduct{
			ID:       id,
			Name:     name,
			Category: "Food",
			Quantity: quantity,
			Price:    price,
		},
		ExpiryDate: expiryDate,
	}, nil
}

func (f *FoodFactory) GetCategory() string {
	return "Food"
}

type ElectronicsFactory struct{}

func NewElectronicsFactory() *ElectronicsFactory {
	return &ElectronicsFactory{}
}

func (e *ElectronicsFactory) CreateProduct(id, name string, quantity int, price float64, params map[string]interface{}) (Product, error) {
	warranty, ok := params["warranty_months"].(int)
	if !ok {
		warranty = 0
	}

	return &Electronics{
		BaseProduct: BaseProduct{
			ID:       id,
			Name:     name,
			Category: "Electronics",
			Quantity: quantity,
			Price:    price,
		},
		WarrantyMonths: warranty,
	}, nil
}

func (e *ElectronicsFactory) GetCategory() string {
	return "Electronics"
}

type ChemicalsFactory struct{}

func NewChemicalsFactory() *ChemicalsFactory {
	return &ChemicalsFactory{}
}

func (c *ChemicalsFactory) CreateProduct(id, name string, quantity int, price float64, params map[string]interface{}) (Product, error) {
	hazardLevel, ok := params["hazard_level"].(int)
	if !ok {
		hazardLevel = 1
	}

	if hazardLevel < 1 || hazardLevel > 5 {
		return nil, customerrors.NewValidationError("уровень опасности должен быть от 1 до 5", "hazard_level")
	}

	return &Chemicals{
		BaseProduct: BaseProduct{
			ID:       id,
			Name:     name,
			Category: "Chemicals",
			Quantity: quantity,
			Price:    price,
		},
		HazardLevel: hazardLevel,
	}, nil
}

func (c *ChemicalsFactory) GetCategory() string {
	return "Chemicals"
}

type ProductFactoryRegistry struct {
	factories map[string]ProductFactory
}

func NewProductFactoryRegistry() *ProductFactoryRegistry {
	registry := &ProductFactoryRegistry{
		factories: make(map[string]ProductFactory),
	}

	registry.Register("Food", NewFoodFactory())
	registry.Register("Electronics", NewElectronicsFactory())
	registry.Register("Chemicals", NewChemicalsFactory())

	return registry
}

func (r *ProductFactoryRegistry) Register(category string, factory ProductFactory) {
	r.factories[category] = factory
}

func (r *ProductFactoryRegistry) GetFactory(category string) (ProductFactory, error) {
	factory, ok := r.factories[category]
	if !ok {
		return nil, fmt.Errorf("неизвестная категория: %s", category)
	}
	return factory, nil
}

func (r *ProductFactoryRegistry) GetCategories() []string {
	categories := make([]string, 0, len(r.factories))
	for cat := range r.factories {
		categories = append(categories, cat)
	}
	return categories
}
