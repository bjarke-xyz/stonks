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

type YFinanceAPIScraper struct {
	appContext *core.AppContext
}

func NewYFinanceAPIScraper(appContext *core.AppContext) Scraper {
	return &YFinanceAPIScraper{appContext: appContext}
}

// SourceIdentifier implements Scraper.
func (y *YFinanceAPIScraper) SourceIdentifier() string {
	return ScrapingSourceIdentifierYFINANCEAPI
}

// Scrape implements Scraper.
func (y *YFinanceAPIScraper) Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error) {
	queries, err := db.OpenQueries(y.appContext.Config)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error opening db: %w", err)
	}

	scrapingSource, err := queries.GetScrapingSourceByID(ctx, y.SourceIdentifier())
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error getting scraping source: %w", err)
	}

	url := fmt.Sprintf("%s/ticker/%s", scrapingSource.BaseUrl, symbol.Symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error creating request: %w", err)
	}

	if y.appContext.Config.YFinanceAPIAuthKey != "" {
		req.Header.Set("Authorization", y.appContext.Config.YFinanceAPIAuthKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error HTTP GETting %v: %w", url, err)
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error reading response body: %w", err)
	}
	bodyStr := string(bodyBytes)

	parsedResponse := yfinanceAPIResponse{}
	err = json.Unmarshal(bodyBytes, &parsedResponse)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing response body (%v): %w", bodyStr, err)
	}

	return ScrapeResult{
		Price:     parsedResponse.LatestPrice,
		Currency:  parsedResponse.Currency,
		Timestamp: parsedResponse.Timestamp,
	}, nil
}

type yfinanceAPIResponse struct {
	Symbol      string          `json:"symbol"`
	LatestPrice decimal.Decimal `json:"latestPrice"`
	Currency    string          `json:"currency"`
	Timestamp   time.Time       `json:"timestamp"`
}
