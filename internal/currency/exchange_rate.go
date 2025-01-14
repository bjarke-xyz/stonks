package currency

import (
	"context"
	"fmt"
	"strings"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/shopspring/decimal"
)

type ExchangeRateService struct {
	appContext *core.AppContext
}

func NewExchangeRateService(appContext *core.AppContext) core.ExchangeRateService {
	return &ExchangeRateService{
		appContext: appContext,
	}
}

// GetExchangeRate implements core.ExchangeRateService.
func (e *ExchangeRateService) GetExchangeRate(ctx context.Context, fromCurrency string, toCurrency string) (decimal.Decimal, error) {
	// TODO: Get real exchange rates
	rates := map[string]map[string]decimal.Decimal{
		"EUR": {
			"DKK": decimal.NewFromFloat(7.46),
		},
		"DKK": {
			"EUR": decimal.NewFromFloat(0.134),
		},
	}

	if rate, ok := rates[strings.ToUpper(fromCurrency)][strings.ToUpper(toCurrency)]; ok {
		return rate, nil
	}
	return decimal.Zero, fmt.Errorf("exchange rate not found for %s to %s", fromCurrency, toCurrency)
}
