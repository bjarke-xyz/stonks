package app

import (
	"github.com/bjarke-xyz/stonks/internal/config"
	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/currency"
	"github.com/bjarke-xyz/stonks/internal/quote"
	"github.com/bjarke-xyz/stonks/internal/repository"
	"github.com/bjarke-xyz/stonks/internal/scrapers"
)

func AppContext(cfg *config.Config) *core.AppContext {
	appContext := &core.AppContext{
		Config: cfg,
	}

	deps := &core.AppDeps{
		ScraperService:      scrapers.NewScraperService(appContext),
		Cache:               repository.NewCacheService(repository.NewCacheRepo(cfg, true)),
		QuoteService:        quote.NewQuoteService(appContext),
		ExchangeRateService: currency.NewExchangeRateService(appContext),
		CurrencyService:     currency.NewCurrencyService(appContext),
	}
	appContext.Deps = deps

	return appContext
}
