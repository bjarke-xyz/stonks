package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/bjarke-xyz/stonks/pkg"
	"github.com/bjarke-xyz/stonks/pkg/chart"
	"github.com/gin-gonic/gin"
)

func (w *web) HandleGetQuote(c *gin.Context) {
	tickerSymbol := c.Param("symbol")
	durationInp := c.DefaultQuery("duration", "24h")
	parsedDuration, err := time.ParseDuration(durationInp)
	if err != nil {
		parsedDuration = 24 * time.Hour
	}
	endDate := pkg.EndOfDay(time.Now().UTC())
	startDate := endDate.Add(-parsedDuration)

	ctx := c.Request.Context()

	quote, err := w.appContext.Deps.QuoteService.GetQuote(ctx, tickerSymbol, startDate, endDate)
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

	chartSvg := ""
	includeChart := c.DefaultQuery("chart", "true")
	if includeChart != "false" {
		chartSvg = makeChart(quote)
	}

	model := views.QuoteViewModel{
		Base:     w.getBaseModel(c, quote.Symbol.Symbol+" | Quote"),
		Quote:    quote,
		ChartSvg: chartSvg,
	}

	format := c.Query("format")
	if format == "table" {
		c.HTML(http.StatusOK, "", views.QuoteTable(model))
		return
	}
	if format == "xml" {
		serializableQuote := quote.ToSerializableQuote()
		c.XML(http.StatusOK, serializableQuote)
		return
	}

	c.HTML(http.StatusOK, "", views.Quote(model))
}

func makeChart(quote core.Quote) string {
	if len(quote.HistoricalPrices) == 0 {
		return ""
	}

	timestampLayout := "15:04"
	firstDate := quote.HistoricalPrices[0].Timestamp
	lastDate := quote.HistoricalPrices[len(quote.HistoricalPrices)-1].Timestamp
	if lastDate.Sub(firstDate).Hours() > 24 {
		timestampLayout = "01-02T15:04"
	}

	prices := make([]float64, len(quote.HistoricalPrices))
	timestamps := make([]string, len(quote.HistoricalPrices))
	for i, histPrice := range quote.HistoricalPrices {
		prices[i] = histPrice.Price.InexactFloat64()
		timestamps[i] = histPrice.Timestamp.Format(timestampLayout)
	}

	return chart.LineChart{
		Title:   quote.Price.Currency,
		Legend:  quote.Symbol.Symbol,
		XLabels: timestamps,
		Values:  prices,
	}.SVG()
}
