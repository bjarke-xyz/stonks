package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
	"github.com/shopspring/decimal"
)

var marketscreenerDefaultHeaders = map[string]string{
	"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:129.0) Gecko/20100101 Firefox/129.0",
	"Accept":     "*/*",
}

type MarketscreenerScraper struct {
	appContext *core.AppContext
}

func NewMarketscreenerScraper(appContext *core.AppContext) Scraper {
	return &MarketscreenerScraper{
		appContext: appContext,
	}
}

// SourceIdentifier implements Scraper.
func (m *MarketscreenerScraper) SourceIdentifier() string {
	return ScrapingSourceIdentifierMARKETSCREENER
}

// Scrape implements Scraper.
func (m *MarketscreenerScraper) Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error) {
	queries, err := db.OpenQueries(m.appContext.Config)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error opening db: %w", err)
	}

	symbolSource, err := queries.GetSymbolSource(ctx, dao.GetSymbolSourceParams{
		SymbolID: symbol.ID,
		SourceID: m.SourceIdentifier(),
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

	// Fetch the HTML content
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ScrapeResult{}, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Parse the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find the script tag with type application/ld+json
	var financialProductJSON string
	doc.Find("script[type='application/ld+json']").EachWithBreak(func(i int, s *goquery.Selection) bool {
		content := s.Text()
		var temp map[string]interface{}
		if err := json.Unmarshal([]byte(content), &temp); err == nil {
			if temp["@type"] == "FinancialProduct" {
				financialProductJSON = content
				return false // Break out of the loop
			}
		}
		return true // Continue the loop
	})

	if financialProductJSON == "" {
		return ScrapeResult{}, fmt.Errorf("no FinancialProduct schema found")
	}

	// Parse the FinancialProduct JSON
	type FinancialProduct struct {
		Offers struct {
			Price         string `json:"price"`
			PriceCurrency string `json:"priceCurrency"`
			ValidFrom     string `json:"validFrom"`
		} `json:"offers"`
	}

	var product FinancialProduct
	if err := json.Unmarshal([]byte(financialProductJSON), &product); err != nil {
		return ScrapeResult{}, fmt.Errorf("failed to parse FinancialProduct JSON: %w", err)
	}

	// Parse the fields into ScrapeResult
	price, err := decimal.NewFromString(product.Offers.Price)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("invalid price format: %w", err)
	}

	timestamp, err := m.parseTimestamp(product.Offers.ValidFrom)
	if err != nil {
		log.Printf("failed to parse marketscreener timestamp (%v): %v", product.Offers.ValidFrom, err)
		timestamp = time.Now() // Fallback to `now`
	}

	result := ScrapeResult{
		Price:     price,
		Currency:  product.Offers.PriceCurrency,
		Timestamp: timestamp,
	}
	return result, nil
}

func (m *MarketscreenerScraper) parseTimestamp(inputTimestamp string) (time.Time, error) {
	location, err := time.LoadLocation("Europe/Copenhagen")
	if err != nil {
		return time.Time{}, fmt.Errorf("error loading Europe/Copenhagen location: %w", err)
	}

	splitStr := "CET"
	if strings.Contains(inputTimestamp, "CEST") {
		splitStr = "CEST"
	}
	inputTimestampParts := strings.Split(inputTimestamp, splitStr)

	formattedInputTimestamp := strings.Join(inputTimestampParts, "T")

	layout := "2006-01-02T15:04:05"

	parsedTimestamp, err := time.ParseInLocation(layout, formattedInputTimestamp, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing timestamp. Original timestamp=%v, formatted=%v: %w", inputTimestamp, formattedInputTimestamp, err)
	}

	return parsedTimestamp, nil
}
