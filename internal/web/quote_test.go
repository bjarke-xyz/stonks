package web

import (
	"bytes"
	"context"
	"encoding/xml"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bjarke-xyz/stonks/internal/config"
	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/shopspring/decimal"
)

type stubQuoteService struct{ quote core.Quote }

func (s stubQuoteService) GetQuote(ctx context.Context, tickerSymbol string, startDate, endDate time.Time) (core.Quote, error) {
	q := s.quote
	q.Symbol.Symbol = tickerSymbol
	return q, nil
}

func (s stubQuoteService) ClearCache(ctx context.Context, tickerSymbol string) error { return nil }

func testQuote() core.Quote {
	ts := time.Date(2026, 7, 8, 12, 0, 0, 0, time.UTC)
	return core.Quote{
		Symbol: core.Symbol{Symbol: "AAPL", Name: "Apple Inc."},
		Price: core.Price{
			Price:                decimal.NewFromFloat(212.5),
			Currency:             "USD",
			Timestamp:            ts,
			OpeningPrice:         decimal.NewFromFloat(210),
			PreviousClosingPrice: decimal.NewFromFloat(209.25),
		},
		HistoricalPrices: []core.SimplePrice{
			{Price: decimal.NewFromFloat(210), Currency: "USD", Timestamp: ts},
			{Price: decimal.NewFromFloat(212.5), Currency: "USD", Timestamp: ts.Add(time.Hour)},
		},
	}
}

func newTestServer(t *testing.T) *http.ServeMux {
	t.Helper()
	h := NewWeb(&core.AppContext{
		Config: &config.Config{},
		Deps:   &core.AppDeps{QuoteService: stubQuoteService{quote: testQuote()}},
	})
	mux := http.NewServeMux()
	h.Route(mux)
	return mux
}

// The XML response is consumed by LibreOffice Calc. gin's c.XML wrote
// xml.NewEncoder(w).Encode(data) under "application/xml; charset=utf-8" and
// emitted no XML declaration; writeXML must produce the identical bytes.
func TestQuoteXMLMatchesGinEncoding(t *testing.T) {
	rec := httptest.NewRecorder()
	newTestServer(t).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/quote/AAPL?format=xml", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/xml; charset=utf-8"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	// xml.Marshal produces exactly what xml.Encoder.Encode writes: no
	// declaration, no indentation, no trailing newline.
	want, err := xml.Marshal(testQuote().ToSerializableQuote())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if got := rec.Body.String(); got != string(want) {
		t.Errorf("body mismatch\n got: %s\nwant: %s", got, want)
	}
}

func TestQuoteRoutingAndFormats(t *testing.T) {
	mux := newTestServer(t)
	tests := []struct {
		name       string
		target     string
		wantStatus int
		wantCtype  string
		wantSymbol bool
	}{
		{"html default", "/quote/AAPL?chart=false", 200, "text/html; charset=utf-8", true},
		{"table format", "/quote/AAPL?format=table&chart=false", 200, "text/html; charset=utf-8", true},
		{"xml format", "/quote/AAPL?format=xml", 200, "application/xml; charset=utf-8", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.target, nil))
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := rec.Header().Get("Content-Type"); got != tt.wantCtype {
				t.Errorf("Content-Type = %q, want %q", got, tt.wantCtype)
			}
			// PathValue("symbol") must reach the service.
			if tt.wantSymbol && !bytes.Contains(rec.Body.Bytes(), []byte("AAPL")) {
				t.Errorf("response does not mention the requested symbol")
			}
		})
	}
}

// TestIndexIsNotACatchAll guards the ServeMux "GET /{$}" anchor: a bare
// "GET /" pattern would render the index for every unmatched path.
func TestIndexIsNotACatchAll(t *testing.T) {
	mux := newTestServer(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-such-page", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /no-such-page = %d, want 404", rec.Code)
	}
}

// gin's RedirectTrailingSlash behaviour, preserved so an existing bookmark or
// spreadsheet URL with a trailing slash keeps working.
func TestQuoteTrailingSlashRedirects(t *testing.T) {
	mux := newTestServer(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/quote/AAPL/?format=xml", nil))

	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("status = %d, want 301", rec.Code)
	}
	if got, want := rec.Header().Get("Location"), "/quote/AAPL?format=xml"; got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
}

func TestStaticAssetsCacheHeaders(t *testing.T) {
	mux := newTestServer(t)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got, want := rec.Header().Get("Cache-Control"), "public, max-age=31536000, immutable"; got != want {
		t.Errorf("Cache-Control = %q, want %q", got, want)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/favicon.ico", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("favicon status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Cache-Control"); got != "" {
		t.Errorf("favicon Cache-Control = %q, want empty (matches previous behaviour)", got)
	}
}
