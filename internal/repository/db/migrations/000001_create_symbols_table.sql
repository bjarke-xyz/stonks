-- +goose Up
CREATE TABLE IF NOT EXISTS symbols(
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    symbol TEXT UNIQUE NOT NULL,  -- Unique ticker/identifier for stocks, ETFs, etc.
    name TEXT,  -- Full name of the stock/ETF, optional
    isin text NOT NULL -- ISIN id of the stock/ETF
);
CREATE INDEX idx_symbols_symbol ON symbols(symbol);

CREATE TABLE IF NOT EXISTS prices (
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    symbol_id INTEGER NOT NULL,  -- Foreign key referencing symbols.id
    price NUMERIC NOT NULL,  -- Price of the stock/ETF
    currency TEXT NOT NULL,  -- Currency (e.g., USD, EUR)
    timestamp DATETIME NOT NULL,  -- Exact time of the price capture
    UNIQUE(symbol_id, currency, timestamp),
    FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);
CREATE INDEX idx_prices_timestamp ON prices(timestamp);

CREATE TABLE IF NOT EXISTS scraping_sources (
    id TEXT PRIMARY KEY,  -- Scraping source identifier (e.g., BORSFRA, MORNINGSTARUK)
    name TEXT NOT NULL,  -- Name of the site (e.g., Yahoo Finance, Bloomberg)
    base_url TEXT NOT NULL,  -- Base URL of the site
    additional_info TEXT  -- Any extra metadata (e.g., API keys, scraping notes, etc.)
);

CREATE TABLE IF NOT EXISTS symbol_sources (
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    symbol_id INTEGER NOT NULL,  -- Foreign key referencing symbols.id
    source_id TEXT NOT NULL,  -- Foreign key referencing scraping_sources.id
    scrape_url TEXT NOT NULL,  -- Specific URL to scrape for the symbol (if different from the base URL)
    active BOOLEAN DEFAULT TRUE,  -- Whether this source is active for the symbol
    last_scraped DATETIME,  -- Timestamp of the last successful scrape
    FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE,
    FOREIGN KEY (source_id) REFERENCES scraping_sources(id) ON DELETE CASCADE
);



-- test data:
-- insert into symbols (symbol, name) values ('EUNL', 'iShares Core MSCI World UCITS ETF');
-- insert into scraping_sources (id, name, base_url ) values ('BORSFRA', 'Börse Frankfurt', 'https://www.boerse-frankfurt.de');
-- insert into symbol_sources (symbol_id, source_id, scrape_url ) values (1, 'BORSFRA', 'https://api.boerse-frankfurt.de/v1/data/price_information/single?isin=IE00B4L5Y983&mic=XETR');