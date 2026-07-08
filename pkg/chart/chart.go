// Package chart renders a minimal single-series line chart as an SVG document.
package chart

import (
	"fmt"
	"html"
	"math"
	"strconv"
	"strings"
)

const (
	lineColor = "#5470c6"
	gridColor = "#e0e6f1"
	textColor = "#666666"

	padLeft   = 60
	padRight  = 20
	padTop    = 36
	padBottom = 32

	yTicks    = 5
	maxXLabel = 10
)

// LineChart is a single series of Values plotted against XLabels.
type LineChart struct {
	Title   string
	Legend  string
	XLabels []string
	Values  []float64
	Width   int
	Height  int
}

// SVG renders the chart. It returns an empty string if there is nothing to plot.
func (c LineChart) SVG() string {
	if len(c.Values) == 0 {
		return ""
	}
	w, h := c.Width, c.Height
	if w <= 0 {
		w = 800
	}
	if h <= 0 {
		h = 400
	}

	x0, y0 := float64(padLeft), float64(padTop)
	x1, y1 := float64(w-padRight), float64(h-padBottom)

	lo, hi, step := niceRange(minMax(c.Values))
	prec := decimals(step)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="100%%" font-family="sans-serif" font-size="12">`, w, h)

	if c.Title != "" {
		fmt.Fprintf(&b, `<text x="%d" y="20" fill="%s" font-size="14">%s</text>`, padLeft, textColor, html.EscapeString(c.Title))
	}
	if c.Legend != "" {
		mid := (x0 + x1) / 2
		fmt.Fprintf(&b, `<line x1="%s" y1="16" x2="%s" y2="16" stroke="%s" stroke-width="2"/>`,
			num(mid-30), num(mid-10), lineColor)
		fmt.Fprintf(&b, `<text x="%s" y="20" fill="%s">%s</text>`, num(mid-4), textColor, html.EscapeString(c.Legend))
	}

	// Horizontal grid lines with y-axis labels.
	for i := 0; i <= int(math.Round((hi-lo)/step)); i++ {
		v := lo + float64(i)*step
		y := y1 - (v-lo)/(hi-lo)*(y1-y0)
		fmt.Fprintf(&b, `<line x1="%s" y1="%s" x2="%s" y2="%s" stroke="%s"/>`, num(x0), num(y), num(x1), num(y), gridColor)
		fmt.Fprintf(&b, `<text x="%s" y="%s" fill="%s" text-anchor="end">%s</text>`,
			num(x0-8), num(y+4), textColor, strconv.FormatFloat(v, 'f', prec, 64))
	}

	// Data points, doubling as x-axis label anchors.
	xs := make([]float64, len(c.Values))
	for i, v := range c.Values {
		if len(c.Values) == 1 {
			xs[i] = (x0 + x1) / 2
		} else {
			xs[i] = x0 + float64(i)/float64(len(c.Values)-1)*(x1-x0)
		}
		y := y1 - (v-lo)/(hi-lo)*(y1-y0)
		if i == 0 {
			b.WriteString(`<polyline fill="none" stroke="` + lineColor + `" stroke-width="2" points="`)
		} else {
			b.WriteByte(' ')
		}
		b.WriteString(num(xs[i]) + "," + num(y))
	}
	b.WriteString(`"/>`)

	stride := max(1, (len(c.XLabels)+maxXLabel-1)/maxXLabel)
	for i, label := range c.XLabels {
		if i%stride != 0 || i >= len(xs) {
			continue
		}
		fmt.Fprintf(&b, `<text x="%s" y="%s" fill="%s" text-anchor="middle">%s</text>`,
			num(xs[i]), num(y1+18), textColor, html.EscapeString(label))
	}

	b.WriteString(`</svg>`)
	return b.String()
}

func minMax(values []float64) (float64, float64) {
	lo, hi := values[0], values[0]
	for _, v := range values[1:] {
		lo, hi = min(lo, v), max(hi, v)
	}
	return lo, hi
}

// niceRange expands [lo, hi] outwards to round numbers divisible by the returned
// step, sized so the axis carries at most yTicks intervals.
func niceRange(lo, hi float64) (float64, float64, float64) {
	if hi <= lo {
		hi = lo + 1
	}
	step := niceNum((hi - lo) / yTicks)
	return math.Floor(lo/step) * step, math.Ceil(hi/step) * step, step
}

// niceNum rounds x up to a 1, 2, 5 or 10 multiple of a power of ten. Rounding up
// rather than to the nearest keeps the interval count at or below yTicks.
func niceNum(x float64) float64 {
	if x <= 0 {
		return 1
	}
	exp := math.Floor(math.Log10(x))
	frac := x / math.Pow(10, exp)
	var nice float64
	switch {
	case frac <= 1:
		nice = 1
	case frac <= 2:
		nice = 2
	case frac <= 5:
		nice = 5
	default:
		nice = 10
	}
	return nice * math.Pow(10, exp)
}

// decimals is the number of fraction digits needed to print multiples of step.
func decimals(step float64) int {
	if step >= 1 {
		return 0
	}
	return int(math.Ceil(-math.Log10(step)))
}

func num(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
