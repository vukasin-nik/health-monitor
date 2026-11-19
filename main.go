package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// monitor checks one target immediately and then on every tick,
// until ctx is cancelled. One goroutine runs this function per target.
func monitor(ctx context.Context, wg *sync.WaitGroup, client *http.Client, t Target, interval, timeout time.Duration) {
	defer wg.Done()

	run := func() {
		checkCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		record(t, Check(checkCtx, client, t.URL))
	}

	run()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			run()
		}
	}
}

func main() {
	cfgPath := getenv("CONFIG_PATH", "targets.json")
	addr := getenv("LISTEN_ADDR", ":8080")

	cfg, err := LoadConfig(cfgPath)
	if err != nil {
		slog.Error("cannot load config", "path", cfgPath, "error", err)
		os.Exit(1)
	}

	// ctx is cancelled on Ctrl+C or when Docker sends SIGTERM (docker stop).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	interval := time.Duration(cfg.IntervalSeconds) * time.Second
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	client := &http.Client{}

	var wg sync.WaitGroup
	for _, t := range cfg.Targets {
		wg.Add(1)
		go monitor(ctx, &wg, client, t, interval, timeout)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("server started", "addr", addr, "targets", len(cfg.Targets))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
	wg.Wait()
}
