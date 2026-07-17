package web

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/bjarke-xyz/stonks/pkg"
	"github.com/bjarke-xyz/stonks/pkg/chart"
)

func (h *web) HandleGetQuote(w http.ResponseWriter, r *http.Request) {
	tickerSymbol := r.PathValue("symbol")
	durationInp := queryOr(r, "duration", "24h")
	parsedDuration, err := time.ParseDuration(durationInp)
	if err != nil {
		parsedDuration = 24 * time.Hour
	}
	endDate := pkg.EndOfDay(time.Now().UTC())
	startDate := endDate.Add(-parsedDuration)

	ctx := r.Context()

	quote, err := h.appContext.Deps.QuoteService.GetQuote(ctx, tickerSymbol, startDate, endDate)
	if err != nil {
		h.handleError(w, r, err)
		return
	}

	currency := r.URL.Query().Get("currency")
	if currency != "" {
		convertedQuote, err := h.appContext.Deps.CurrencyService.ConvertQuoteCurrency(ctx, quote, currency)
		if err != nil {
			h.handleError(w, r, fmt.Errorf("error converting currency: %w", err))
			return
		}
		quote = convertedQuote
	}

	chartSvg := ""
	includeChart := queryOr(r, "chart", "true")
	if includeChart != "false" {
		chartSvg = makeChart(quote)
	}

	model := views.QuoteViewModel{
		Base:     h.getBaseModel(r, quote.Symbol.Symbol+" | Quote"),
		Quote:    quote,
		ChartSvg: template.HTML(chartSvg),
	}

	switch r.URL.Query().Get("format") {
	case "table":
		err = views.Render(w, http.StatusOK, "quote_table.html", model)
	case "xml":
		err = writeXML(w, http.StatusOK, quote.ToSerializableQuote())
	default:
		err = views.Render(w, http.StatusOK, "quote.html", model)
	}
	if err != nil {
		slog.Error("rendering quote failed", "symbol", tickerSymbol, "error", err)
	}
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
