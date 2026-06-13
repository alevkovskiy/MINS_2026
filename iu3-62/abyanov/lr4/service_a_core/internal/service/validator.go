package service

import (
	"time"
	"warehouse_system/internal/domain"
	customerrors "warehouse_system/pkg/errors"
)

const DateLayout = "2006-01-02"

type DefaultProductValidator struct{}

func NewDefaultProductValidator() *DefaultProductValidator {
	return &DefaultProductValidator{}
}

func (v *DefaultProductValidator) Validate(p domain.Product) error {
	if p == nil {
		return customerrors.NewValidationError("товар не может быть nil", "product")
	}

	if err := v.validateCommon(p); err != nil {
		return err
	}

	return v.validateSpecific(p)
}

func (v *DefaultProductValidator) validateCommon(p domain.Product) error {
	if p.GetID() == "" {
		return customerrors.NewValidationError("ID не может быть пустым", "id")
	}
	if p.GetName() == "" {
		return customerrors.NewValidationError("название не может быть пустым", "name")
	}
	if p.GetQuantity() < 0 {
		return customerrors.NewValidationError("количество не может быть отрицательным", "quantity")
	}
	if p.GetPrice() <= 0 {
		return customerrors.NewValidationError("цена должна быть положительной", "price")
	}
	return nil
}

func (v *DefaultProductValidator) validateSpecific(p domain.Product) error {
	switch p := p.(type) {
	case *domain.Food:
		return v.validateFood(p)
	case *domain.Electronics:
		return v.validateElectronics(p)
	case *domain.Chemicals:
		return v.validateChemicals(p)
	}
	return nil
}

func (v *DefaultProductValidator) validateFood(f *domain.Food) error {
	if _, err := time.Parse(DateLayout, f.ExpiryDate); err != nil {
		return customerrors.NewValidationError("неверный формат даты (ГГГГ-ММ-ДД)", "expiry_date")
	}
	return nil
}

func (v *DefaultProductValidator) validateElectronics(e *domain.Electronics) error {
	if e.WarrantyMonths < 0 {
		return customerrors.NewValidationError("гарантия не может быть отрицательной", "warranty_months")
	}
	return nil
}

func (v *DefaultProductValidator) validateChemicals(c *domain.Chemicals) error {
	if c.HazardLevel < 1 || c.HazardLevel > 5 {
		return customerrors.NewValidationError("уровень опасности должен быть от 1 до 5", "hazard_level")
	}
	return nil
}
