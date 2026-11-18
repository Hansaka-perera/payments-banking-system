package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("/wallet", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("health check", "remote", r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"status\":\"ok\"}" + "\n"))

	})

	SRV := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting wallet service", "addr", SRV.Addr)
		if err := SRV.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("error starting wallet service", "error", err)
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info("shutting down wallet service")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := SRV.Shutdown(shutdownCtx); err != nil {
		logger.Error("error shutting down wallet service", "error", err)
	}

	logger.Info("wallet service stopped")
}
