package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amelamela/vault-lab/internal/model"
)

type PriceRepository interface {
	Create(ctx context.Context, price *model.Price) (*model.Price, error)
	FindByAsset(ctx context.Context, assetID uuid.UUID) ([]*model.Price, error)
	FindLatest(ctx context.Context, assetID uuid.UUID) (*model.Price, error)
}

type priceRepo struct {
	db *pgxpool.Pool
}

func (r *priceRepo) Create(ctx context.Context, price *model.Price) (*model.Price, error) {
	p := &model.Price{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO prices (asset_id, date, open, high, low, close, volume, source)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (asset_id, date) DO UPDATE SET close = $6, source = $8
		 RETURNING id, asset_id, date, open, high, low, close, volume, source, created_at`,
		price.AssetID, price.Date, price.Open, price.High, price.Low, price.Close, price.Volume, price.Source,
	).Scan(&p.ID, &p.AssetID, &p.Date, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume, &p.Source, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *priceRepo) FindByAsset(ctx context.Context, assetID uuid.UUID) ([]*model.Price, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, asset_id, date, open, high, low, close, volume, source, created_at
		 FROM prices WHERE asset_id = $1 ORDER BY date DESC LIMIT 365`,
		assetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*model.Price
	for rows.Next() {
		p := &model.Price{}
		if err := rows.Scan(&p.ID, &p.AssetID, &p.Date, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume, &p.Source, &p.CreatedAt); err != nil {
			return nil, err
		}
		prices = append(prices, p)
	}
	return prices, nil
}

func (r *priceRepo) FindLatest(ctx context.Context, assetID uuid.UUID) (*model.Price, error) {
	p := &model.Price{}
	err := r.db.QueryRow(ctx,
		`SELECT id, asset_id, date, open, high, low, close, volume, source, created_at
		 FROM prices WHERE asset_id = $1 ORDER BY date DESC LIMIT 1`,
		assetID,
	).Scan(&p.ID, &p.AssetID, &p.Date, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume, &p.Source, &p.CreatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Ensure unused import compiles
var _ time.Time
