package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bjarke-xyz/stonks/internal/api"
	"github.com/bjarke-xyz/stonks/internal/app"
	"github.com/bjarke-xyz/stonks/internal/config"
	"github.com/bjarke-xyz/stonks/internal/core"
	"github.com/bjarke-xyz/stonks/internal/repository/db"
	"github.com/bjarke-xyz/stonks/internal/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {

	// Create a context that will be canceled when we receive a shutdown signal
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	// Channel to listen for OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := config.NewConfig()
	if err != nil {
		log.Printf("failed to load config: %v", err)
	}

	dbConn, err := db.Open(cfg)
	if err != nil {
		log.Printf("error opening db: %v", err)
	}
	if dbConn != nil {
		err = db.Migrate("up", dbConn)
		if err != nil {
			log.Printf("failed to migrate: %v", err)
		}
	}

	appContext := app.AppContext(cfg)

	runMetricsServer()

	srv := Server(appContext)
	go func() {
		log.Printf("Starting server on http://localhost%v", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("ListenAndServe(): %v", err)
		}
	}()

	// Block until we receive a signal
	<-stop
	log.Println("Shutting down server...")

	// Cancel the context to signal all handlers that the server is shutting down
	// cancel()

	// Create a context with a timeout for the server shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	// Shutdown the server gracefully
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server exited properly")

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
