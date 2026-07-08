package views

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
)

//go:embed *.html
var files embed.FS

var funcs = template.FuncMap{
	"rfc3339": func(t time.Time) string { return t.Format(time.RFC3339) },
	"stamp":   func(t time.Time) string { return t.Format(time.Stamp) },
}

// Each page is parsed into its own template set, because every page defines a
// template named "content" that layout.html renders.
var pages = map[string]*template.Template{}

func init() {
	for _, page := range []string{"index.html", "quote.html", "quote_table.html", "error.html"} {
		pages[page] = template.Must(
			template.New(page).Funcs(funcs).ParseFS(files, "layout.html", page))
	}
}

type BaseViewModel struct {
	Path          string
	UnixBuildTime int64
	Title         string
}

type IndexViewModel struct {
	Base BaseViewModel
}

type ErrViewModel struct {
	Base  BaseViewModel
	Error error
}

type QuoteViewModel struct {
	Base     BaseViewModel
	Quote    core.Quote
	ChartSvg template.HTML
}

// Render writes the named page wrapped in layout.html. Output is buffered so a
// template error neither emits a half-written page nor commits a status code.
func Render(w http.ResponseWriter, status int, name string, data any) error {
	tmpl, ok := pages[name]
	if !ok {
		return fmt.Errorf("views: unknown page %q", name)
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		return err
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := buf.WriteTo(w)
	return err
}
