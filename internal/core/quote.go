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

func (q Quote) ToSerializableQuote() SerializableQuote {
	return SerializableQuote{
		Quote: q,
		Price: q.Price.ToSerializablePrice(),
	}
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
func (p Price) ToSerializablePrice() SerializablePrice {
	return SerializablePrice{
		Price:                 p,
		PriceChangeAbsolute:   p.PriceChangeAbsolute(),
		PriceChangePercentage: p.PriceChangePercentage(),
	}
}

type SerializablePrice struct {
	Price
	PriceChangeAbsolute   decimal.Decimal
	PriceChangePercentage decimal.Decimal
}

type SerializableQuote struct {
	Quote
	Price SerializablePrice
}

type QuoteService interface {
	GetQuote(ctx context.Context, tickerSymbol string, startDate time.Time, endDate time.Time) (Quote, error)
	ClearCache(ctx context.Context, tickerSymbol string) error
}
