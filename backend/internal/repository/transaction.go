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
