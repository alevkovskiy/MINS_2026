package handler

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"warehouse_system/internal/domain"
	"warehouse_system/internal/factory"
	"warehouse_system/internal/repository"
	"warehouse_system/internal/service"
	customerrors "warehouse_system/pkg/errors"
)

type ConsoleHandler struct {
	service    *service.WarehouseService
	repo       repository.Repository
	checker    service.CompatibilityChecker
	reader     *bufio.Reader
	factoryReg *factory.ProductFactoryRegistry
}

func NewConsoleHandler(svc *service.WarehouseService, repo repository.Repository, checker service.CompatibilityChecker) *ConsoleHandler {
	return &ConsoleHandler{
		service:    svc,
		repo:       repo,
		checker:    checker,
		reader:     bufio.NewReader(os.Stdin),
		factoryReg: factory.NewProductFactoryRegistry(),
	}
}

func (h *ConsoleHandler) Run() error {
	if jsonRepo, ok := h.repo.(*repository.JSONRepository); ok {
		if err := jsonRepo.Load(); err != nil {
			fmt.Printf("Не удалось загрузить данные: %v\n", err)
		}
	}

	if len(h.service.GetAllProducts()) == 0 {
		h.addTestData()
	}

	fmt.Println("СИСТЕМА УПРАВЛЕНИЯ СКЛАДОМ")
	fmt.Println(strings.Repeat("=", 50))

	for {
		h.printMenu()

		fmt.Print("\nВыберите действие (1-8): ")
		choice, _ := h.reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			h.addProductInteractive()
		case "2":
			h.removeProductInteractive()
		case "3":
			h.showStatistics()
		case "4":
			h.showAllProducts()
		case "5":
			h.checkCompatibility()
		case "6":
			h.calculateDiscount()
		case "7":
			h.saveAndExit()
			return nil
		case "8":
			h.showLotteryInfo()
		default:
			fmt.Println("\nНеверный выбор. Попробуйте снова.")
		}

		fmt.Println("\nНажмите Enter для продолжения...")
		h.reader.ReadString('\n')
	}
}

func (h *ConsoleHandler) printMenu() {
	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("ГЛАВНОЕ МЕНЮ")
	fmt.Println("1.  Добавить товар")
	fmt.Println("2.  Удалить товар")
	fmt.Println("3.  Показать статистику")
	fmt.Println("4.  Список всех товаров")
	fmt.Println("5.  Проверить совместимость")
	fmt.Println("6.  Рассчитать стоимость со скидкой")
	fmt.Println("7.  Сохранить и выйти")
	fmt.Println("8.  🎰 ЛОТЕРЕЯ 'СЧАСТЛИВОЕ ЧИСЛО 13' 🎰")
	fmt.Println(strings.Repeat("-", 50))
}

func (h *ConsoleHandler) addProductInteractive() {
	fmt.Println("\nДОБАВЛЕНИЕ НОВОГО ТОВАРА")

	lottery := h.service.GetLuckyLottery()
	if lottery.IsLuckyMoment() {
		fmt.Println("\n🎰🎰🎰 ВНИМАНИЕ! СЕЙЧАС СЧАСТЛИВОЕ ВРЕМЯ! 🎰🎰🎰")
		fmt.Println("   Добавьте товар и, возможно, выиграете джек-пот!")
		fmt.Println("   " + lottery.GetJackpotStatus())
		fmt.Println()
	}

	categories := h.factoryReg.GetCategories()
	fmt.Println("\nДоступные категории:")
	for i, cat := range categories {
		emoji := h.getCategoryEmoji(cat)
		fmt.Printf("  %d. %s %s\n", i+1, emoji, cat)
	}

	fmt.Print("\nВыберите категорию (1-3): ")
	catChoice, _ := h.reader.ReadString('\n')
	catChoice = strings.TrimSpace(catChoice)

	catIndex, err := strconv.Atoi(catChoice)
	if err != nil || catIndex < 1 || catIndex > len(categories) {
		fmt.Println("Неверная категория!")
		return
	}

	category := categories[catIndex-1]

	productFactory, err := h.factoryReg.GetFactory(category)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Print("Введите ID товара: ")
	id, _ := h.reader.ReadString('\n')
	id = strings.TrimSpace(id)

	if h.repo.Exists(id) {
		fmt.Println("Товар с таким ID уже существует!")
		return
	}

	fmt.Print("Введите название товара: ")
	name, _ := h.reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("Введите количество: ")
	qtyStr, _ := h.reader.ReadString('\n')
	qtyStr = strings.TrimSpace(qtyStr)
	quantity, err := strconv.Atoi(qtyStr)
	if err != nil || quantity <= 0 {
		fmt.Println("Некорректное количество!")
		return
	}

	fmt.Print("Введите цену: ")
	priceStr, _ := h.reader.ReadString('\n')
	priceStr = strings.TrimSpace(priceStr)
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		fmt.Println("Некорректная цена!")
		return
	}

	params := make(map[string]interface{})

	switch category {
	case "Food":
		fmt.Print("Введите срок годности (ГГГГ-ММ-ДД): ")
		expiry, _ := h.reader.ReadString('\n')
		expiry = strings.TrimSpace(expiry)
		params["expiry_date"] = expiry

	case "Electronics":
		fmt.Print("Введите срок гарантии (месяцев): ")
		warrantyStr, _ := h.reader.ReadString('\n')
		warrantyStr = strings.TrimSpace(warrantyStr)
		warranty, _ := strconv.Atoi(warrantyStr)
		params["warranty_months"] = warranty

	case "Chemicals":
		fmt.Print("Введите уровень опасности (1-5): ")
		hazardStr, _ := h.reader.ReadString('\n')
		hazardStr = strings.TrimSpace(hazardStr)
		hazard, _ := strconv.Atoi(hazardStr)
		params["hazard_level"] = hazard
	}

	product, err := productFactory.CreateProduct(id, name, quantity, price, params)
	if err != nil {
		fmt.Printf("Ошибка создания товара: %v\n", err)
		return
	}

	fmt.Println("\n🔍 Проверка совместимости...")
	if err := h.service.AddProduct(product); err != nil {
		switch e := err.(type) {
		case *customerrors.CompatibilityError:
			fmt.Printf("Ошибка совместимости: %v\n", e)
		case *customerrors.ValidationError:
			fmt.Printf("Ошибка валидации: %v\n", e)
		default:
			fmt.Printf("Ошибка: %v\n", e)
		}
	} else {
		fmt.Printf("Товар '%s' успешно добавлен на склад!\n", name)
	}
}

func (h *ConsoleHandler) removeProductInteractive() {
	fmt.Println("\nУДАЛЕНИЕ ТОВАРА")

	h.showAllProducts()

	fmt.Print("\nВведите ID товара для удаления: ")
	id, _ := h.reader.ReadString('\n')
	id = strings.TrimSpace(id)

	fmt.Print("Введите количество для удаления: ")
	qtyStr, _ := h.reader.ReadString('\n')
	qtyStr = strings.TrimSpace(qtyStr)
	quantity, err := strconv.Atoi(qtyStr)
	if err != nil || quantity <= 0 {
		fmt.Println("Некорректное количество!")
		return
	}

	if err := h.service.RemoveProduct(id, quantity); err != nil {
		switch e := err.(type) {
		case *customerrors.NotFoundError:
			fmt.Printf("❌ %v\n", e)
		case *customerrors.StockError:
			fmt.Printf("❌ %v\n", e)
		default:
			fmt.Printf("Ошибка: %v\n", e)
		}
	} else {
		if quantity == 1 {
			fmt.Printf("Удалена 1 единица товара %s\n", id)
		} else {
			fmt.Printf("Удалено %d единиц товара %s\n", quantity, id)
		}
	}
}

func (h *ConsoleHandler) showStatistics() {
	fmt.Println("\nСТАТИСТИКА СКЛАДА")
	fmt.Println(strings.Repeat("-", 50))

	stats := h.service.GetStatistics()

	fmt.Printf("Всего товаров: %d шт.\n", stats.TotalItems)
	fmt.Printf("Общая стоимость: %.2f руб.\n", stats.TotalValue)

	if stats.TotalItems > 0 {
		fmt.Printf("Средняя цена: %.2f руб.\n", stats.TotalValue/float64(stats.TotalItems))
	}

	fmt.Println("\nПо категориям:")

	if len(stats.Categories) == 0 {
		fmt.Println("Нет товаров")
	} else {
		for cat, count := range stats.Categories {
			emoji := h.getCategoryEmoji(cat)
			fmt.Printf("  %s %s: %d шт.\n", emoji, cat, count)
		}
	}
}

func (h *ConsoleHandler) showAllProducts() {
	fmt.Println("\nВСЕ ТОВАРЫ НА СКЛАДЕ (в порядке добавления)")
	fmt.Println(strings.Repeat("-", 50))

	products := h.service.GetAllProducts()
	if len(products) == 0 {
		fmt.Println("Склад пуст")
		return
	}

	fmt.Printf("Всего товаров: %d\n\n", len(products))

	for i, p := range products {
		emoji := h.getCategoryEmoji(p.GetCategory())
		fmt.Printf("%d. %s %s\n", i+1, emoji, p.GetName())
		fmt.Printf("ID: %s\n", p.GetID())
		fmt.Printf("Категория: %s\n", p.GetCategory())
		fmt.Printf("Количество: %d\n", p.GetQuantity())
		fmt.Printf("Цена: %.2f руб.\n", p.GetPrice())

		switch v := p.(type) {
		case *domain.Food:
			fmt.Printf("Срок годности: %s\n", v.ExpiryDate)
		case *domain.Electronics:
			fmt.Printf("Гарантия: %d мес.\n", v.WarrantyMonths)
		case *domain.Chemicals:
			fmt.Printf("Уровень опасности: %d\n", v.HazardLevel)
		}
		fmt.Println()
	}
}

func (h *ConsoleHandler) checkCompatibility() {
	fmt.Println("\nПРОВЕРКА СОВМЕСТИМОСТИ")
	fmt.Println(strings.Repeat("-", 50))

	fmt.Println("Выберите категорию для проверки:")
	fmt.Println("1. Еда")
	fmt.Println("2. Электроника")
	fmt.Println("3. Химия")

	fmt.Print("\nВыбор: ")
	choice, _ := h.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	categories := map[string]string{
		"1": "Food",
		"2": "Electronics",
		"3": "Chemicals",
	}

	category, ok := categories[choice]
	if !ok {
		fmt.Println("Неверный выбор!")
		return
	}

	testProduct := &domain.BaseProduct{
		ID:       "test",
		Name:     "Тестовый товар",
		Category: category,
		Quantity: 1,
		Price:    0,
	}

	if err := h.service.CheckProductCompatibility(testProduct); err != nil {
		fmt.Printf("Нарушение совместимости: %v\n", err)
	} else {
		fmt.Printf("Товары категории '%s' совместимы со всеми товарами на складе\n", category)
	}
}

func (h *ConsoleHandler) calculateDiscount() {
	fmt.Println("\nРАСЧЕТ СТОИМОСТИ СО СКИДКОЙ")
	fmt.Println(strings.Repeat("-", 50))

	products := h.service.GetAllProducts()
	if len(products) == 0 {
		fmt.Println("Склад пуст, нечего продавать")
		return
	}

	h.showAllProducts()

	fmt.Print("\nВведите ID товара: ")
	id, _ := h.reader.ReadString('\n')
	id = strings.TrimSpace(id)

	product, err := h.repo.FindByID(id)
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		return
	}

	fmt.Print("Введите количество: ")
	qtyStr, _ := h.reader.ReadString('\n')
	qtyStr = strings.TrimSpace(qtyStr)
	quantity, err := strconv.Atoi(qtyStr)
	if err != nil || quantity <= 0 {
		fmt.Println("Некорректное количество!")
		return
	}

	if product.GetQuantity() < quantity {
		fmt.Printf("Недостаточно товара на складе (доступно: %d)\n", product.GetQuantity())
		return
	}

	lottery := h.service.GetLuckyLottery()

	if lottery.IsLuckyMoment() {
		fmt.Println("\n🎰 ДОПОЛНИТЕЛЬНЫЙ ВАРИАНТ:")
		fmt.Println("7.  Лотерейная скидка 13% (только для счастливчиков!)")
	}

	fmt.Println("\nВыберите стратегию скидки:")
	fmt.Println("1. Без скидки")
	fmt.Println("2. Оптовая скидка (от 10 шт. - 10%)")
	fmt.Println("3. Оптовая скидка (от 20 шт. - 15%)")
	fmt.Println("4. Категорийная скидка (Food - 5%)")
	fmt.Println("5. Категорийная скидка (Electronics - 3%)")
	fmt.Println("6. Сезонная скидка (8%)")

	fmt.Print("\nВыбор: ")
	choice, _ := h.reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	if lottery.IsLuckyMoment() && choice == "7" {
		discount := lottery.CalculateLuckyDiscount(product, quantity)
		if discount > 0 {
			total := product.GetPrice() * float64(quantity)
			finalTotal := total - discount
			fmt.Println("\n" + strings.Repeat("-", 50))
			fmt.Println("🎰 ЛОТЕРЕЙНАЯ СКИДКА АКТИВИРОВАНА! 🎰")
			fmt.Printf("   Товар: %s\n", product.GetName())
			fmt.Printf("   ID товара: %s\n", product.GetID())
			fmt.Printf("   Цена за единицу: %.2f руб.\n", product.GetPrice())
			fmt.Printf("   Количество: %d шт.\n", quantity)
			fmt.Printf("   Сумма скидки 13%%: %.2f руб.\n", discount)
			fmt.Println(strings.Repeat("-", 50))
			fmt.Printf("ИТОГО К ОПЛАТЕ: %.2f руб.\n", finalTotal)
			fmt.Println(strings.Repeat("-", 50))
			return
		}
	}

	var strategy service.DiscountStrategy

	switch choice {
	case "1":
		strategy = service.NewNoDiscountStrategy()
	case "2":
		strategy = service.NewBulkDiscountStrategy(10, 10.0)
	case "3":
		strategy = service.NewBulkDiscountStrategy(20, 15.0)
	case "4":
		strategy = service.NewCategoryDiscountStrategy("Food", 5.0)
	case "5":
		strategy = service.NewCategoryDiscountStrategy("Electronics", 3.0)
	case "6":
		strategy = service.NewSeasonDiscountStrategy(8.0)
	default:
		fmt.Println("Неверный выбор!")
		return
	}

	fmt.Printf("\nВыбрана стратегия: %s\n", strategy.GetName())

	calculator := service.NewDiscountCalculator(strategy)
	discount := calculator.CalculateDiscount(product, quantity)
	total := product.GetPrice() * float64(quantity)
	finalTotal := total - discount

	fmt.Println("\n" + strings.Repeat("-", 50))
	fmt.Println("РЕЗУЛЬТАТ РАСЧЕТА:")
	fmt.Printf("   Товар: %s\n", product.GetName())
	fmt.Printf("   ID товара: %s\n", product.GetID())
	fmt.Printf("   Цена за единицу: %.2f руб.\n", product.GetPrice())
	fmt.Printf("   Количество: %d шт.\n", quantity)
	fmt.Printf("   Стратегия скидки: %s\n", strategy.GetName())
	fmt.Printf("   Сумма скидки: %.2f руб.\n", discount)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("ИТОГО К ОПЛАТЕ: %.2f руб.\n", finalTotal)
	fmt.Println(strings.Repeat("-", 50))
}

func (h *ConsoleHandler) addTestData() {
	foodFactory := factory.NewFoodFactory()
	electronicsFactory := factory.NewElectronicsFactory()

	food, _ := foodFactory.CreateProduct("1", "Молоко", 10, 100.0, map[string]interface{}{
		"expiry_date": "2026-04-05",
	})

	electronics, _ := electronicsFactory.CreateProduct("4", "Ноутбук", 5, 50000.0, map[string]interface{}{
		"warranty_months": 12,
	})

	chemicalsFactory := factory.NewChemicalsFactory()
	chemicals, _ := chemicalsFactory.CreateProduct("7", "Бытовая химия", 15, 250.0, map[string]interface{}{
		"hazard_level": 3,
	})

	h.service.AddProduct(food)
	h.service.AddProduct(electronics)
	h.service.AddProduct(chemicals)
}

func (h *ConsoleHandler) saveAndExit() {
	fmt.Println("\n💾 Сохранение данных...")
	if jsonRepo, ok := h.repo.(*repository.JSONRepository); ok {
		if err := jsonRepo.SaveToFile(); err != nil {
			fmt.Printf("Ошибка сохранения: %v\n", err)
		} else {
			fmt.Println("Данные успешно сохранены")
		}
	}
	fmt.Println("\nДо свидания!")
}

func (h *ConsoleHandler) getCategoryEmoji(category string) string {
	switch category {
	case "Food":
		return "🍎"
	case "Electronics":
		return "💻"
	case "Chemicals":
		return "🧪"
	default:
		return "📦"
	}
}

func (h *ConsoleHandler) showLotteryInfo() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("🎰🎰🎰 СКЛАДСКАЯ ЛОТЕРЕЯ 'СЧАСТЛИВОЕ ЧИСЛО 13' 🎰🎰🎰")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	lottery := h.service.GetLuckyLottery()

	fmt.Println("		ПРАВИЛА ЛОТЕРЕИ:")
	fmt.Println("		Если сегодня 13-е число месяца,")
	fmt.Println("		сейчас 13 часов 13 минут,")
	fmt.Println("		и вы добавляете 13-й товар за сегодня —")
	fmt.Println("			ВЫ ВЫИГРЫВАЕТЕ ДЖЕК-ПОТ!")
	fmt.Println("		ПРИЗЫ ДЖЕК-ПОТА:")
	fmt.Println("   • Количество товара увеличивается на 13 шт.")
	fmt.Println("   • Цена товара становится 13.13 руб. (любая цена была!)")
	fmt.Println("   • Название получает суффикс 'СЧАСТЛИВЫЙ БИЛЕТ'")
	fmt.Println("   • Автоматическая скидка 13% на этот товар")
	fmt.Println("		ВАЖНЫЕ УСЛОВИЯ:")
	fmt.Println("   • Джек-пот разыгрывается только 1 раз в день")
	fmt.Println("   • В 13:14 магия исчезает до следующего месяца")
	fmt.Println("   • Никто не знает, почему именно 13")
	fmt.Println("		СТАТУС ЛОТЕРЕИ:")
	fmt.Println("   " + lottery.GetJackpotStatus())

	if lottery.IsLuckyMoment() {
		fmt.Println()
		fmt.Println(strings.Repeat("!", 60))
		fmt.Println("🎲🎲🎲 СЕЙЧАС 13:13 13-ГО ЧИСЛА! СЧАСТЛИВОЕ ВРЕМЯ! 🎲🎲🎲")
		fmt.Println("   Добавьте 13-й товар и выиграйте джек-пот!")
		fmt.Println(strings.Repeat("!", 60))
	} else {
		now := time.Now()
		if now.Day() < 13 {
			fmt.Printf("\n⏰ Следующий шанс: %d-го числа в 13:13\n", 13)
		} else if now.Day() > 13 {
			fmt.Printf("\n⏰ Следующий шанс: 13-го числа следующего месяца в 13:13\n")
		} else if now.Hour() < 13 {
			fmt.Printf("\n⏰ Следующий шанс: сегодня в 13:13\n")
		} else if now.Hour() > 13 {
			fmt.Printf("\n⏰ Следующий шанс: 13-го числа следующего месяца в 13:13\n")
		} else if now.Minute() < 13 {
			fmt.Printf("\n⏰ Следующий шанс: через %d минут (в 13:13)\n", 13-now.Minute())
		} else {
			fmt.Printf("\n⏰ Следующий шанс: 13-го числа следующего месяца в 13:13\n")
		}
	}

}
