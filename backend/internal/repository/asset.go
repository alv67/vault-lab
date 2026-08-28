package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/amelamela/vault-lab/internal/model"
)

const assetColumns = "id, ticker, isin, name, type, category_id, country, currency, exchange, sector, industry, created_at, price_fetched_at, history_backfilled"

type AssetRepository interface {
	Create(ctx context.Context, asset *model.Asset) (*model.Asset, error)
	Update(ctx context.Context, asset *model.Asset) (*model.Asset, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error)
	FindByTicker(ctx context.Context, ticker string) (*model.Asset, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]*model.Asset, error)
	Search(ctx context.Context, query string) ([]*model.Asset, error)
	List(ctx context.Context) ([]*model.Asset, error)
	MarkPricesFetched(ctx context.Context, ids []uuid.UUID, at time.Time) error
	MarkHistoryBackfilled(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	Currencies(ctx context.Context) ([]string, error)
}

type assetRepo struct {
	db DBTX
}

func (r *assetRepo) Create(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO assets (ticker, isin, name, type, category_id, country, currency, exchange, sector, industry)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING `+assetColumns,
		asset.Ticker, asset.ISIN, asset.Name, asset.Type, asset.CategoryID, asset.Country, asset.Currency,
		asset.Exchange, asset.Sector, asset.Industry,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) Update(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`UPDATE assets SET ticker=$1, isin=$2, name=$3, type=$4, category_id=$5, country=$6, currency=$7, exchange=$8, sector=$9, industry=$10
		 WHERE id=$11
		 RETURNING `+assetColumns,
		asset.Ticker, asset.ISIN, asset.Name, asset.Type, asset.CategoryID, asset.Country, asset.Currency,
		asset.Exchange, asset.Sector, asset.Industry, asset.ID,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`SELECT `+assetColumns+` FROM assets WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) FindByTicker(ctx context.Context, ticker string) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`SELECT `+assetColumns+` FROM assets WHERE ticker = $1`,
		ticker,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled)
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
		`SELECT `+assetColumns+` FROM assets WHERE id = ANY($1::uuid[])`,
		ids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (r *assetRepo) Search(ctx context.Context, query string) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+assetColumns+` FROM assets WHERE ticker ILIKE $1 OR name ILIKE $1 LIMIT 20`,
		"%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

func (r *assetRepo) List(ctx context.Context) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT `+assetColumns+` FROM assets ORDER BY ticker`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.Exchange, &a.Sector, &a.Industry, &a.CreatedAt, &a.PriceFetchedAt, &a.HistoryBackfilled); err != nil {
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

func (r *assetRepo) MarkHistoryBackfilled(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE assets SET history_backfilled = TRUE WHERE id = $1`, id)
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
