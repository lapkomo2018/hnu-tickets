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

	"hnu-tickets/internal/analytics"
	"hnu-tickets/internal/auth"
	"hnu-tickets/internal/categories"
	"hnu-tickets/internal/comments"
	"hnu-tickets/internal/history"
	"hnu-tickets/internal/platform/config"
	"hnu-tickets/internal/platform/database"
	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/platform/swagger"
	"hnu-tickets/internal/tickets"
	"hnu-tickets/internal/users"
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

	router := newRouter(pool, cfg)
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

func newRouter(pool *pgxpool.Pool, cfg config.Config) http.Handler {
	usersRepository := users.NewPostgresRepository(pool)
	usersService := users.NewService(usersRepository)
	usersHandler := users.NewHandler(usersService)

	tokenManager := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.TTL)
	authService := auth.NewService(usersRepository, tokenManager)
	authHandler := auth.NewHandler(authService)

	categoriesRepository := categories.NewPostgresRepository(pool)
	categoriesService := categories.NewService(categoriesRepository)
	categoriesHandler := categories.NewHandler(categoriesService)

	historyRepository := history.NewPostgresRepository(pool)
	ticketsRepository := tickets.NewPostgresRepository(pool)
	ticketsService := tickets.NewService(
		ticketsRepository, historyRepository, usersRepository, database.NewTransactor(pool),
	)
	ticketsHandler := tickets.NewHandler(ticketsService)

	commentsRepository := comments.NewPostgresRepository(pool)
	commentsService := comments.NewService(commentsRepository, ticketsService)
	commentsHandler := comments.NewHandler(commentsService)
	historyService := history.NewService(historyRepository, ticketsService)
	historyHandler := history.NewHandler(historyService)

	analyticsRepository := analytics.NewPostgresRepository(pool)
	analyticsService := analytics.NewService(analyticsRepository)
	analyticsHandler := analytics.NewHandler(analyticsService)

	router := chi.NewRouter()
	router.Use(appmiddleware.RequestID)
	router.Use(appmiddleware.Recoverer)
	router.Use(appmiddleware.RequestLogger)
	router.Use(appmiddleware.CORS(cfg.CORS.AllowedOrigins))
	router.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusNotFound, map[string]any{
			"error": map[string]string{"code": "ROUTE_NOT_FOUND", "message": "route not found"},
		})
	})
	router.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		httpx.JSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": map[string]string{"code": "METHOD_NOT_ALLOWED", "message": "method not allowed"},
		})
	})

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
	router.Get("/openapi.yaml", swagger.OpenAPI)
	router.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusPermanentRedirect)
	})
	router.Get("/swagger/", swagger.UI)

	router.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(authRouter chi.Router) {
			auth.PublicRoutes(authRouter, authHandler)
			authRouter.Group(func(protectedAuth chi.Router) {
				protectedAuth.Use(tokenManager.Middleware)
				auth.ProtectedRoutes(protectedAuth, authHandler)
			})
		})
		api.Group(func(protected chi.Router) {
			protected.Use(tokenManager.Middleware)
			protected.Route("/users", func(usersRouter chi.Router) {
				users.Routes(usersRouter, usersHandler)
			})
			protected.Route("/categories", func(categoriesRouter chi.Router) {
				categories.Routes(categoriesRouter, categoriesHandler)
			})
			protected.Route("/tickets", func(ticketsRouter chi.Router) {
				tickets.Routes(ticketsRouter, ticketsHandler, commentsHandler, historyHandler)
			})
			protected.Route("/analytics", func(analyticsRouter chi.Router) {
				analytics.Routes(analyticsRouter, analyticsHandler)
			})
		})
	})
	return router
}
