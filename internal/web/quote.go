package web

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/web/views"
	"github.com/bjarke-xyz/stonks/pkg"
	"github.com/gin-gonic/gin"
	charts "github.com/vicanso/go-charts/v2"
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

	chartPngBase64 := ""
	includeChart := c.DefaultQuery("chart", "true")
	if includeChart != "false" {
		chartPngBytes, err := w.makeChart(quote)
		if err != nil {
			log.Printf("error making chart for symbol %v: %v", quote.Symbol.Symbol, err)
		}
		chartPngBase64 = base64.StdEncoding.EncodeToString(chartPngBytes)
	}

	model := views.QuoteViewModel{
		Base:           w.getBaseModel(c, quote.Symbol.Symbol+" | Quote"),
		Quote:          quote,
		ChartPngBase64: chartPngBase64,
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

func (w *web) makeChart(quote core.Quote) ([]byte, error) {
	if len(quote.HistoricalPrices) == 0 {
		return []byte{}, nil
	}
	values := [][]float64{}
	xAxisValues := make([]string, len(quote.HistoricalPrices))

	timestampLayout := "15:04"
	firstDate := quote.HistoricalPrices[0].Timestamp
	lastDate := quote.HistoricalPrices[len(quote.HistoricalPrices)-1].Timestamp
	if lastDate.Sub(firstDate).Hours() > 24 {
		timestampLayout = "01-02T15:04"
	}

	priceValues := make([]float64, len(quote.HistoricalPrices))
	for i, histPrice := range quote.HistoricalPrices {
		priceValues[i] = histPrice.Price.InexactFloat64()
		xAxisValues[i] = histPrice.Timestamp.Format(timestampLayout)
	}
	values = append(values, priceValues)

	p, err := charts.LineRender(
		values,
		charts.TitleTextOptionFunc(quote.Price.Currency),
		charts.XAxisDataOptionFunc(xAxisValues),
		charts.LegendLabelsOptionFunc([]string{quote.Symbol.Symbol}, charts.PositionCenter),
		charts.WidthOptionFunc(2048),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to make chart: %w", err)
	}

	buf, err := p.Bytes()
	if err != nil {
		return nil, fmt.Errorf("failed to render chart: %w", err)
	}

	return buf, nil
}
