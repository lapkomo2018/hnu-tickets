package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"hnu-tickets/internal/platform/config"
	"hnu-tickets/internal/platform/database"
	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	startupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := database.Open(startupCtx, cfg.Database.URL)
	if err != nil {
		return err
	}
	defer pool.Close()

	router := newRouter(pool, cfg.CORS.AllowedOrigins)
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.App.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("http server started", "address", server.Addr, "environment", cfg.App.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	select {
	case <-shutdownSignal.Done():
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("serve HTTP: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	return server.Shutdown(shutdownCtx)
}

func newRouter(pool *pgxpool.Pool, allowedOrigins []string) http.Handler {
	router := chi.NewRouter()
	router.Use(appmiddleware.RequestID)
	router.Use(appmiddleware.Recoverer)
	router.Use(appmiddleware.RequestLogger)
	router.Use(appmiddleware.CORS(allowedOrigins))

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	return router
}
