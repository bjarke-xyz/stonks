-- name: GetSymbolByID :one
-- Get a symbol by its ID
SELECT *
FROM symbols
WHERE id = ?;

-- name: GetSymbolByTicker :one
-- Get a symbol by its ticker
SELECT *
FROM symbols
WHERE symbol = ?;

-- name: InsertSymbol :one
-- Insert a new symbol
INSERT INTO symbols (symbol, name)
VALUES (?, ?)
RETURNING id, symbol, name;

-- name: InsertPrice :exec
-- Insert or update a price entry on conflict
INSERT INTO prices (symbol_id, price, currency, timestamp)
VALUES (?, ?, ?, ?)
ON CONFLICT (symbol_id, timestamp, currency) DO UPDATE
SET price = EXCLUDED.price, 
    timestamp = EXCLUDED.timestamp;

-- name: UpdateLastScraped :exec
-- Update the last scraped timestamp for a specific symbol and source
UPDATE symbol_sources
SET last_scraped = ?
WHERE symbol_id = ? AND source_id = ?;

-- name: GetAllScrapingSources :many
-- Get all scraping sources
SELECT *
FROM scraping_sources;

-- name: InsertScrapingSource :exec
-- Insert a new scraping source
INSERT INTO scraping_sources (id, name, base_url, additional_info)
VALUES (?, ?, ?, ?);

-- name: InsertSymbolSource :exec
-- Add a new symbol-source mapping
INSERT INTO symbol_sources (symbol_id, source_id, scrape_url, active)
VALUES (?, ?, ?, ?);

-- name: DeactivateSymbolSource :exec
-- Deactivate a scraping source for a specific symbol
UPDATE symbol_sources
SET active = FALSE
WHERE symbol_id = ? AND source_id = ?;

-- name: GetLatestPriceForSymbol :one
-- Get the latest price for a given symbol
SELECT *
FROM prices p
WHERE p.symbol_id = ?
ORDER BY p.timestamp DESC
LIMIT 1;

-- name: GetSymbolsWithNoPrices :many
-- Get all symbols that have no prices recorded
SELECT *
FROM symbols s
LEFT JOIN prices p ON s.id = p.symbol_id
WHERE p.id IS NULL;

-- name: GetScrapingSourceByID :one
-- Get a scraping source by its ID
SELECT *
FROM scraping_sources
WHERE id = ?;

-- name: GetSymbolSources :many
-- Get all sources mapped to a symbol
SELECT *
FROM symbol_sources ss
WHERE ss.symbol_id = ?;

-- name: GetSymbolSource :one
-- Get symbol source by symbol ID and scraping source id
SELECT * 
FROM symbol_sources ss
WHERE ss.symbol_id = ? AND ss.source_id = ?;


-- name: GetSourcesNotScrapedRecently :many
-- Get scraping sources that have not been scraped in the last 10 minutes
SELECT *
FROM symbol_sources ss
WHERE ss.active = TRUE 
  AND (
    ss.last_scraped IS NULL OR
    DATETIME(ss.last_scraped, '+10 minutes') <= DATETIME('now')
  );
