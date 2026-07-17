package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bjarke-xyz/stonks/internal/api"
	"github.com/bjarke-xyz/stonks/internal/app"
	"github.com/bjarke-xyz/stonks/internal/config"
	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/logging"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	logging.Setup()

	// Create a context that will be canceled when we receive a shutdown signal
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Error("failed to load config", "error", err)
	}

	dbConn, err := db.Open(cfg)
	if err != nil {
		slog.Error("opening db failed", "error", err)
	}
	if dbConn != nil {
		err = db.Migrate("up", dbConn)
		if err != nil {
			slog.Error("migration failed", "error", err)
		}
	}

	appContext := app.AppContext(cfg)

	runMetricsServer()

	srv := Server(appContext)
	go func() {
		slog.Info("listening", "addr", srv.Addr, "url", "http://localhost"+srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen and serve failed", "error", err)
			os.Exit(1)
		}
	}()

	// Block until we receive a signal
	<-stop
	slog.Info("shutting down server")

	// Cancel the context to signal all handlers that the server is shutting down
	// cancel()

	// Create a context with a timeout for the server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown the server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown failed", "error", err)
		os.Exit(1)
	}
	slog.Info("server exited properly")

}

func Server(appContext *core.AppContext) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%d", appContext.Config.Port),
		Handler: routes(appContext),
	}
}

func routes(appContext *core.AppContext) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealth)

	apiHandlers := api.NewAPI(appContext)
	apiHandlers.Route(mux)

	webHandlers := web.NewWeb(appContext)
	webHandlers.Route(mux)

	// recovery outermost, so a panic in requestLog or a handler is still caught.
	return recovery(requestLog(mux))
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	body, err := json.Marshal(map[string]string{"status": "ok"})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(body)
}

func runMetricsServer() {
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9091", mux)
	}()
}
