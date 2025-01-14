package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
	"github.com/shopspring/decimal"
)

type NasdaqScraper struct {
	appContext *core.AppContext
}

func NewNasdaqScraper(appContext *core.AppContext) Scraper {
	return &NasdaqScraper{
		appContext: appContext,
	}
}

// SourceIdentifier implements Scraper.
func (n *NasdaqScraper) SourceIdentifier() string {
	return ScrapingSourceIdentifierNASDAQ
}

var nasdaqDefaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:129.0) Gecko/20100101 Firefox/129.0",
	"Accept":     "*/*",
}

// Scrape implements Scraper.
func (n *NasdaqScraper) Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error) {
	queries, err := db.OpenQueries(n.appContext.Config)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error opening db: %w", err)
	}

	symbolSource, err := queries.GetSymbolSource(ctx, dao.GetSymbolSourceParams{
		SymbolID: symbol.ID,
		SourceID: n.SourceIdentifier(),
	})
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error getting symbol source: %w", err)
	}

	req, err := http.NewRequest("GET", symbolSource.ScrapeUrl, nil)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error making http req: %w", err)
	}
	for k, v := range marketscreenerDefaultHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error HTTP GETting %v: %w", symbolSource.ScrapeUrl, err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error reading response body: %w", err)
	}
	bodyStr := string(bodyBytes)

	var nasdaqResponse nasdaqResponse
	err = json.Unmarshal(bodyBytes, &nasdaqResponse)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing response (%v): %w", bodyStr, err)
	}

	timestamp, err := n.parseTimestamp(nasdaqResponse)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing timestamp: %w", err)
	}

	price, err := n.parsePrice(nasdaqResponse)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing price: %w", err)
	}

	return ScrapeResult{
		Price:     price,
		Currency:  nasdaqResponse.Data.QdHeader.Currency,
		Timestamp: timestamp,
	}, nil
}

func (n *NasdaqScraper) parsePrice(nasdaqResponse nasdaqResponse) (decimal.Decimal, error) {
	inputPrice := nasdaqResponse.Data.QdHeader.PrimaryData.LastSalePrice
	currency := nasdaqResponse.Data.QdHeader.Currency

	formattedPrice := strings.TrimSpace(strings.ReplaceAll(inputPrice, currency, ""))

	price, err := decimal.NewFromString(formattedPrice)
	if err != nil {
		return decimal.Zero, fmt.Errorf("error parsing decimal from string %v (original value: %v): %w", formattedPrice, inputPrice, err)
	}
	return price, nil
}

func (n *NasdaqScraper) parseTimestamp(nasdaqResponse nasdaqResponse) (time.Time, error) {
	location, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return time.Time{}, fmt.Errorf("error loading Europe/Copenhagen location: %w", err)
	}
	inputTimestamp := nasdaqResponse.Data.QdHeader.PrimaryData.LastTradeTimestamp

	formattedInputTimestamp := inputTimestamp[:19]

	layout := "2006-01-02 15:04:05"

	parsedTimestamp, err := time.ParseInLocation(layout, formattedInputTimestamp, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing timestamp. Original timestamp=%v, formatted=%v: %w", inputTimestamp, formattedInputTimestamp, err)
	}

	return parsedTimestamp, nil
}

type nasdaqResponse struct {
	Data struct {
		QdHeader struct {
			Symbol      string `json:"symbol"`
			Currency    string `json:"currency"`
			PrimaryData struct {
				LastSalePrice      string `json:"lastSalePrice"`
				LastTradeTimestamp string `json:"lastTradeTimestamp"`
			} `json:"primaryData"`
		} `json:"qdHeader"`
	} `json:"data"`
}
