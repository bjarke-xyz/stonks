package scrapers

import (
	"context"
	"fmt"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
)

type ScraperJob struct {
	appContext *core.AppContext
}

func NewScraperJob(appContext *core.AppContext) *ScraperJob {
	return &ScraperJob{
		appContext: appContext,
	}
}

func (j *ScraperJob) RunJob(ctx context.Context) error {
	queries, err := db.OpenQueries(j.appContext.Config)
	if err != nil {
		return fmt.Errorf("error opening db: %w", err)
	}

	symbolSources, err := queries.GetSourcesNotScrapedRecently(ctx)
	if err != nil {
		return fmt.Errorf("error getting sources not scraped recently: %w", err)
	}

	for _, symbolSource := range symbolSources {
		scraper, err := MakeScraper(symbolSource.SourceID, j.appContext)
		if err != nil {
			return fmt.Errorf("error making scraper for %v: %w", symbolSource.SourceID, err)
		}

		symbol, err := queries.GetSymbolByID(ctx, symbolSource.SymbolID)
		if err != nil {
			return fmt.Errorf("error getting symbol for %v: %w", symbolSource.SymbolID, err)
		}

		scrapeResult, err := scraper.Scrape(ctx, symbol)
		if err != nil {
			return fmt.Errorf("error scraping symbol %v: %w", symbol.Symbol, err)
		}

		err = queries.InsertPrice(ctx, dao.InsertPriceParams{
			SymbolID:  symbol.ID,
			Price:     scrapeResult.Price,
			Currency:  scrapeResult.Currency,
			Timestamp: scrapeResult.Timestamp,
		})
		if err != nil {
			return fmt.Errorf("error inserting price for symbol %v: %w", symbol.Symbol, err)
		}
	}

	return nil
}
