package parse

import (
	"strconv"
	"time"
)

type Amount struct {
	Unit     string
	Quantity string
}

type Price struct {
	Date  time.Time
	Unit  string
	Price Amount
}

func ParsePriceDB(buffer string) ([]Price, error) {
	prices := make([]Price, 0)
	lexer := lex("price db", buffer, lexPriceDB)
	for {
		next := lexer.nextItem()
		if next.typ == itemPriceSentinel {
			price, err := parsePrice(lexer)
			if err != nil {
				return nil, err
			}
			prices = append(prices, price)
		} else {
			break
		}
	}
	return prices, nil
}

func parsePrice(lexer *lexer) (Price, error) {
	yearItem := lexer.nextItem()
	year, err := strconv.Atoi(yearItem.value)
	if err != nil {
		return Price{}, err
	}
	monthItem := lexer.nextItem()
	month, err := strconv.Atoi(monthItem.value)
	if err != nil {
		return Price{}, err
	}
	dayOfMonthItem := lexer.nextItem()
	dayOfMonth, err := strconv.Atoi(dayOfMonthItem.value)
	if err != nil {
		return Price{}, err
	}
	unitItem := lexer.nextItem()
	amountUnitItem := lexer.nextItem()
	amountQuantityItem := lexer.nextItem()

	return Price{
		Date: time.Date(year, time.Month(month), dayOfMonth, 0, 0, 0, 0, time.UTC),
		Unit: unitItem.value,
		Price: Amount{
			Unit:     amountUnitItem.value,
			Quantity: amountQuantityItem.value,
		},
	}, nil
}
