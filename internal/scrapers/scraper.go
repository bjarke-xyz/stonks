package scrapers

import (
	"context"
	"fmt"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
	"github.com/shopspring/decimal"
)

type ScrapeResult struct {
	Price     decimal.Decimal
	Currency  string
	Timestamp time.Time
}

type Scraper interface {
	Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error)
	SourceIdentifier() string
}

const (
	ScrapingSourceIdentifierBORSFRA = "BORSFRA"
)

func MakeScraper(scrapingSourceIdentifier string, appContext *core.AppContext) (Scraper, error) {
	switch scrapingSourceIdentifier {
	case ScrapingSourceIdentifierBORSFRA:
		return NewBorseFrankfurtScraper(appContext), nil
	default:
		return nil, fmt.Errorf("error making scraper, invalid scraping source identifier: %v", scrapingSourceIdentifier)
	}
}
