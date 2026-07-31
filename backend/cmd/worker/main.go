package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/amelamela/vault-lab/internal/config"
	"github.com/amelamela/vault-lab/internal/price"
	"github.com/amelamela/vault-lab/internal/repository"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}
	defer dbPool.Close()

	repos := repository.New(dbPool)
	fetcher := price.NewYahooFetcher(repos, cfg.PriceFetchInterval)

	log.Info().Dur("interval", cfg.PriceFetchInterval).Msg("price worker started")

	ticker := time.NewTicker(cfg.PriceFetchInterval)
	defer ticker.Stop()

	// Run once immediately
	if err := fetcher.FetchAll(ctx); err != nil {
		log.Warn().Err(err).Msg("initial price fetch failed")
	}

	for {
		select {
		case <-ticker.C:
			log.Info().Msg("fetching prices...")
			if err := fetcher.FetchAll(ctx); err != nil {
				log.Warn().Err(err).Msg("price fetch failed")
			}
		case <-ctx.Done():
			log.Info().Msg("worker shutting down")
			return
		case <-func() chan os.Signal {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			return c
		}():
			log.Info().Msg("worker shutting down")
			return
		}
	}
}
