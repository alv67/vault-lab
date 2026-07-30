package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amelamela/vault-lab/internal/model"
)

type AssetRepository interface {
	Create(ctx context.Context, asset *model.Asset) (*model.Asset, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error)
	Search(ctx context.Context, query string) ([]*model.Asset, error)
	List(ctx context.Context) ([]*model.Asset, error)
}

type assetRepo struct {
	db *pgxpool.Pool
}

func (r *assetRepo) Create(ctx context.Context, asset *model.Asset) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO assets (ticker, isin, name, type, category_id, country, currency)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, ticker, isin, name, type, category_id, country, currency, created_at`,
		asset.Ticker, asset.ISIN, asset.Name, asset.Type, asset.CategoryID, asset.Country, asset.Currency,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Asset, error) {
	a := &model.Asset{}
	err := r.db.QueryRow(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at FROM assets WHERE id = $1`,
		id,
	).Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (r *assetRepo) Search(ctx context.Context, query string) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at
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
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}

func (r *assetRepo) List(ctx context.Context) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, ticker, isin, name, type, category_id, country, currency, created_at FROM assets ORDER BY ticker`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []*model.Asset
	for rows.Next() {
		a := &model.Asset{}
		if err := rows.Scan(&a.ID, &a.Ticker, &a.ISIN, &a.Name, &a.Type, &a.CategoryID, &a.Country, &a.Currency, &a.CreatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, nil
}
