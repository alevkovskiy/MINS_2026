package observer

import (
	"fmt"
	"time"
	"warehouse_system/internal/domain"
)

type EventType string

const (
	ProductAdded   EventType = "PRODUCT_ADDED"
	ProductRemoved EventType = "PRODUCT_REMOVED"
	LowStock       EventType = "LOW_STOCK"
	ExpiringSoon   EventType = "EXPIRING_SOON"
)

type Event struct {
	Type      EventType
	Product   domain.Product
	Quantity  int
	Timestamp time.Time
	Message   string
}

type Observer interface {
	Update(event Event)
	GetName() string
}

type Subject interface {
	Attach(observer Observer)
	Detach(observer Observer)
	Notify(event Event)
}

type WarehouseNotifier struct {
	observers []Observer
}

func NewWarehouseNotifier() *WarehouseNotifier {
	return &WarehouseNotifier{
		observers: make([]Observer, 0),
	}
}

func (n *WarehouseNotifier) Attach(observer Observer) {
	n.observers = append(n.observers, observer)
	fmt.Printf("Наблюдатель '%s' подключен\n", observer.GetName())
}

func (n *WarehouseNotifier) Detach(observer Observer) {
	for i, obs := range n.observers {
		if obs.GetName() == observer.GetName() {
			n.observers = append(n.observers[:i], n.observers[i+1:]...)
			fmt.Printf("Наблюдатель '%s' отключен\n", observer.GetName())
			break
		}
	}
}

func (n *WarehouseNotifier) Notify(event Event) {
	for _, observer := range n.observers {
		observer.Update(event)
	}
}

type ConsoleLogger struct {
	name string
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{name: "ConsoleLogger"}
}

func (l *ConsoleLogger) GetName() string {
	return l.name
}

func (l *ConsoleLogger) Update(event Event) {
	timestamp := event.Timestamp.Format("15:04:05")
	switch event.Type {
	case ProductAdded:
		fmt.Printf("\n[%s] %s: Добавлен товар '%s' (ID: %s) в количестве %d шт.\n",
			timestamp, l.name, event.Product.GetName(), event.Product.GetID(), event.Quantity)
	case ProductRemoved:
		fmt.Printf("\n[%s] %s: Удален товар '%s' (ID: %s) в количестве %d шт.\n",
			timestamp, l.name, event.Product.GetName(), event.Product.GetID(), event.Quantity)
	case LowStock:
		fmt.Printf("\n[%s] %s: ВНИМАНИЕ! Товар '%s' (ID: %s) заканчивается! Осталось: %d шт.\n",
			timestamp, l.name, event.Product.GetName(), event.Product.GetID(), event.Product.GetQuantity())
	case ExpiringSoon:
		fmt.Printf("\n[%s] %s: ВНИМАНИЕ! Товар '%s' (ID: %s) скоро истекает! Срок годности: %s\n",
			timestamp, l.name, event.Product.GetName(), event.Product.GetID(), event.Message)
	}
}

type LowStockAlert struct {
	name      string
	threshold int
}

func NewLowStockAlert(threshold int) *LowStockAlert {
	return &LowStockAlert{
		name:      fmt.Sprintf("LowStockAlert(%d)", threshold),
		threshold: threshold,
	}
}

func (l *LowStockAlert) GetName() string {
	return l.name
}

func (l *LowStockAlert) Update(event Event) {
	if event.Type == ProductRemoved && event.Product.GetQuantity() <= l.threshold && event.Product.GetQuantity() > 0 {
		fmt.Printf("\n[%s] Товар '%s' достиг порогового значения! Осталось: %d шт. (порог: %d шт.)\n",
			l.name, event.Product.GetName(), event.Product.GetQuantity(), l.threshold)
	}
}

type ExpiryTracker struct {
	name string
	days int
}

func NewExpiryTracker(days int) *ExpiryTracker {
	return &ExpiryTracker{
		name: fmt.Sprintf("ExpiryTracker(%d дней)", days),
		days: days,
	}
}

func (e *ExpiryTracker) GetName() string {
	return e.name
}

func (e *ExpiryTracker) Update(event Event) {
	if event.Type == ProductAdded {
		if food, ok := event.Product.(*domain.Food); ok {
			expiryDate, err := time.Parse("2006-01-02", food.ExpiryDate)
			if err == nil {
				daysUntilExpiry := int(time.Until(expiryDate).Hours() / 24)
				if daysUntilExpiry >= 0 && daysUntilExpiry <= e.days {
					fmt.Printf("\n[%s] Товар '%s' истекает через %d дней! (срок: %s)\n",
						e.name, event.Product.GetName(), daysUntilExpiry, food.ExpiryDate)
				}
			}
		}
	}
}
