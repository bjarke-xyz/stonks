CREATE TABLE  IF NOT EXISTS symbols (
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    symbol TEXT UNIQUE NOT NULL,  -- Unique ticker/identifier for stocks, ETFs, etc.
    name TEXT  -- Full name of the stock/ETF, optional
);
CREATE INDEX idx_symbols_symbol ON symbols(symbol);

CREATE TABLE IF NOT EXISTS prices (
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    symbol_id INTEGER NOT NULL,  -- Foreign key referencing Symbols.id
    price NUMERIC NOT NULL,  -- Price of the stock/ETF
    currency TEXT NOT NULL,  -- Currency (e.g., USD, EUR)
    timestamp DATETIME NOT NULL,  -- Exact time of the price capture
    FOREIGN KEY (symbol_id) REFERENCES Symbols(id) ON DELETE CASCADE
);
CREATE INDEX idx_prices_timestamp ON prices(timestamp);

CREATE TABLE IF NOT EXISTS scraping_sources (
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    name TEXT NOT NULL,  -- Name of the site (e.g., Yahoo Finance, Bloomberg)
    base_url TEXT NOT NULL,  -- Base URL of the site
    additional_info TEXT  -- Any extra metadata (e.g., API keys, scraping notes, etc.)
);

CREATE TABLE IF NOT EXISTS symbol_sources (
    id INTEGER PRIMARY KEY,  -- Auto-incrementing ID
    symbol_id INTEGER NOT NULL,  -- Foreign key referencing Symbols.id
    source_id INTEGER NOT NULL,  -- Foreign key referencing ScrapingSources.id
    scrape_url TEXT,  -- Specific URL to scrape for the symbol (if different from the base URL)
    active BOOLEAN DEFAULT TRUE,  -- Whether this source is active for the symbol
    last_scraped DATETIME,  -- Timestamp of the last successful scrape
    FOREIGN KEY (symbol_id) REFERENCES symbol(id) ON DELETE CASCADE,
    FOREIGN KEY (source_id) REFERENCES scraping_sources(id) ON DELETE CASCADE
);
