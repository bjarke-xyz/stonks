package web

import (
	"net/http"

	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/gin-gonic/gin"
)

func (w *web) HandleGetIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", views.IndexViewModel{
		Base: w.getBaseModel(c, "stonks"),
	})
}
