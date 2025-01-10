package app

import (
	"github.com/bjarke-xyz/stonks/internal/config"
	"github.com/bjarke-xyz/stonks/internal/core"
)

func AppContext(cfg *config.Config) *core.AppContext {
	appContext := &core.AppContext{
		Config: cfg,
	}

	return appContext
}
