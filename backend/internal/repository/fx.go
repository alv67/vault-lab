package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type FXRepository interface {
	Upsert(ctx context.Context, base, quote string, rate decimal.Decimal) error
	LatestByQuotes(ctx context.Context, quotes []string) (map[string]decimal.Decimal, error)
	FetchedAt(ctx context.Context, quote string) (*time.Time, error)
}

type fxRepo struct {
	db *pgxpool.Pool
}

func (r *fxRepo) Upsert(ctx context.Context, base, quote string, rate decimal.Decimal) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO fx_rates (base_currency, quote_currency, rate, fetched_at)
		 VALUES ($1, $2, $3, NOW())
		 ON CONFLICT (base_currency, quote_currency)
		 DO UPDATE SET rate = EXCLUDED.rate, fetched_at = NOW()`,
		base, quote, rate,
	)
	return err
}

func (r *fxRepo) LatestByQuotes(ctx context.Context, quotes []string) (map[string]decimal.Decimal, error) {
	if len(quotes) == 0 {
		return map[string]decimal.Decimal{}, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT quote_currency, rate FROM fx_rates WHERE base_currency = 'USD' AND quote_currency = ANY($1::text[])`,
		quotes,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rates := make(map[string]decimal.Decimal)
	for rows.Next() {
		var quote string
		var rate decimal.Decimal
		if err := rows.Scan(&quote, &rate); err != nil {
			return nil, err
		}
		rates[quote] = rate
	}
	return rates, rows.Err()
}

func (r *fxRepo) FetchedAt(ctx context.Context, quote string) (*time.Time, error) {
	var at time.Time
	err := r.db.QueryRow(ctx,
		`SELECT fetched_at FROM fx_rates WHERE base_currency = 'USD' AND quote_currency = $1`,
		quote,
	).Scan(&at)
	if err != nil {
		return nil, err
	}
	return &at, nil
}
