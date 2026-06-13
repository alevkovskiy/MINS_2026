package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"warehouse_system/internal/domain"
	customerrors "warehouse_system/pkg/errors"
)

type JSONRepository struct {
	mu       sync.RWMutex
	items    map[string]domain.Product
	order    []string
	filePath string
}

func NewJSONRepository(filePath string) *JSONRepository {
	return &JSONRepository{
		items:    make(map[string]domain.Product),
		order:    make([]string, 0),
		filePath: filePath,
	}
}

func (r *JSONRepository) Save(product domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.items[product.GetID()]; ok {
		existing.SetQuantity(existing.GetQuantity() + product.GetQuantity())
		r.items[product.GetID()] = existing
	} else {
		r.items[product.GetID()] = product
		r.order = append(r.order, product.GetID())
	}

	return nil
}

func (r *JSONRepository) Update(product domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[product.GetID()]; !ok {
		return customerrors.NewNotFoundError("товар не найден для обновления", product.GetID())
	}

	r.items[product.GetID()] = product
	return nil
}

func (r *JSONRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.items[id]; !ok {
		return customerrors.NewNotFoundError("товар не найден", id)
	}

	delete(r.items, id)

	for i, itemID := range r.order {
		if itemID == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}

	return nil
}

func (r *JSONRepository) FindByID(id string) (domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, ok := r.items[id]
	if !ok {
		return nil, customerrors.NewNotFoundError("товар не найден", id)
	}
	return product, nil
}

func (r *JSONRepository) FindAll() []domain.Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Product, 0, len(r.order))
	for _, id := range r.order {
		if product, ok := r.items[id]; ok {
			result = append(result, product)
		}
	}
	return result
}

func (r *JSONRepository) FindByCategory(category string) []domain.Product {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]domain.Product, 0)
	for _, id := range r.order {
		if product, ok := r.items[id]; ok && product.GetCategory() == category {
			result = append(result, product)
		}
	}
	return result
}

func (r *JSONRepository) Exists(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.items[id]
	return ok
}

func (r *JSONRepository) GetOrderedIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, len(r.order))
	copy(result, r.order)
	return result
}

func (r *JSONRepository) Load() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, err := os.Stat(r.filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return fmt.Errorf("ошибка чтения файла: %w", err)
	}

	var storageData struct {
		Order []string                     `json:"order"`
		Items []domain.SerializableProduct `json:"items"`
	}

	if err := json.Unmarshal(data, &storageData); err != nil {
		return fmt.Errorf("ошибка десериализации: %w", err)
	}

	r.items = make(map[string]domain.Product)
	r.order = storageData.Order

	for _, sp := range storageData.Items {
		p := sp.ToProduct()
		if p != nil {
			r.items[p.GetID()] = p
		}
	}

	return nil
}

func (r *JSONRepository) SaveToFile() error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	serializable := make([]domain.SerializableProduct, 0, len(r.items))
	for _, id := range r.order {
		if p, ok := r.items[id]; ok {
			serializable = append(serializable, domain.FromProduct(p))
		}
	}

	storageData := struct {
		Order []string                     `json:"order"`
		Items []domain.SerializableProduct `json:"items"`
	}{
		Order: r.order,
		Items: serializable,
	}

	data, err := json.MarshalIndent(storageData, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации: %w", err)
	}

	if err := os.WriteFile(r.filePath, data, 0644); err != nil {
		return fmt.Errorf("ошибка записи файла: %w", err)
	}

	return nil
}
