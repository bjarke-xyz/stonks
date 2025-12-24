package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
	"github.com/shopspring/decimal"
)

type NasdaqEuScraper struct {
	appContext *core.AppContext
}

func NewNasdaqEuScraper(appContext *core.AppContext) Scraper {
	return &NasdaqEuScraper{appContext: appContext}
}

// SourceIdentifier implements Scraper.
func (n *NasdaqEuScraper) SourceIdentifier() string {
	return ScrapingSourceIdentifierNASDAQ_EU
}

// Scrape implements Scraper.
func (n *NasdaqEuScraper) Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error) {

	// Initial request to get redirect with id in Location header
	initialURL := fmt.Sprintf("https://www.nasdaq.com/european-market-activity/funds/%s", symbol.Symbol)
	req, err := http.NewRequest("GET", initialURL, nil)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error making http request: %w", err)
	}
	for k, v := range nasdaqDefaultHeaders {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error HTTP GETting %v: %w", initialURL, err)
	}
	defer resp.Body.Close()

	loc := resp.Header.Get("Location")
	if loc == "" {
		return ScrapeResult{}, fmt.Errorf("no Location header in redirect response from %v (status %d)", initialURL, resp.StatusCode)
	}

	locURL, err := url.Parse(loc)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing Location header %v: %w", loc, err)
	}
	if !locURL.IsAbs() {
		base, _ := url.Parse("https://www.nasdaq.com")
		locURL = base.ResolveReference(locURL)
	}
	id := locURL.Query().Get("id")
	if id == "" {
		return ScrapeResult{}, fmt.Errorf("no id query param in Location header %v", locURL.String())
	}

	apiURL := fmt.Sprintf("https://api.nasdaq.com/api/nordic/listing?instrumentIds=%s&type=CARD_VIEW", id)
	req2, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error making api http request: %w", err)
	}
	for k, v := range nasdaqDefaultHeaders {
		req2.Header.Set(k, v)
	}
	req2.Header.Set("Accept", "application/json, text/plain, */*")

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error HTTP GETting api %v: %w", apiURL, err)
	}
	defer resp2.Body.Close()

	bodyBytes, err := io.ReadAll(resp2.Body)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error reading response body: %w", err)
	}
	bodyStr := string(bodyBytes)

	var apiResp nordicListingResponse
	if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing response (%v): %w", bodyStr, err)
	}

	if len(apiResp.Data.CardView) == 0 {
		return ScrapeResult{}, fmt.Errorf("no card view items in api response")
	}

	card := apiResp.Data.CardView[0]
	formattedPrice := strings.TrimSpace(strings.ReplaceAll(card.LastSalePrice, card.Currency, ""))
	price, err := decimal.NewFromString(formattedPrice)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing decimal from string %v (original value: %v): %w", formattedPrice, card.LastSalePrice, err)
	}

	// Parse timestamp, e.g. "Dec 24, 2025 10:09 CET"
	layout := "Jan 2, 2006 15:04 MST"
	parsedTimestamp, err := time.Parse(layout, apiResp.Data.DataAsOf)
	if err != nil {
		// fallback: try parse without timezone and assume Europe/Copenhagen
		location, lerr := time.LoadLocation("Europe/Copenhagen")
		if lerr == nil {
			parts := strings.Fields(apiResp.Data.DataAsOf)
			if len(parts) >= 1 {
				withoutZone := strings.Join(parts[:len(parts)-1], " ")
				parsedTimestamp, err = time.ParseInLocation("Jan 2, 2006 15:04", withoutZone, location)
			}
		}
		if err != nil {
			return ScrapeResult{}, fmt.Errorf("error parsing dataAsOf timestamp %v: %w", apiResp.Data.DataAsOf, err)
		}
	}

	return ScrapeResult{
		Price:     price,
		Currency:  card.Currency,
		Timestamp: parsedTimestamp,
	}, nil
}

type nordicListingResponse struct {
	Data struct {
		DataAsOf string `json:"dataAsOf"`
		CardView []struct {
			OrderbookId      string `json:"orderbookId"`
			AssetClass       string `json:"assetClass"`
			Symbol           string `json:"symbol"`
			MarketName       string `json:"marketName"`
			Currency         string `json:"currency"`
			LastSalePrice    string `json:"lastSalePrice"`
			NetChange        string `json:"netChange"`
			PercentageChange string `json:"percentageChange"`
			DeltaIndicator   string `json:"deltaIndicator"`
		} `json:"cardView"`
	} `json:"data"`
}
