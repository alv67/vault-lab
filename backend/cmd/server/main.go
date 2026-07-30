package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/amelamela/vault-lab/internal/auth"
	"github.com/amelamela/vault-lab/internal/config"
	"github.com/amelamela/vault-lab/internal/handler"
	"github.com/amelamela/vault-lab/internal/repository"
	"github.com/amelamela/vault-lab/internal/service"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	cfg := config.Load()

	zerolog.SetGlobalLevel(parseLogLevel(cfg.LogLevel))

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrations(cfg)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer dbPool.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn().Err(err).Msg("redis not available, continuing without cache")
	}

	jwtAuth := auth.NewJWTAuth(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	repos := repository.New(dbPool)
	svc := service.New(repos, jwtAuth)

	h := handler.New(svc, jwtAuth)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	setupRoutes(r, h, jwtAuth)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("forced shutdown")
	}
	log.Info().Msg("server stopped")
}

func runMigrations(cfg *config.Config) {
	m, err := migrate.New(
		"file://migrations",
		cfg.DSN(),
	)
	if err != nil {
		log.Fatal().Err(err).Msg("migration init failed")
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Err(err).Msg("migration failed")
	}

	log.Info().Msg("migrations applied successfully")
}

func setupRoutes(r chi.Router, h *handler.Handler, jwtAuth *auth.JWTAuth) {
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)
		r.Post("/auth/refresh", h.RefreshToken)

		r.Group(func(r chi.Router) {
			r.Use(jwtAuth.Middleware)
			r.Get("/users/me", h.GetCurrentUser)
			r.Patch("/users/me", h.UpdateCurrentUser)

			r.Get("/assets", h.ListAssets)
			r.Get("/assets/search", h.SearchAssets)
			r.Get("/assets/lookup", h.LookupAsset)
			r.Get("/assets/{id}", h.GetAsset)
			r.Post("/assets", h.CreateAsset)

			r.Get("/portfolios", h.ListPortfolios)
			r.Post("/portfolios", h.CreatePortfolio)
			r.Get("/portfolios/{id}", h.GetPortfolio)
			r.Patch("/portfolios/{id}", h.UpdatePortfolio)
			r.Delete("/portfolios/{id}", h.DeletePortfolio)

			r.Get("/portfolios/{id}/transactions", h.ListTransactions)
			r.Post("/portfolios/{id}/transactions", h.CreateTransaction)

			r.Get("/portfolios/{id}/summary", h.GetPortfolioSummary)
			r.Get("/portfolios/{id}/performance", h.GetPortfolioPerformance)
			r.Get("/portfolios/{id}/allocation", h.GetPortfolioAllocation)
			r.Get("/portfolios/{id}/roi", h.GetPortfolioROI)

			r.Get("/prices/{assetID}", h.GetPrices)
			r.Post("/prices/refresh", h.RefreshPrices)
		})
	})
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
