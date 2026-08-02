package repository

import (
	"context"
	"time"
)

type LookupRepository interface {
	Get(ctx context.Context, query string, maxAge time.Duration) ([]byte, error)
	Set(ctx context.Context, query string, results []byte) error
}

type lookupRepo struct {
	db DBTX
}

func (r *lookupRepo) Get(ctx context.Context, query string, maxAge time.Duration) ([]byte, error) {
	var results []byte
	err := r.db.QueryRow(ctx,
		`SELECT results FROM lookup_cache WHERE query = $1 AND created_at >= $2`,
		query, time.Now().Add(-maxAge),
	).Scan(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *lookupRepo) Set(ctx context.Context, query string, results []byte) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO lookup_cache (query, results) VALUES ($1, $2)
		 ON CONFLICT (query) DO UPDATE SET results = EXCLUDED.results, created_at = NOW()`,
		query, results,
	)
	return err
}
