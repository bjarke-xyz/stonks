package quote

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
)

type QuoteService struct {
	appContext *core.AppContext
}

func NewQuoteService(appContext *core.AppContext) core.QuoteService {
	return &QuoteService{appContext: appContext}
}

func (q *QuoteService) ClearCache(ctx context.Context, tickerSymbol string) error {
	return q.appContext.Deps.Cache.DeleteByPrefix("QUOTE:" + tickerSymbol)
}
func (q *QuoteService) GetQuote(ctx context.Context, tickerSymbol string, startDate time.Time, endDate time.Time) (core.Quote, error) {
	tickerSymbol = strings.ToUpper(tickerSymbol)
	cacheKey := fmt.Sprintf("QUOTE:%v:%v:%v", tickerSymbol, startDate.Unix(), endDate.Unix())
	quote := core.Quote{}
	inCache, _ := q.appContext.Deps.Cache.GetObj(cacheKey, &quote)
	if inCache {
		log.Printf("got %v quote from cache", tickerSymbol)
		return quote, nil
	}
	queries, err := db.OpenQueries(q.appContext.Config)
	if err != nil {
		return core.Quote{}, fmt.Errorf("error opening queries: %w", err)
	}

	symbol, err := queries.GetSymbolByTicker(ctx, tickerSymbol)
	if err != nil {
		return core.Quote{}, fmt.Errorf("error getting symbol: %w", err)
	}

	priceQuote, err := queries.GetQuote(ctx, symbol.ID)
	if err != nil {
		return core.Quote{}, fmt.Errorf("error getting price for symbol %v: %w", symbol.Symbol, err)
	}

	dbHistoricalPrices, err := queries.GetHistoricalPrices(ctx, dao.GetHistoricalPricesParams{
		SymbolID:  symbol.ID,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return core.Quote{}, fmt.Errorf("error getting historical prices for symbol %v: %w", symbol.Symbol, err)
	}
	historicalPrices := make([]core.SimplePrice, len(dbHistoricalPrices))
	for i, histPrice := range dbHistoricalPrices {
		historicalPrices[i] = core.SimplePrice{
			Price:     histPrice.Price,
			Currency:  histPrice.Currency,
			Timestamp: histPrice.Timestamp,
		}
	}

	quote = core.Quote{
		Symbol: core.Symbol{
			Symbol: symbol.Symbol,
			Name:   symbol.Name.String,
		},
		Price: core.Price{
			Price:                priceQuote.LatestPrice,
			Currency:             priceQuote.Currency,
			Timestamp:            priceQuote.Timestamp,
			OpeningPrice:         priceQuote.OpeningPrice,
			PreviousClosingPrice: priceQuote.PreviousClosingPrice,
		},
		HistoricalPrices: historicalPrices,
	}
	q.appContext.Deps.Cache.InsertObj(cacheKey, quote, 30)
	return quote, nil
}
