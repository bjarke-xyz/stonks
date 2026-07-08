package db

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// PriceQuote is the latest price for a symbol, alongside today's opening price
// and the previous day's closing price. Change absolute/percentage are derived
// from these in the domain layer, not in SQL.
type PriceQuote struct {
	LatestPrice          decimal.Decimal
	Currency             string
	Timestamp            time.Time
	OpeningPrice         decimal.Decimal
	PreviousClosingPrice decimal.Decimal
}

// HistoricalPrice is a price point without its symbol or row id.
type HistoricalPrice struct {
	Price     decimal.Decimal
	Currency  string
	Timestamp time.Time
}

func (r *Repo) InsertPrice(ctx context.Context, symbolID int64, price decimal.Decimal, currency string, timestamp time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO prices (symbol_id, price, currency, timestamp)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT (symbol_id, timestamp, currency) DO UPDATE
		 SET price = EXCLUDED.price, timestamp = EXCLUDED.timestamp`,
		symbolID, price, currency, timestamp)
	return err
}

func (r *Repo) Quote(ctx context.Context, symbolID int64) (PriceQuote, error) {
	var q PriceQuote
	err := r.db.QueryRowContext(ctx,
		`WITH latest_price AS (
		     SELECT price, currency, timestamp
		     FROM prices
		     WHERE symbol_id = ?1
		     ORDER BY timestamp DESC
		     LIMIT 1
		 ),
		 opening_price AS (
		     SELECT price AS opening_price
		     FROM prices
		     WHERE symbol_id = ?1 AND DATE(timestamp) = DATE('now')
		     ORDER BY timestamp ASC
		     LIMIT 1
		 ),
		 previous_closing_price AS (
		     SELECT price AS closing_price
		     FROM prices
		     WHERE symbol_id = ?1 AND DATE(timestamp) = DATE('now', '-1 day')
		     ORDER BY timestamp DESC
		     LIMIT 1
		 )
		 SELECT lp.price, lp.currency, lp.timestamp,
		        COALESCE(op.opening_price, 0.0),
		        COALESCE(pc.closing_price, 0.0)
		 FROM latest_price lp
		 LEFT JOIN opening_price op ON 1=1
		 LEFT JOIN previous_closing_price pc ON 1=1`, symbolID,
	).Scan(&q.LatestPrice, &q.Currency, &q.Timestamp, &q.OpeningPrice, &q.PreviousClosingPrice)
	return q, err
}

func (r *Repo) HistoricalPrices(ctx context.Context, symbolID int64, startDate time.Time, endDate time.Time) ([]HistoricalPrice, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT price, currency, timestamp
		 FROM prices
		 WHERE symbol_id = ?
		   AND DATETIME(timestamp) BETWEEN DATETIME(?) AND DATETIME(?)
		 ORDER BY timestamp ASC`, symbolID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []HistoricalPrice
	for rows.Next() {
		var p HistoricalPrice
		if err := rows.Scan(&p.Price, &p.Currency, &p.Timestamp); err != nil {
			return nil, fmt.Errorf("error scanning historical price: %w", err)
		}
		prices = append(prices, p)
	}
	return prices, rows.Err()
}
