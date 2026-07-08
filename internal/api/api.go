package api

import (
	"context"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
)

type api struct {
	appContext *core.AppContext
}

func NewAPI(appContext *core.AppContext) *api {
	return &api{
		appContext: appContext,
	}
}

func (a *api) Route(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/job", a.RunJob())
}

func (a *api) RunJob() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != a.appContext.Config.JobKey {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		fireAndForget := r.URL.Query().Get("fireAndForget") == "true"

		if fireAndForget {
			// Detached from the request context so scraping outlives the response.
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
			go func() {
				defer cancel()
				a.appContext.Deps.ScraperService.ScrapeSymbols(ctx)
			}()
		} else {
			a.appContext.Deps.ScraperService.ScrapeSymbols(r.Context())
		}
		w.WriteHeader(http.StatusOK)
	}
}
