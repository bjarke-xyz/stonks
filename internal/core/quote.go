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
	Price     decimal.Decimal
	Currency  string
	Timestamp time.Time
}

type QuoteService interface {
	GetQuote(ctx context.Context, tickerSymbol string) (Quote, error)
}
