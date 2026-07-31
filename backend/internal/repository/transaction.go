package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/amelamela/vault-lab/internal/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) (*model.Transaction, error)
	FindByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]model.TransactionWithAsset, error)
	CountByAsset(ctx context.Context, assetID uuid.UUID) (int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	Update(ctx context.Context, tx *model.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type transactionRepo struct {
	db *pgxpool.Pool
}

func (r *transactionRepo) Create(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	t := &model.Transaction{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO transactions (portfolio_id, asset_id, type, quantity, price, currency, exchange_rate, fees, date, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		 RETURNING id, portfolio_id, asset_id, type, quantity, price, currency, exchange_rate, fees, date, notes, created_at`,
		tx.PortfolioID, tx.AssetID, tx.Type, tx.Quantity, tx.Price, tx.Currency, tx.ExchangeRate, tx.Fees, tx.Date, tx.Notes,
	).Scan(&t.ID, &t.PortfolioID, &t.AssetID, &t.Type, &t.Quantity, &t.Price, &t.Currency, &t.ExchangeRate, &t.Fees, &t.Date, &t.Notes, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *transactionRepo) CountByAsset(ctx context.Context, assetID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE asset_id = $1`,
		assetID,
	).Scan(&count)
	return count, err
}

func (r *transactionRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	t := &model.Transaction{}
	err := r.db.QueryRow(ctx,
		`SELECT id, portfolio_id, asset_id, type, quantity, price, currency, exchange_rate, fees, date, notes, created_at
		 FROM transactions WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.PortfolioID, &t.AssetID, &t.Type, &t.Quantity, &t.Price, &t.Currency, &t.ExchangeRate, &t.Fees, &t.Date, &t.Notes, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *transactionRepo) Update(ctx context.Context, tx *model.Transaction) error {
	_, err := r.db.Exec(ctx,
		`UPDATE transactions
		 SET asset_id = $1, type = $2, quantity = $3, price = $4, currency = $5,
		     exchange_rate = $6, fees = $7, date = $8, notes = $9
		 WHERE id = $10`,
		tx.AssetID, tx.Type, tx.Quantity, tx.Price, tx.Currency, tx.ExchangeRate, tx.Fees, tx.Date, tx.Notes, tx.ID,
	)
	return err
}

func (r *transactionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	return err
}

func (r *transactionRepo) FindByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]model.TransactionWithAsset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.portfolio_id, t.asset_id, a.ticker, a.name, t.type,
		        t.quantity, t.price, t.currency, t.exchange_rate, t.fees, t.date, t.notes, t.created_at
		 FROM transactions t
		 JOIN assets a ON a.id = t.asset_id
		 WHERE t.portfolio_id = $1
		 ORDER BY t.date DESC`,
		portfolioID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []model.TransactionWithAsset
	for rows.Next() {
		var tx model.TransactionWithAsset
		if err := rows.Scan(
			&tx.ID, &tx.PortfolioID, &tx.AssetID,
			&tx.AssetTicker, &tx.AssetName, &tx.Type,
			&tx.Quantity, &tx.Price, &tx.Currency,
			&tx.ExchangeRate, &tx.Fees, &tx.Date, &tx.Notes, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
