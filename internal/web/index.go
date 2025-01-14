package web

import (
	"net/http"

	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/gin-gonic/gin"
)

func (w *web) HandleGetIndex(c *gin.Context) {
	model := w.getBaseModel(c, "stonks")

	c.HTML(http.StatusOK, "", views.Index(model))
}
