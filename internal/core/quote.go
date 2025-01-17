package core

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type Quote struct {
	Symbol           Symbol
	Price            Price
	HistoricalPrices []SimplePrice
}

type Symbol struct {
	Symbol string
	Name   string
}

type Price struct {
	Price                decimal.Decimal
	Currency             string
	Timestamp            time.Time
	OpeningPrice         decimal.Decimal
	PreviousClosingPrice decimal.Decimal
}

type SimplePrice struct {
	Price     decimal.Decimal
	Currency  string
	Timestamp time.Time
}

func (p Price) PriceChangeAbsolute() decimal.Decimal {
	// assuming a price will never go to 0...
	if p.PreviousClosingPrice.Equal(decimal.NewFromInt(0)) {
		return decimal.NewFromInt(0)
	}
	return p.Price.Sub(p.PreviousClosingPrice)
}

func (p Price) PriceChangePercentage() decimal.Decimal {
	// Check if OpeningPrice is greater than 0
	if p.PreviousClosingPrice.GreaterThan(decimal.NewFromInt(0)) {
		// Calculate price change percentage: ((Price - OpeningPrice) * 100) / OpeningPrice
		priceChange := p.Price.Sub(p.PreviousClosingPrice).Mul(decimal.NewFromInt(100)).Div(p.PreviousClosingPrice)
		return priceChange
	}

	return decimal.NewFromInt(0)
}

type QuoteService interface {
	GetQuote(ctx context.Context, tickerSymbol string, startDate time.Time, endDate time.Time) (Quote, error)
	ClearCache(ctx context.Context, tickerSymbol string) error
}
