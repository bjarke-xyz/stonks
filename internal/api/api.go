package api

import (
	"context"
	"net/http"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/gin-gonic/gin"
)

type api struct {
	appContext *core.AppContext
}

func NewAPI(appContext *core.AppContext) *api {
	return &api{
		appContext: appContext,
	}
}

func (a *api) Route(r *gin.Engine) {
	apiGroup := r.Group("/api")
	apiGroup.POST("/job", a.RunJob())
}

func (a *api) RunJob() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") != a.appContext.Config.JobKey {
			c.AbortWithStatus(401)
			return
		}

		fireAndForget := c.Query("fireAndForget") == "true"
		//using context.Background to not cancel, if this method times out
		ctx := context.Background()

		if fireAndForget {
			go a.appContext.Deps.ScraperService.ScrapeSymbols(ctx)
		} else {
			a.appContext.Deps.ScraperService.ScrapeSymbols(ctx)
		}
		c.Status(http.StatusOK)
	}
}
