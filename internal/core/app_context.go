package core

import (
	"github.com/bjarke-xyz/stonks/internal/config"
)

type AppContext struct {
	Config *config.Config
	Deps   *AppDeps
}

type AppDeps struct {
	ScraperService ScraperService
}
