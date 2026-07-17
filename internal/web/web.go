package web

import (
	"embed"
	"encoding/xml"
	"io/fs"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/web/views"
)

//go:embed static/*
var static embed.FS

type web struct {
	appContext *core.AppContext
}

func NewWeb(appContext *core.AppContext) *web {
	return &web{appContext: appContext}
}

func (h *web) Route(mux *http.ServeMux) {
	staticFiles(mux, static)
	// "/{$}" matches only "/". A bare "GET /" is a catch-all prefix and would
	// swallow every otherwise-unmatched path. GET patterns also serve HEAD.
	mux.HandleFunc("GET /{$}", h.HandleGetIndex)
	mux.HandleFunc("GET /quote/{symbol}", h.HandleGetQuote)
	// gin redirected /quote/AAPL/ to /quote/AAPL. ServeMux would 404 it, so keep
	// the redirect for bookmarked or hand-typed URLs.
	mux.HandleFunc("GET /quote/{symbol}/{$}", redirectTrailingSlash)
}

func redirectTrailingSlash(w http.ResponseWriter, r *http.Request) {
	target := strings.TrimSuffix(r.URL.Path, "/")
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}

func (h *web) getBaseModel(r *http.Request, title string) views.BaseViewModel {
	var unixBuildTime int64 = 0
	if h.appContext.Config.BuildTime != nil {
		unixBuildTime = h.appContext.Config.BuildTime.Unix()
	} else {
		unixBuildTime = time.Now().Unix()
	}
	return views.BaseViewModel{
		Path:          r.URL.Path,
		UnixBuildTime: unixBuildTime,
		Title:         title + " | stonks",
	}
}

func (h *web) handleError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("handler error", "method", r.Method, "path", r.URL.Path, "error", err)
	if renderErr := views.Render(w, http.StatusInternalServerError, "error.html", views.ErrViewModel{
		Base:  h.getBaseModel(r, "error"),
		Error: err,
	}); renderErr != nil {
		slog.Error("rendering error page failed", "error", renderErr)
	}
}

// queryOr returns the query parameter, or defaultVal when it is absent or empty.
func queryOr(r *http.Request, key string, defaultVal string) string {
	if val := r.URL.Query().Get(key); val != "" {
		return val
	}
	return defaultVal
}

// writeXML mirrors what gin's c.XML emitted: this exact content type, and no
// XML declaration. LibreOffice Calc consumes /quote/{symbol}?format=xml.
func writeXML(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.WriteHeader(status)
	return xml.NewEncoder(w).Encode(data)
}

func staticFiles(mux *http.ServeMux, staticFs fs.FS) {
	staticWeb, err := fs.Sub(staticFs, "static")
	if err != nil {
		slog.Error("static fs sub failed", "error", err)
		return
	}
	mux.Handle("GET /static/", immutableCache(http.StripPrefix("/static/", http.FileServerFS(staticWeb))))
	for _, name := range []string{
		"favicon.ico",
		"favicon-16x16.png",
		"favicon-32x32.png",
		"apple-touch-icon.png",
		"site.webmanifest",
	} {
		mux.Handle("GET /"+name, serveFile(staticWeb, name))
	}
}

func serveFile(fsys fs.FS, name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFileFS(w, r, fsys, name)
	})
}

// immutableCache marks the cache-busted js and css assets as permanently
// cacheable. It must wrap StripPrefix rather than sit inside it, so that
// r.URL.Path is still the full request path.
func immutableCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/static/js") || strings.HasPrefix(r.URL.Path, "/static/css") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		next.ServeHTTP(w, r)
	})
}
