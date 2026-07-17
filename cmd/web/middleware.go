package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

// recovery turns a panic in a handler into a 500 instead of taking the process
// down. It replaces gin.Recovery().
func recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// net/http treats ErrAbortHandler as a deliberate abort and already
			// suppresses it, so pass it through rather than logging a stack.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			// The stack is its own attribute rather than part of the message:
			// under the JSON handler that keeps the whole trace in one record
			// and one Loki event, recoverable with `jq -r .stack`.
			slog.Error("panic serving request",
				"method", r.Method,
				"path", r.URL.Path,
				"panic", fmt.Sprint(rec),
				"stack", string(debug.Stack()),
			)
			w.WriteHeader(http.StatusInternalServerError)
		}()
		next.ServeHTTP(w, r)
	})
}

// statusRecorder captures the status code so requestLog can report it.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// requestLog replaces gin.Logger().
func requestLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		path := r.URL.Path
		if r.URL.RawQuery != "" {
			path += "?" + r.URL.RawQuery
		}
		// duration_ms is a float rather than a time.Duration because the two
		// handlers encode a Duration differently — JSON writes raw nanoseconds,
		// text writes "1.23ms" — which would make the field's type depend on
		// LOG_FORMAT. Milliseconds rather than whole ms: these handlers mostly
		// finish in under one.
		slog.Info("request",
			"method", r.Method,
			"path", path,
			"status", rec.status,
			"duration_ms", float64(time.Since(start).Microseconds())/1000,
			"ip", clientIP(r),
		)
	})
}

// clientIP is only ever used for the log line. In production this sits behind
// Cloudflare, which is the only party allowed to name the client — RemoteAddr
// would just be the proxy.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.Header.Get("CF-Connecting-IP")
}
