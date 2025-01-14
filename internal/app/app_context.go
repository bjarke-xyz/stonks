package app

import (
	"github.com/bjarke-xyz/stonks/internal/config"
	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/scrapers"
)

func AppContext(cfg *config.Config) *core.AppContext {
	appContext := &core.AppContext{
		Config: cfg,
	}

	deps := &core.AppDeps{
		ScraperService: scrapers.NewScraperService(appContext),
	}
	appContext.Deps = deps

	return appContext
}
