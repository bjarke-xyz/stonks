package scrapers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/repository/db/dao"
	"github.com/shopspring/decimal"
)

type YahooScraper struct {
	appContext *core.AppContext
}

func NewYahooScraper(appContext *core.AppContext) Scraper {
	return &YahooScraper{appContext: appContext}
}

// SourceIdentifier implements Scraper.
func (y *YahooScraper) SourceIdentifier() string {
	return ScrapingSourceIdentifierYAHOO
}

// Scrape implements Scraper.
func (y *YahooScraper) Scrape(ctx context.Context, symbol dao.Symbol) (ScrapeResult, error) {
	queries, err := db.OpenQueries(y.appContext.Config)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error opening db: %w", err)
	}

	symbolSource, err := queries.GetSymbolSource(ctx, dao.GetSymbolSourceParams{
		SymbolID: symbol.ID,
		SourceID: y.SourceIdentifier(),
	})
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error getting symbol source: %w", err)
	}

	req, err := http.NewRequest("GET", symbolSource.ScrapeUrl, nil)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error making http req: %w", err)
	}
	for k, v := range nasdaqDefaultHeaders {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error HTTP GETting %v: %w", symbolSource.ScrapeUrl, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error reading response body: %w", err)
	}
	bodyStr := string(bodyBytes)

	// Find price in <span data-testid="qsp-price">VALUE</span>
	re := regexp.MustCompile(`<span[^>]*data-testid="qsp-price"[^>]*>([^<]+)</span>`)
	m := re.FindStringSubmatch(bodyStr)
	if len(m) < 2 {
		return ScrapeResult{}, fmt.Errorf("couldn't find price span in page")
	}
	priceRaw := strings.TrimSpace(m[1])
	priceRaw = strings.ReplaceAll(priceRaw, ",", "")
	price, err := decimal.NewFromString(priceRaw)
	if err != nil {
		return ScrapeResult{}, fmt.Errorf("error parsing price %v: %w", priceRaw, err)
	}

	// Find currency by searching for "currency":"..." or escaped \"currency\":\"...\"
	currency := ""
	if idx := strings.Index(bodyStr, `\\"currency\\":"`); idx != -1 {
		start := idx + len(`\\"currency\\":"`)
		if end := strings.Index(bodyStr[start:], `"`); end != -1 {
			currency = bodyStr[start : start+end]
		}
	}
	if currency == "" {
		if idx := strings.Index(bodyStr, `\"currency\":\"`); idx != -1 {
			start := idx + len(`\"currency\":\"`)
			if end := strings.Index(bodyStr[start:], `\"`); end != -1 {
				currency = bodyStr[start : start+end]
			}
		}
	}
	if currency == "" {
		// try unescaped form
		if idx := strings.Index(bodyStr, `"currency":"`); idx != -1 {
			start := idx + len(`"currency":"`)
			if end := strings.Index(bodyStr[start:], `"`); end != -1 {
				currency = bodyStr[start : start+end]
			}
		}
	}
	// final fallback via regex
	if currency == "" {
		rec := regexp.MustCompile(`"currency"\s*:\s*"([^"]+)"`)
		mc := rec.FindStringSubmatch(bodyStr)
		if len(mc) >= 2 {
			currency = mc[1]
		}
	}
	if currency == "" {
		return ScrapeResult{}, fmt.Errorf("couldn't find currency in page")
	}

	// Find timestamp: look for "At close: ..."
	atIdx := strings.Index(bodyStr, "At close:")
	if atIdx == -1 {
		return ScrapeResult{}, fmt.Errorf("couldn't find 'At close:' in page")
	}
	tsRaw := bodyStr[atIdx+len("At close:"):]
	// stop at next tag or quote or newline
	if end := strings.IndexAny(tsRaw, "<\"\n"); end != -1 {
		tsRaw = tsRaw[:end]
	}
	tsRaw = strings.TrimSpace(tsRaw) // e.g. "December 19 at 5:00:00 PM GMT+1"

	// Extract timezone token if present (e.g. GMT+1)
	parts := strings.Fields(tsRaw)
	tzTok := ""
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if strings.HasPrefix(last, "GMT") {
			tzTok = last
			// remove tz from timestamp string
			tsRaw = strings.TrimSpace(strings.TrimSuffix(tsRaw, tzTok))
		}
	}

	// Append current year and parse
	year := time.Now().Year()
	dtStr := fmt.Sprintf("%s %d", tsRaw, year) // e.g. "December 19 at 5:00:00 PM 2025"
	layout := "January 2 at 3:04:05 PM 2006"
	var parsed time.Time
	if tzTok != "" && strings.HasPrefix(tzTok, "GMT") {
		// parse offset like GMT+1 or GMT-01:00
		off := 0
		sign := tzTok[3]
		val := tzTok[4:]
		if val == "" {
			off = 0
		} else if strings.Contains(val, ":") {
			parts := strings.Split(val, ":")
			h, _ := strconv.Atoi(parts[0])
			m, _ := strconv.Atoi(parts[1])
			off = h*3600 + m*60
		} else {
			h, _ := strconv.Atoi(val)
			off = h * 3600
		}
		if sign == '-' {
			off = -off
		}
		loc := time.FixedZone(tzTok, off)
		parsed, err = time.ParseInLocation(layout, dtStr, loc)
		if err != nil {
			return ScrapeResult{}, fmt.Errorf("error parsing timestamp %v: %w", dtStr, err)
		}
	} else {
		loc, _ := time.LoadLocation("Europe/Copenhagen")
		parsed, err = time.ParseInLocation(layout, dtStr, loc)
		if err != nil {
			return ScrapeResult{}, fmt.Errorf("error parsing timestamp %v: %w", dtStr, err)
		}
	}

	return ScrapeResult{
		Price:     price,
		Currency:  currency,
		Timestamp: parsed,
	}, nil
}
