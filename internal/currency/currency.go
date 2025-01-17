package currency

import (
	"context"
	"fmt"
	"strings"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/shopspring/decimal"
)

type CurrencyService struct {
	appContext *core.AppContext
}

func NewCurrencyService(appContext *core.AppContext) core.CurrencyService {
	return &CurrencyService{
		appContext: appContext,
	}
}

// ConvertCurrency implements core.CurrencyService.
func (c *CurrencyService) ConvertCurrency(ctx context.Context, amount decimal.Decimal, fromCurrency string, toCurrency string) (decimal.Decimal, error) {
	if strings.ToUpper(fromCurrency) == strings.ToUpper(toCurrency) {
		return amount, nil
	}
	exchangeRate, err := c.appContext.Deps.ExchangeRateService.GetExchangeRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error getting exchange rate: %w", err)
	}

	return amount.Mul(exchangeRate), nil
}

// ConvertQuoteCurrency implements core.CurrencyService.
func (c *CurrencyService) ConvertQuoteCurrency(ctx context.Context, quote core.Quote, toCurrency string) (core.Quote, error) {
	newPrice, err := c.ConvertCurrency(ctx, quote.Price.Price, quote.Price.Currency, toCurrency)
	if err != nil {
		return core.Quote{}, err
	}

	newOpeningPrice, err := c.ConvertCurrency(ctx, quote.Price.OpeningPrice, quote.Price.Currency, toCurrency)
	if err != nil {
		return core.Quote{}, err
	}

	newPreviousClosingPrice, err := c.ConvertCurrency(ctx, quote.Price.PreviousClosingPrice, quote.Price.Currency, toCurrency)
	if err != nil {
		return core.Quote{}, err
	}

	quote.Price.Price = newPrice
	quote.Price.OpeningPrice = newOpeningPrice
	quote.Price.PreviousClosingPrice = newPreviousClosingPrice
	quote.Price.Currency = toCurrency

	if len(quote.HistoricalPrices) > 0 {
		for i, price := range quote.HistoricalPrices {
			newHistoricalPrice, err := c.ConvertCurrency(ctx, price.Price, price.Currency, toCurrency)
			if err != nil {
				return core.Quote{}, err
			}
			quote.HistoricalPrices[i].Price = newHistoricalPrice
			quote.HistoricalPrices[i].Currency = toCurrency
		}
	}
	return quote, nil
}
