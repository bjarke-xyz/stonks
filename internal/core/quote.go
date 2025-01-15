package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type Quote struct {
	Symbol Symbol
	Price  Price
}

type Symbol struct {
	Symbol string
	Name   string
}

type Price struct {
	Price        decimal.Decimal
	Currency     string
	Timestamp    time.Time
	OpeningPrice decimal.Decimal
}

func (p Price) PriceChangeAbsolute() decimal.Decimal {
	// assuming a price will never go to 0...
	if p.OpeningPrice.Equal(decimal.NewFromInt(0)) {
		return decimal.NewFromInt(0)
	}
	return p.Price.Sub(p.OpeningPrice)
}

func (p Price) PriceChangePercentage() decimal.Decimal {
	// Check if OpeningPrice is greater than 0
	if p.OpeningPrice.GreaterThan(decimal.NewFromInt(0)) {
		// Calculate price change percentage: ((Price - OpeningPrice) * 100) / OpeningPrice
		priceChange := p.Price.Sub(p.OpeningPrice).Mul(decimal.NewFromInt(100)).Div(p.OpeningPrice)
		return priceChange
	}

	return decimal.NewFromInt(0)
}

type QuoteService interface {
	GetQuote(ctx context.Context, tickerSymbol string) (Quote, error)
}
