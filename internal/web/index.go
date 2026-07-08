package web

import (
	"log"
	"net/http"

	"github.com/bjarke-xyz/stonks/internal/web/views"
)

func (h *web) HandleGetIndex(w http.ResponseWriter, r *http.Request) {
	err := views.Render(w, http.StatusOK, "index.html", views.IndexViewModel{
		Base: h.getBaseModel(r, "stonks"),
	})
	if err != nil {
		log.Printf("error rendering index: %v", err)
	}
}
