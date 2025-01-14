package quote

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
)

type QuoteService struct {
	appContext *core.AppContext
}

func NewQuoteService(appContext *core.AppContext) core.QuoteService {
	return &QuoteService{appContext: appContext}
}

func (q *QuoteService) GetQuote(ctx context.Context, tickerSymbol string) (core.Quote, error) {
	tickerSymbol = strings.ToUpper(tickerSymbol)
	cacheKey := "QUOTE:v2:" + tickerSymbol
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
		return core.Quote{}, fmt.Errorf("error getting price for symbol %+v: %w", symbol, err)
	}

	quote = core.Quote{
		Symbol: core.Symbol{
			Symbol: symbol.Symbol,
			Name:   symbol.Name.String,
		},
		Price: core.Price{
			Price:        priceQuote.LatestPrice,
			Currency:     priceQuote.Currency,
			Timestamp:    priceQuote.Timestamp,
			OpeningPrice: priceQuote.OpeningPrice,
		},
	}
	q.appContext.Deps.Cache.InsertObj(cacheKey, quote, 30)
	return quote, nil
}
