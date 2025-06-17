package api

import (
	"context"
	"net/http"
	"time"

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

		if fireAndForget {
			// Do not use request context when using fireAndForget, timeout after 5 mins
			ctx := context.Background()
			ctx, _ = context.WithTimeout(ctx, time.Minute*5)
			go a.appContext.Deps.ScraperService.ScrapeSymbols(ctx)
		} else {
			ctx := c.Request.Context()
			a.appContext.Deps.ScraperService.ScrapeSymbols(ctx)
		}
		c.Status(http.StatusOK)
	}
}
