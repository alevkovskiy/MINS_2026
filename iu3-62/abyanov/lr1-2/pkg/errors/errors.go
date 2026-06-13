package errors

import "fmt"

type BaseError struct {
	Message string
	Op      string
}

func (e *BaseError) Error() string {
	if e.Op != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	}
	return e.Message
}

type StockError struct {
	BaseError
	Available int
	Requested int
}

func NewStockError(msg string, available, requested int) *StockError {
	return &StockError{
		BaseError: BaseError{Message: msg, Op: "StockError"},
		Available: available,
		Requested: requested,
	}
}

func (e *StockError) Error() string {
	return fmt.Sprintf("ошибка склада: %s (доступно: %d, запрошено: %d)", e.Message, e.Available, e.Requested)
}

type CompatibilityError struct {
	BaseError
	ProductID   string
	ProductName string
}

func NewCompatibilityError(msg, productID, productName string) *CompatibilityError {
	return &CompatibilityError{
		BaseError:   BaseError{Message: msg, Op: "CompatibilityError"},
		ProductID:   productID,
		ProductName: productName,
	}
}

func (e *CompatibilityError) Error() string {
	return fmt.Sprintf("нарушение товарного соседства: %s (ID: %s, название: %s)", e.Message, e.ProductID, e.ProductName)
}

type NotFoundError struct {
	BaseError
	ID string
}

func NewNotFoundError(msg, id string) *NotFoundError {
	return &NotFoundError{
		BaseError: BaseError{Message: msg, Op: "NotFoundError"},
		ID:        id,
	}
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("товар не найден: ID %s. %s", e.ID, e.Message)
}

type ValidationError struct {
	BaseError
	Field string
}

func NewValidationError(msg, field string) *ValidationError {
	return &ValidationError{
		BaseError: BaseError{Message: msg, Op: "ValidationError"},
		Field:     field,
	}
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("ошибка валидации поля '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("ошибка валидации: %s", e.Message)
}
