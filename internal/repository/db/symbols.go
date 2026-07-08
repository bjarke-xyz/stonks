package db

import (
	"context"
	"fmt"
	"time"
)

func (r *Repo) SymbolByID(ctx context.Context, id int64) (Symbol, error) {
	var s Symbol
	err := r.db.QueryRowContext(ctx,
		`SELECT id, symbol, name, isin FROM symbols WHERE id = ?`, id,
	).Scan(&s.ID, &s.Symbol, &s.Name, &s.Isin)
	return s, err
}

func (r *Repo) SymbolByTicker(ctx context.Context, ticker string) (Symbol, error) {
	var s Symbol
	err := r.db.QueryRowContext(ctx,
		`SELECT id, symbol, name, isin FROM symbols WHERE symbol = ?`, ticker,
	).Scan(&s.ID, &s.Symbol, &s.Name, &s.Isin)
	return s, err
}

func (r *Repo) ScrapingSourceByID(ctx context.Context, id string) (ScrapingSource, error) {
	var s ScrapingSource
	err := r.db.QueryRowContext(ctx,
		`SELECT id, name, base_url, additional_info FROM scraping_sources WHERE id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.BaseUrl, &s.AdditionalInfo)
	return s, err
}

func (r *Repo) SymbolSource(ctx context.Context, symbolID int64, sourceID string) (SymbolSource, error) {
	var s SymbolSource
	err := r.db.QueryRowContext(ctx,
		`SELECT id, symbol_id, source_id, scrape_url, active, last_scraped
		 FROM symbol_sources
		 WHERE symbol_id = ? AND source_id = ?`, symbolID, sourceID,
	).Scan(&s.ID, &s.SymbolID, &s.SourceID, &s.ScrapeUrl, &s.Active, &s.LastScraped)
	return s, err
}

// SourcesNotScrapedRecently returns active symbol sources that have not been
// scraped within the last 10 minutes.
func (r *Repo) SourcesNotScrapedRecently(ctx context.Context) ([]SymbolSource, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, symbol_id, source_id, scrape_url, active, last_scraped
		 FROM symbol_sources
		 WHERE active = TRUE
		   AND (last_scraped IS NULL OR DATETIME(last_scraped, '+10 minutes') <= DATETIME('now'))`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []SymbolSource
	for rows.Next() {
		var s SymbolSource
		if err := rows.Scan(&s.ID, &s.SymbolID, &s.SourceID, &s.ScrapeUrl, &s.Active, &s.LastScraped); err != nil {
			return nil, fmt.Errorf("error scanning symbol source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (r *Repo) UpdateLastScraped(ctx context.Context, symbolID int64, sourceID string, lastScraped time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE symbol_sources SET last_scraped = ? WHERE symbol_id = ? AND source_id = ?`,
		lastScraped, symbolID, sourceID)
	return err
}
