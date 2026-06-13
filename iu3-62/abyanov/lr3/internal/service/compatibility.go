package service

import (
	"warehouse_system/internal/domain"
	customerrors "warehouse_system/pkg/errors"
)

type CompatibilityChecker interface {
	Check(newProduct domain.Product, existing []domain.Product) error
}

type SimpleCompatibilityChecker struct{}

func NewSimpleCompatibilityChecker() *SimpleCompatibilityChecker {
	return &SimpleCompatibilityChecker{}
}

func (c *SimpleCompatibilityChecker) Check(newProduct domain.Product, existing []domain.Product) error {
	newCat := newProduct.GetCategory()
	for _, p := range existing {
		if (newCat == "Food" && p.GetCategory() == "Chemicals") ||
			(newCat == "Chemicals" && p.GetCategory() == "Food") {
			return customerrors.NewCompatibilityError(
				"нельзя хранить продукты питания рядом с химикатами",
				newProduct.GetID(),
				newProduct.GetName(),
			)
		}
	}
	return nil
}

type StrictCompatibilityChecker struct{}

func NewStrictCompatibilityChecker() *StrictCompatibilityChecker {
	return &StrictCompatibilityChecker{}
}

func (c *StrictCompatibilityChecker) Check(newProduct domain.Product, existing []domain.Product) error {
	newCat := newProduct.GetCategory()
	for _, p := range existing {
		if (newCat == "Food" && p.GetCategory() == "Chemicals") ||
			(newCat == "Chemicals" && p.GetCategory() == "Food") {
			return customerrors.NewCompatibilityError(
				"нельзя хранить продукты питания рядом с химикатами",
				newProduct.GetID(),
				newProduct.GetName(),
			)
		}
		if (newCat == "Electronics" && p.GetCategory() == "Chemicals") ||
			(newCat == "Chemicals" && p.GetCategory() == "Electronics") {
			return customerrors.NewCompatibilityError(
				"нельзя хранить электронику рядом с химикатами",
				newProduct.GetID(),
				newProduct.GetName(),
			)
		}
	}
	return nil
}
