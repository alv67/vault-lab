package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	DB          *pgxpool.Pool
	User        UserRepository
	Asset       AssetRepository
	Portfolio   PortfolioRepository
	Transaction TransactionRepository
	Price       PriceRepository
	Lookup      LookupRepository
	FX          FXRepository
	Split       SplitRepository
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{
		DB:          db,
		User:        &userRepo{db},
		Asset:       &assetRepo{db},
		Portfolio:   &portfolioRepo{db: db, transactions: &transactionRepo{db}, splits: &splitRepo{db}},
		Transaction: &transactionRepo{db},
		Price:       &priceRepo{db},
		Lookup:      &lookupRepo{db},
		FX:          &fxRepo{db},
		Split:       &splitRepo{db},
	}
}

func (r *Repository) Ping(ctx context.Context) error {
	return r.DB.Ping(ctx)
}
