package web

import (
	"fmt"
	"net/http"

	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/gin-gonic/gin"
)

func (w *web) HandleGetQuote(c *gin.Context) {
	tickerSymbol := c.Param("symbol")

	ctx := c.Request.Context()

	quote, err := w.appContext.Deps.QuoteService.GetQuote(ctx, tickerSymbol)
	if err != nil {
		w.handleError(c, err)
		return
	}

	currency := c.Query("currency")
	if currency != "" {
		convertedQuote, err := w.appContext.Deps.CurrencyService.ConvertQuoteCurrency(ctx, quote, currency)
		if err != nil {
			w.handleError(c, fmt.Errorf("error converting currency: %w", err))
			return
		}
		quote = convertedQuote
	}

	model := views.QuoteViewModel{
		Base:  w.getBaseModel(c, quote.Symbol.Symbol+" | Quote"),
		Quote: quote,
	}

	format := c.Query("format")
	if format == "table" {
		c.HTML(http.StatusOK, "", views.QuoteTable(model))
		return
	}

	c.HTML(http.StatusOK, "", views.Quote(model))
}
