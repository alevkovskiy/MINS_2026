package repository

import (
	"warehouse_system/internal/domain"
)

type Repository interface {
	Save(product domain.Product) error
	Update(product domain.Product) error
	Delete(id string) error
	FindByID(id string) (domain.Product, error)
	FindAll() []domain.Product
	FindByCategory(category string) []domain.Product
	Exists(id string) bool
	GetOrderedIDs() []string
}
