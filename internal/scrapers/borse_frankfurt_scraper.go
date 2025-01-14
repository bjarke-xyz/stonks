package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
	"github.com/shopspring/decimal"
)

type BorseFrankfurtScraper struct {
	appContext *core.AppContext
}

func NewBorseFrankfurtScraper(appContext *core.AppContext) Scraper {
	return &BorseFrankfurtScraper{appContext: appContext}
}

// SourceIdentifier implements Scraper.
func (b *BorseFrankfurtScraper) SourceIdentifier() string {
	return ScrapingSourceIdentifierBORSFRA
}

// Scrape implements Scraper.
func (b *BorseFrankfurtScraper) Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error) {
	queries, err := db.OpenQueries(b.appContext.Config)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error opening db: %w", err)
	}

	symbolSource, err := queries.GetSymbolSource(ctx, dao.GetSymbolSourceParams{
		SymbolID: symbol.ID,
		SourceID: b.SourceIdentifier(),
	})
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error getting symbol source: %w", err)
	}

	resp, err := http.Get(symbolSource.ScrapeUrl)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error HTTP GETting %v: %w", symbolSource.ScrapeUrl, err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error reading response body: %w", err)
	}
	bodyStr := string(bodyBytes)

	parsedResponse := borseFrankfurtResponse{}
	err = json.Unmarshal(bodyBytes, &parsedResponse)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing response body (%v): %w", bodyStr, err)
	}

	return ScrapeResult{
		Price:     parsedResponse.LastPrice,
		Currency:  parsedResponse.Currency.OriginalValue,
		Timestamp: parsedResponse.TimestampLastPrice,
	}, nil
}

type borseFrankfurtResponse struct {
	LastPrice          decimal.Decimal                `json:"lastPrice"`
	TimestampLastPrice time.Time                      `json:"timestampLastPrice"`
	Currency           borseFrankfurtCurrencyResponse `json:"currency"`
}
type borseFrankfurtCurrencyResponse struct {
	OriginalValue string `json:"originalValue"`
}
