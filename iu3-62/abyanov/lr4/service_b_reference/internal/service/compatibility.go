package service

import (
	"sync"
	"time"
)

type CompatibilityChecker struct {
	mu          sync.RWMutex
	conflicts   map[string][]string
	startTime   time.Time
	failureMode bool
}

func NewCompatibilityChecker() *CompatibilityChecker {
	return &CompatibilityChecker{
		conflicts: map[string][]string{
			"Food":        {"Chemicals"},
			"Chemicals":   {"Food", "Electronics"},
			"Electronics": {"Chemicals"},
		},
		startTime:   time.Now(),
		failureMode: false,
	}
}

func (c *CompatibilityChecker) Check(category string, existingCategories []string) (bool, []string, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Проверка на failure mode для тестирования ошибок
	if c.failureMode {
		return false, nil, "service temporarily unavailable"
	}

	conflicting, exists := c.conflicts[category]
	if !exists {
		return true, nil, ""
	}

	incompatible := make([]string, 0)
	conflictMap := make(map[string]bool)
	for _, conflict := range conflicting {
		conflictMap[conflict] = true
	}

	for _, existing := range existingCategories {
		if conflictMap[existing] {
			incompatible = append(incompatible, existing)
		}
	}

	if len(incompatible) > 0 {
		return false, incompatible, "Товар категории '" + category + "' несовместим с категориями: " + joinCategories(incompatible)
	}

	return true, nil, ""
}

func (c *CompatibilityChecker) GetCategories() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	categories := make([]string, 0, len(c.conflicts))
	for cat := range c.conflicts {
		categories = append(categories, cat)
	}
	return categories
}

func (c *CompatibilityChecker) AddConflictRule(category string, conflictsWith []string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.conflicts[category] = conflictsWith
}

func (c *CompatibilityChecker) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Проверка: сервис здоров, если не в failure mode и запущен менее 24 часов
	uptime := time.Since(c.startTime)
	return !c.failureMode && uptime < 24*time.Hour
}

func (c *CompatibilityChecker) SetFailureMode(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failureMode = enabled
}

func joinCategories(cats []string) string {
	result := ""
	for i, cat := range cats {
		if i > 0 {
			result += ", "
		}
		result += cat
	}
	return result
}
