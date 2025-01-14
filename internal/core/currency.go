package core

import (
	"context"

	"github.com/shopspring/decimal"
)

type CurrencyService interface {
	ConvertCurrency(ctx context.Context, amount decimal.Decimal, fromCurrency, toCurrency string) (decimal.Decimal, error)
	ConvertQuoteCurrency(ctx context.Context, quote Quote, toCurrency string) (Quote, error)
}

type ExchangeRateService interface {
	GetExchangeRate(ctx context.Context, fromCurrency, toCurrency string) (decimal.Decimal, error)
}
