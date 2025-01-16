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


-- name: GetQuote :one
-- Get the latest price and today's price change (absolute and percentage) for a given symbol ID
WITH latest_price AS (
    SELECT id, price, currency, timestamp
    FROM prices p
    WHERE p.symbol_id = ?1
    ORDER BY timestamp DESC
    LIMIT 1
),
opening_price AS (
    SELECT price AS opening_price
    FROM prices p
    WHERE p.symbol_id = ?1
      AND DATE(timestamp) = DATE('now')
    ORDER BY timestamp ASC
    LIMIT 1
),
previous_closing_price AS (
    SELECT price AS closing_price
    FROM prices
    WHERE symbol_id = ?1
      AND DATE(timestamp) = DATE('now', '-1 day')
    ORDER BY timestamp DESC
    LIMIT 1
)
SELECT 
    lp.id AS id,
    lp.price AS latest_price,
    lp.currency AS currency,
    lp.timestamp AS timestamp,
    COALESCE(op.opening_price, 0.0) AS opening_price,
    CAST(lp.price - COALESCE(op.opening_price, 0.0) AS NUMERIC) AS price_change_absolute,
    CASE 
        WHEN COALESCE(op.opening_price, 0.0) > 0 THEN 
            CAST(((lp.price - COALESCE(op.opening_price, 0.0)) * 100.0) / COALESCE(op.opening_price, 0.0) AS NUMERIC)
        ELSE 
            CAST(0.0 AS NUMERIC)
    END AS price_change_percentage,
    COALESCE(pc.closing_price, 0.0) AS previous_closing_price
FROM latest_price lp
LEFT JOIN opening_price op ON 1=1
LEFT JOIN previous_closing_price pc ON 1=1;
