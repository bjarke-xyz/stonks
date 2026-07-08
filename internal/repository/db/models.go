package db

import (
	"database/sql"
	"time"

	"github.com/shopspring/decimal"
)

type Symbol struct {
	ID     int64
	Symbol string
	Name   sql.NullString
	Isin   string
}

type Price struct {
	ID        int64
	SymbolID  int64
	Price     decimal.Decimal
	Currency  string
	Timestamp time.Time
}

type ScrapingSource struct {
	ID             string
	Name           string
	BaseUrl        string
	AdditionalInfo sql.NullString
}

type SymbolSource struct {
	ID          int64
	SymbolID    int64
	SourceID    string
	ScrapeUrl   string
	Active      sql.NullBool
	LastScraped sql.NullTime
}
