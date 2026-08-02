package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/amelamela/vault-lab/internal/model"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx *model.Transaction) (*model.Transaction, error)
	FindByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]model.TransactionWithAsset, error)
	FindByPortfoliosAsc(ctx context.Context, portfolioIDs []uuid.UUID) ([]model.TransactionWithAsset, error)
	MinDateByAsset(ctx context.Context, assetIDs []uuid.UUID) (map[uuid.UUID]time.Time, error)
	CountByAsset(ctx context.Context, assetID uuid.UUID) (int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error)
	Update(ctx context.Context, tx *model.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type transactionRepo struct {
	db DBTX
}

func (r *transactionRepo) Create(ctx context.Context, tx *model.Transaction) (*model.Transaction, error) {
	t := &model.Transaction{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO transactions (portfolio_id, asset_id, type, quantity, price, fees, date, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, portfolio_id, asset_id, type, quantity, price, fees, date, notes, created_at`,
		tx.PortfolioID, tx.AssetID, tx.Type, tx.Quantity, tx.Price, tx.Fees, tx.Date, tx.Notes,
	).Scan(&t.ID, &t.PortfolioID, &t.AssetID, &t.Type, &t.Quantity, &t.Price, &t.Fees, &t.Date, &t.Notes, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *transactionRepo) MinDateByAsset(ctx context.Context, assetIDs []uuid.UUID) (map[uuid.UUID]time.Time, error) {
	if len(assetIDs) == 0 {
		return map[uuid.UUID]time.Time{}, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT asset_id, MIN(date) FROM transactions WHERE asset_id = ANY($1::uuid[]) GROUP BY asset_id`,
		assetIDs,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := map[uuid.UUID]time.Time{}
	for rows.Next() {
		var id uuid.UUID
		var min time.Time
		if err := rows.Scan(&id, &min); err != nil {
			return nil, err
		}
		result[id] = min
	}
	return result, rows.Err()
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
		`SELECT id, portfolio_id, asset_id, type, quantity, price, fees, date, notes, created_at
		 FROM transactions WHERE id = $1`,
		id,
	).Scan(&t.ID, &t.PortfolioID, &t.AssetID, &t.Type, &t.Quantity, &t.Price, &t.Fees, &t.Date, &t.Notes, &t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *transactionRepo) Update(ctx context.Context, tx *model.Transaction) error {
	_, err := r.db.Exec(ctx,
		`UPDATE transactions
		 SET asset_id = $1, type = $2, quantity = $3, price = $4,
		     fees = $5, date = $6, notes = $7
		 WHERE id = $8`,
		tx.AssetID, tx.Type, tx.Quantity, tx.Price, tx.Fees, tx.Date, tx.Notes, tx.ID,
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
		        t.quantity, t.price, t.fees, t.date, t.notes, t.created_at
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
			&tx.Quantity, &tx.Price, &tx.Fees, &tx.Date, &tx.Notes, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func (r *transactionRepo) FindByPortfoliosAsc(ctx context.Context, portfolioIDs []uuid.UUID) ([]model.TransactionWithAsset, error) {
	rows, err := r.db.Query(ctx,
		`SELECT t.id, t.portfolio_id, t.asset_id, a.ticker, a.name, t.type,
		        t.quantity, t.price, t.fees, t.date, t.notes, t.created_at
		 FROM transactions t
		 JOIN assets a ON a.id = t.asset_id
		 WHERE t.portfolio_id = ANY($1::uuid[])
		 ORDER BY t.date ASC, t.created_at ASC`,
		portfolioIDs,
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
			&tx.Quantity, &tx.Price, &tx.Fees, &tx.Date, &tx.Notes, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
