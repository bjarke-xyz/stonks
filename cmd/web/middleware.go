package main

import (
	"log"
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
			log.Printf("panic serving %s %s: %v\n%s", r.Method, r.URL.Path, rec, debug.Stack())
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
		log.Printf("%d %-4s %s %v", rec.status, r.Method, r.URL.Path, time.Since(start).Round(time.Microsecond))
	})
}
