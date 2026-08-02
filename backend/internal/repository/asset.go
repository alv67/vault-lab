package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/amelamela/vault-lab/internal/model"
)

type AssetRepository interface {
	Create(ctx context.Context, asset *model.Asset) (*model.Asset, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error)
	FindByTicker(ctx context.Context, ticker string) (*model.Asset, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Asset, error)
	Search(ctx context.Context, query string) ([]*model.Asset, error)
	List(ctx context.Context) ([]*model.Asset, error)
	MarkPricesFetched(ctx context.Context, ids []uuid.UUID, at time.Time) error
	Delete(ctx context.Context, id uuid.UUID) error
	Currencies(ctx context.Context) ([]string, error)
}

type assetRepo struct {
	db DBTX
}

func (r *assetRepo) Create(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO assets (ticker, isin, name, type, category_id, country, currency)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, ticker, isin, name, type, category_id, country, currency, created_at, price_fetched_at`,
		asset.Ticker, asset.ISIN, asset.Name, asset.Type, asset.CategoryID, asset.Country, asset.Currency,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt, &a.PriceFetchedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at, price_fetched_at FROM assets WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt, &a.PriceFetchedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) FindByTicker(ctx context.Context, ticker string) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at, price_fetched_at
		 FROM assets WHERE ticker = $1`,
		ticker,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt, &a.PriceFetchedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Asset, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at, price_fetched_at
		 FROM assets WHERE id = ANY($1::uuid[])`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt, &a.PriceFetchedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (r *assetRepo) Search(ctx context.Context, query string) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at, price_fetched_at
		 FROM assets WHERE ticker ILIKE $1 OR name ILIKE $1 LIMIT 20`,
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt, &a.PriceFetchedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

func (r *assetRepo) List(ctx context.Context) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at, price_fetched_at FROM assets ORDER BY ticker`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt, &a.PriceFetchedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

func (r *assetRepo) MarkPricesFetched(ctx context.Context, ids []uuid.UUID, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	_, err := r.db.Exec(ctx,
		`UPDATE assets SET price_fetched_at = $1 WHERE id = ANY($2::uuid[])`,
		at, ids,
	)
	return err
}

func (r *assetRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM assets WHERE id = $1`, id)
	return err
}

func (r *assetRepo) Currencies(ctx context.Context) ([]string, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT currency FROM assets WHERE currency <> '' ORDER BY currency`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}
	return currencies, rows.Err()
}
