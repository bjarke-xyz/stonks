package views

import (
	"bytes"
	"embed"
	"html/template"
	"net/http"
	"time"

	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/gin-gonic/gin/render"
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

type Flash struct {
	Type string
	Msg  string
}

// FlashesOf turns core's capitalised flash type into a lowercase CSS suffix.
func FlashesOf(flashType string, msgs []string) []Flash {
	suffix := map[string]string{
		core.FlashTypeInfo:  "info",
		core.FlashTypeWarn:  "warn",
		core.FlashTypeError: "error",
	}[flashType]
	flashes := make([]Flash, len(msgs))
	for i, msg := range msgs {
		flashes[i] = Flash{Type: suffix, Msg: msg}
	}
	return flashes
}

type BaseViewModel struct {
	Path          string
	UnixBuildTime int64
	Title         string
	Flashes       []Flash
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

// Renderer implements gin's render.HTMLRender. Handlers keep calling
// c.HTML(status, "quote.html", model).
type Renderer struct{}

func (Renderer) Instance(name string, data any) render.Render {
	tmpl, ok := pages[name]
	if !ok {
		panic("views: unknown page " + name)
	}
	return pageRender{tmpl: tmpl, data: data}
}

type pageRender struct {
	tmpl *template.Template
	data any
}

func (p pageRender) Render(w http.ResponseWriter) error {
	// Buffer so a template error doesn't emit a half-written page.
	var buf bytes.Buffer
	if err := p.tmpl.ExecuteTemplate(&buf, "layout.html", p.data); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

func (pageRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
}
