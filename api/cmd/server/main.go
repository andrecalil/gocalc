// Command server is the HTTP entry point. Wiring only — all logic lives in
// internal packages so it's testable without spinning up a server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/andrecalil/go-calc/api/internal/httpapi"
)

func main() {
	// `server -health` is used by the Docker HEALTHCHECK. distroless-static
	// has no shell and no curl, so the binary gains a tiny subcommand that
	// GETs its own /healthz and exits 0 on success.
	if len(os.Args) > 1 && os.Args[1] == "-health" {
		runHealthCheck()
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           httpapi.NewHandler(logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	// Graceful shutdown: SIGINT / SIGTERM gives in-flight requests up to 10s.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
}

// runHealthCheck is called when the binary is invoked with `-health`. It
// sends a single GET to /healthz on the address the server would use and
// exits 0 on 2xx, 1 otherwise. Used by the Docker HEALTHCHECK on distroless
// images where no shell or curl is available.
func runHealthCheck() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	url := "http://127.0.0.1" + addr + "/healthz"
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		os.Exit(1)
	}
}
