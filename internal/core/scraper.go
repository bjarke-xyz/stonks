package core

import "context"

type ScraperService interface {
	ScrapeSymbols(ctx context.Context)
}
