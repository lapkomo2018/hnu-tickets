package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"hnu-tickets/docs"

	"hnu-tickets/internal/analytics"
	"hnu-tickets/internal/auth"
	"hnu-tickets/internal/categories"
	"hnu-tickets/internal/comments"
	"hnu-tickets/internal/history"
	"hnu-tickets/internal/platform/apidocs"
	"hnu-tickets/internal/platform/config"
	"hnu-tickets/internal/platform/database"
	"hnu-tickets/internal/platform/httpx"
	appmiddleware "hnu-tickets/internal/platform/middleware"
	"hnu-tickets/internal/tickets"
	"hnu-tickets/internal/users"
)

//go:generate go run hnu-tickets/cmd/swagger

// @title HNU Tickets API
// @version 1.0.0
// @description Technical support ticket management API with role based access control.
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @tag.name Auth
// @tag.description Registration, login and the authenticated profile
// @tag.name Users
// @tag.description User administration and role changes
// @tag.name Categories
// @tag.description Ticket categories
// @tag.name Tickets
// @tag.description Ticket lifecycle, assignment and listing
// @tag.name Comments
// @tag.description Conversation on a ticket
// @tag.name History
// @tag.description Immutable ticket audit log
// @tag.name Analytics
// @tag.description Administrative ticket metrics
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

	router := newRouter(pool, cfg, specification(cfg))
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

// specification regenerates the document from the annotations while the sources
// are at hand, and falls back to the copy compiled into the binary.
func specification(cfg config.Config) string {
	if cfg.App.Environment != config.EnvironmentDevelopment {
		return docs.SwaggerInfo.ReadDoc()
	}
	generated, err := apidocs.Generate()
	if err != nil {
		slog.Warn("serving the embedded API specification", "error", err)
		return docs.SwaggerInfo.ReadDoc()
	}
	slog.Info("API specification regenerated from source")
	return generated
}

func newRouter(pool *pgxpool.Pool, cfg config.Config, specification string) http.Handler {
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
	errorMiddleware := httpx.NewErrorMiddleware(slog.Default())

	router := chi.NewRouter()
	router.Use(appmiddleware.RequestID)
	router.Use(appmiddleware.RequestLogger)
	router.Use(errorMiddleware.Recoverer)
	router.Use(appmiddleware.CORS(cfg.CORS.AllowedOrigins))
	router.NotFound(errorMiddleware.NotFound)
	router.MethodNotAllowed(errorMiddleware.MethodNotAllowed)

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
	router.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/index.html", http.StatusPermanentRedirect)
	})
	router.Get("/docs/doc.json", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, specification)
	})
	router.Get("/docs/*", httpSwagger.Handler(httpSwagger.URL("/docs/doc.json")))

	router.Route("/api/v1", func(api chi.Router) {
		api.Route("/auth", func(authRouter chi.Router) {
			auth.PublicRoutes(authRouter, authHandler, errorMiddleware)
			authRouter.Group(func(protectedAuth chi.Router) {
				protectedAuth.Use(tokenManager.Middleware(errorMiddleware))
				auth.ProtectedRoutes(protectedAuth, authHandler, errorMiddleware)
			})
		})
		api.Group(func(protected chi.Router) {
			protected.Use(tokenManager.Middleware(errorMiddleware))
			protected.Route("/users", func(usersRouter chi.Router) {
				users.Routes(usersRouter, usersHandler, errorMiddleware)
			})
			protected.Route("/categories", func(categoriesRouter chi.Router) {
				categories.Routes(categoriesRouter, categoriesHandler, errorMiddleware)
			})
			protected.Route("/tickets", func(ticketsRouter chi.Router) {
				tickets.Routes(ticketsRouter, ticketsHandler, commentsHandler, historyHandler, errorMiddleware)
			})
			protected.Route("/analytics", func(analyticsRouter chi.Router) {
				analytics.Routes(analyticsRouter, analyticsHandler, errorMiddleware)
			})
		})
	})
	return router
}
