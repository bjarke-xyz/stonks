-- Repairs timestamps written as time.Time.String() ("2026-07-09 18:51:56 +0200 +0200")
-- while the DSN lacked _time_format=sqlite. Such values parse back neither in the
-- driver nor in SQLite's date functions, which broke price lookups and silently
-- stopped scraping. Rows written before this all use the canonical format, matched
-- here by the space that only String() puts before the offset.

-- +goose Up

-- Drop the rows whose canonical form already exists, then convert the rest.
DELETE FROM prices
WHERE timestamp LIKE '____-__-__ __:__:__ %'
  AND EXISTS (
      SELECT 1 FROM prices canonical
      WHERE canonical.symbol_id = prices.symbol_id
        AND canonical.currency = prices.currency
        AND canonical.timestamp = substr(prices.timestamp, 1, 19)
                               || substr(prices.timestamp, 21, 3) || ':'
                               || substr(prices.timestamp, 24, 2)
  );

UPDATE prices
SET timestamp = substr(timestamp, 1, 19)
             || substr(timestamp, 21, 3) || ':'
             || substr(timestamp, 24, 2)
WHERE timestamp LIKE '____-__-__ __:__:__ %';

-- last_scraped is a bookkeeping value, so clear it rather than convert it: the
-- scraper treats NULL as due and rewrites it on the next run.
UPDATE symbol_sources
SET last_scraped = NULL
WHERE last_scraped LIKE '____-__-__ __:__:__% %';

-- +goose Down
-- Irreversible: the original values carried no information the canonical form lacks.
