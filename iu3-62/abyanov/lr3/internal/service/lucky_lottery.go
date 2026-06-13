package service

import (
	"fmt"
	"time"
	"warehouse_system/internal/domain"
)

type LuckyLottery struct {
	todayCounter int
	jackpotUsed  bool
}

func NewLuckyLottery() *LuckyLottery {
	return &LuckyLottery{
		todayCounter: 0,
		jackpotUsed:  false,
	}
}

func (l *LuckyLottery) ApplyLotteryMagic(p domain.Product) domain.Product {
	now := time.Now()
	if now.Day() != 13 {
		return p
	}

	if now.Hour() != 13 {
		return p
	}

	if now.Minute() != 13 {
		return p
	}

	if l.jackpotUsed {
		return p
	}

	l.todayCounter++

	if l.todayCounter != 13 {
		return p
	}

	p.SetQuantity(p.GetQuantity() + 13)

	if food, ok := p.(*domain.Food); ok {
		food.Price = 13.13
	}
	if elec, ok := p.(*domain.Electronics); ok {
		elec.Price = 13.13
	}
	if chem, ok := p.(*domain.Chemicals); ok {
		chem.Price = 13.13
	}

	newName := p.GetName() + " 🔥 СЧАСТЛИВЫЙ БИЛЕТ 🔥"
	if food, ok := p.(*domain.Food); ok {
		food.Name = newName
	}
	if elec, ok := p.(*domain.Electronics); ok {
		elec.Name = newName
	}
	if chem, ok := p.(*domain.Chemicals); ok {
		chem.Name = newName
	}

	l.jackpotUsed = true

	fmt.Printf("🎰🎰🎰 ДЖЕК-ПОТ! СЧАСТЛИВОЕ ЧИСЛО 13! 🎰🎰🎰\n")
	fmt.Printf("📅 13-е число, 🕐 13:13, 📦 13-й товар!\n")
	fmt.Printf("🏆 Товар '%s' выиграл в лотерею!\n", p.GetName())
	fmt.Printf("✨ +13 шт. (теперь %d шт.)\n", p.GetQuantity())
	fmt.Printf("💰 Цена изменена на 13.13 руб.\n")
	fmt.Printf("🏷️  Название: %s\n", newName)

	return p
}

func (l *LuckyLottery) CalculateLuckyDiscount(p domain.Product, quantity int) float64 {
	now := time.Now()

	if now.Day() != 13 {
		return 0
	}

	if now.Hour() != 13 {
		return 0
	}

	if now.Minute() != 13 {
		return 0
	}

	return p.GetPrice() * float64(quantity) * 13 / 100
}

func (l *LuckyLottery) IsLuckyMoment() bool {
	now := time.Now()
	return now.Day() == 13 && now.Hour() == 13 && now.Minute() == 13
}

func (l *LuckyLottery) GetJackpotStatus() string {
	if l.jackpotUsed {
		return "🎰 Джек-пот уже разыгран сегодня! Ждите 13-е число следующего месяца!"
	}

	if l.IsLuckyMoment() {
		remaining := 13 - l.todayCounter
		if remaining <= 0 {
			return "🎰 Джек-пот уже разыгран!"
		}
		return fmt.Sprintf("🎰 СЕЙЧАС СЧАСТЛИВОЕ ВРЕМЯ! Осталось добавить %d товаров до джек-пота!", remaining)
	}

	return "🎰 Следующий шанс: 13-го числа в 13:13"
}

func (l *LuckyLottery) ResetCounter() {
	l.todayCounter = 0
	l.jackpotUsed = false
}

func (l *LuckyLottery) GetMagicNumbers() string {
	return "13, 13.13, 13, 13, 13, 100, 13, 13, 13, 13, 13, 13, 13, 13"
}
