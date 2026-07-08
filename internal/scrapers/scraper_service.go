package scrapers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/samber/lo"
)

type ScraperService struct {
	appContext *core.AppContext
}

func NewScraperService(appContext *core.AppContext) core.ScraperService {
	return &ScraperService{appContext: appContext}
}

func (s *ScraperService) ScrapeSymbols(ctx context.Context) {
	log.Printf("scraping symbols...")
	err := s.internalScrapeSymbols(ctx)
	if err != nil {
		log.Printf("error scraping symbols: %v", err)
	}
}

func (s *ScraperService) internalScrapeSymbols(ctx context.Context) error {
	repo, err := db.OpenRepo(s.appContext.Config)
	if err != nil {
		return fmt.Errorf("error opening db")
	}

	scrapeSources, err := repo.SourcesNotScrapedRecently(ctx)
	if err != nil {
		return fmt.Errorf("error getting sources not scraped recently: %w", err)
	}

	log.Printf("found %v scrape sources that has not been scraped recently", len(scrapeSources))

	groupedScrapeSources := lo.GroupBy(scrapeSources, func(ss db.SymbolSource) string {
		return ss.SourceID
	})

	for sourceIdentifier, scrapeSources := range groupedScrapeSources {
		symbolIds := lo.Map(scrapeSources, func(ss db.SymbolSource, _ int) int64 { return ss.SymbolID })
		err := s.scrapeSymbolsForSourceIdentifier(ctx, repo, sourceIdentifier, symbolIds)
		if err != nil {
			return fmt.Errorf("error scraping symbols for source identifier %v: %w", sourceIdentifier, err)
		}
	}
	return nil
}

func (s *ScraperService) scrapeSymbolsForSourceIdentifier(ctx context.Context, repo *db.Repo, sourceIdentifier string, symbolIds []int64) error {
	scraper, err := MakeScraper(sourceIdentifier, s.appContext)
	if err != nil {
		return fmt.Errorf("error making scraping: %w", err)
	}

	for _, symbolId := range symbolIds {
		err = s.scrapeAndStoreSymbol(ctx, repo, scraper, symbolId)
		if err != nil {
			return fmt.Errorf("error scraping and storing symbol: %w", err)
		}
	}
	return nil
}

func (s *ScraperService) scrapeAndStoreSymbol(ctx context.Context, repo *db.Repo, scraper Scraper, symbolId int64) error {
	symbol, err := repo.SymbolByID(ctx, symbolId)
	if err != nil {
		return fmt.Errorf("error getting symbol for id %v: %w", symbolId, err)
	}

	scrapeResult, err := scraper.Scrape(ctx, symbol)
	if err != nil {
		return fmt.Errorf("error scraping symbol %+v: %w", symbol, err)
	}

	err = repo.InsertPrice(ctx, symbol.ID, scrapeResult.Price, scrapeResult.Currency, scrapeResult.Timestamp)
	if err != nil {
		return fmt.Errorf("error inserting price for symbol %+v: %w", symbol, err)
	}
	log.Printf("scraped symbol %v: %v (%v) at %v", symbol.Symbol, scrapeResult.Price, scrapeResult.Currency, scrapeResult.Timestamp)

	err = repo.UpdateLastScraped(ctx, symbol.ID, scraper.SourceIdentifier(), time.Now().UTC())
	if err != nil {
		// not important if this fails, just log it, dont return the err
		log.Printf("failed to update last scraped timestamp for symbol %v, source %v: %v", symbol.ID, scraper.SourceIdentifier(), err)
	}
	s.appContext.Deps.QuoteService.ClearCache(ctx, symbol.Symbol)

	return nil
}
