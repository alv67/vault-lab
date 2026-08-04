package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/position"
)

type PortfolioRepository interface {
	Create(ctx context.Context, p *model.Portfolio) (*model.Portfolio, error)
	FindByID(ctx context.Context, id uuid.UUID) (*model.Portfolio, error)
	FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.Portfolio, error)
	Update(ctx context.Context, p *model.Portfolio) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetSummary(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioSummary, error)
	GetAllocation(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetAllocation, error)
	GetPerformance(ctx context.Context, portfolioID uuid.UUID) ([]*model.PortfolioPerformance, error)
	GetROI(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetROI, error)
	HeldAssets(ctx context.Context, portfolioID uuid.UUID) ([]*model.Asset, error)
	HoldingsDetailed(ctx context.Context, portfolioIDs []uuid.UUID) ([]*model.Holding, error)
	FindAll(ctx context.Context) ([]uuid.UUID, error)
}

type portfolioRepo struct {
	db           DBTX
	transactions TransactionRepository
	splits       SplitRepository
}

func (r *portfolioRepo) Create(ctx context.Context, p *model.Portfolio) (*model.Portfolio, error) {
	portfolio := &model.Portfolio{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO portfolios (user_id, name, description, currency)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, user_id, name, description, currency, created_at, updated_at`,
		p.UserID, p.Name, p.Description, p.Currency,
	).Scan(&portfolio.ID, &portfolio.UserID, &portfolio.Name, &portfolio.Description, &portfolio.Currency, &portfolio.CreatedAt, &portfolio.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (r *portfolioRepo) FindByID(ctx context.Context, id uuid.UUID) (*model.Portfolio, error) {
	p := &model.Portfolio{}
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, name, description, currency, created_at, updated_at FROM portfolios WHERE id = $1`,
		id,
	).Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Currency, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *portfolioRepo) FindByUser(ctx context.Context, userID uuid.UUID) ([]*model.Portfolio, error) {
	rows, err := r.db.Query(ctx,
		`SELECT p.id, p.user_id, p.name, p.description, p.currency, p.created_at, p.updated_at
		 FROM portfolios p
		 LEFT JOIN portfolio_shares ps ON ps.portfolio_id = p.id
		 WHERE p.user_id = $1 OR ps.user_id = $1
		 ORDER BY p.created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var portfolios []*model.Portfolio
	for rows.Next() {
		p := &model.Portfolio{}
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Currency, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		portfolios = append(portfolios, p)
	}
	return portfolios, nil
}

// FindAll returns the ids of every portfolio in the system.
func (r *portfolioRepo) FindAll(ctx context.Context) ([]uuid.UUID, error) {
	rows, err := r.db.Query(ctx, `SELECT id FROM portfolios`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (r *portfolioRepo) Update(ctx context.Context, p *model.Portfolio) error {
	_, err := r.db.Exec(ctx,
		`UPDATE portfolios SET name = $1, description = $2, currency = $3, updated_at = NOW() WHERE id = $4`,
		p.Name, p.Description, p.Currency, p.ID,
	)
	return err
}

func (r *portfolioRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM portfolios WHERE id = $1`, id)
	return err
}

func (r *portfolioRepo) GetSummary(ctx context.Context, portfolioID uuid.UUID) (*model.PortfolioSummary, error) {
	s := &model.PortfolioSummary{}
	err := r.db.QueryRow(ctx, `
		WITH
		buy AS (
			SELECT asset_id,
				SUM(quantity * price + fees) AS total_cost,
				SUM(quantity) AS total_qty
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'buy'
			GROUP BY asset_id
		),
		sell AS (
			SELECT asset_id,
				SUM(quantity) AS sold_qty,
				SUM(quantity * price - fees) AS total_proceeds
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'sell'
			GROUP BY asset_id
		),
		holdings AS (
			SELECT b.asset_id,
				(b.total_qty - COALESCE(s.sold_qty, 0)) AS qty,
				b.total_cost - COALESCE(s.total_proceeds, 0) AS cost_basis
			FROM buy b
			LEFT JOIN sell s ON s.asset_id = b.asset_id
		),
		latest_prices AS (
			SELECT DISTINCT ON (asset_id) asset_id, close
			FROM prices
			WHERE asset_id IN (SELECT asset_id FROM holdings WHERE qty > 0)
			ORDER BY asset_id, date DESC
		)
		SELECT
			COUNT(*) AS asset_count,
			COALESCE(SUM(lp.close * h.qty), 0) AS total_value,
			COALESCE(SUM(h.cost_basis), 0) AS total_cost
		FROM holdings h
		LEFT JOIN latest_prices lp ON lp.asset_id = h.asset_id
		WHERE h.qty > 0
	`, portfolioID).Scan(&s.AssetCount, &s.TotalValue, &s.TotalCost)

	if err != nil {
		return nil, err
	}
	s.GainLoss = s.TotalValue.Sub(s.TotalCost)
	if s.TotalCost.IsPositive() {
		s.GainLossPct = s.GainLoss.Div(s.TotalCost).Mul(decimal.NewFromInt(100))
	}
	return s, nil
}

func (r *portfolioRepo) GetAllocation(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetAllocation, error) {
	rows, err := r.db.Query(ctx, `
		WITH
		buy AS (
			SELECT asset_id,
				SUM(quantity * price + fees) AS total_cost,
				SUM(quantity) AS total_qty
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'buy'
			GROUP BY asset_id
		),
		sell AS (
			SELECT asset_id,
				SUM(quantity) AS sold_qty,
				SUM(quantity * price - fees) AS total_proceeds
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'sell'
			GROUP BY asset_id
		),
		holdings AS (
			SELECT b.asset_id,
				(b.total_qty - COALESCE(s.sold_qty, 0)) AS qty
			FROM buy b
			LEFT JOIN sell s ON s.asset_id = b.asset_id
		),
		latest_prices AS (
			SELECT DISTINCT ON (asset_id) asset_id, close
			FROM prices
			WHERE asset_id IN (SELECT asset_id FROM holdings WHERE qty > 0)
			ORDER BY asset_id, date DESC
		),
		valuations AS (
			SELECT h.asset_id, a.ticker, a.name, COALESCE(lp.close * h.qty, 0) AS value
			FROM holdings h
			LEFT JOIN latest_prices lp ON lp.asset_id = h.asset_id
			JOIN assets a ON a.id = h.asset_id
			WHERE h.qty > 0
		)
		SELECT asset_id, ticker, name, value,
			value * 100.0 / SUM(value) OVER () AS alloc_pct
		FROM valuations
		ORDER BY value DESC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocs []*model.AssetAllocation
	for rows.Next() {
		a := &model.AssetAllocation{}
		if err := rows.Scan(&a.AssetID, &a.Ticker, &a.Name, &a.Value, &a.AllocPct); err != nil {
			return nil, err
		}
		allocs = append(allocs, a)
	}
	return allocs, nil
}

func (r *portfolioRepo) GetPerformance(ctx context.Context, portfolioID uuid.UUID) ([]*model.PortfolioPerformance, error) {
	rows, err := r.db.Query(ctx, `
		WITH
		buy AS (
			SELECT asset_id, date, quantity, price
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'buy'
		),
		sell AS (
			SELECT asset_id, date, quantity
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'sell'
		),
		dates AS (
			SELECT DISTINCT p.date
			FROM prices p
			WHERE p.asset_id IN (SELECT asset_id FROM buy)
			AND p.date >= (SELECT MIN(date) FROM buy)
		),
		holdings_at_date AS (
			SELECT d.date, b.asset_id, b.quantity - COALESCE(s.sold_qty, 0) AS qty
			FROM dates d
			JOIN buy b ON b.date <= d.date
			LEFT JOIN (
				SELECT asset_id, date, SUM(quantity) AS sold_qty
				FROM sell
				GROUP BY asset_id, date
			) s ON s.asset_id = b.asset_id AND s.date <= d.date
		)
		SELECT d.date,
			COALESCE(SUM(h.qty * p.close), 0) AS value
		FROM dates d
		LEFT JOIN holdings_at_date h ON h.date = d.date
		LEFT JOIN prices p ON p.asset_id = h.asset_id AND p.date = d.date
		GROUP BY d.date
		ORDER BY d.date
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perf []*model.PortfolioPerformance
	for rows.Next() {
		p := &model.PortfolioPerformance{}
		if err := rows.Scan(&p.Date, &p.Value); err != nil {
			return nil, err
		}
		perf = append(perf, p)
	}
	return perf, nil
}

func (r *portfolioRepo) GetROI(ctx context.Context, portfolioID uuid.UUID) ([]*model.AssetROI, error) {
	rows, err := r.db.Query(ctx, `
		WITH
		buy AS (
			SELECT asset_id,
				SUM(quantity * price + fees) AS total_cost,
				SUM(quantity) AS total_qty
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'buy'
			GROUP BY asset_id
		),
		sell AS (
			SELECT asset_id,
				SUM(quantity) AS sold_qty,
				SUM(quantity * price - fees) AS total_proceeds
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'sell'
			GROUP BY asset_id
		),
		holdings AS (
			SELECT b.asset_id,
				(b.total_qty - COALESCE(s.sold_qty, 0)) AS qty,
				b.total_cost - COALESCE(s.total_proceeds, 0) AS cost_basis,
				b.total_cost
			FROM buy b
			LEFT JOIN sell s ON s.asset_id = b.asset_id
		),
		latest_prices AS (
			SELECT DISTINCT ON (asset_id) asset_id, close
			FROM prices
			WHERE asset_id IN (SELECT asset_id FROM holdings WHERE qty > 0)
			ORDER BY asset_id, date DESC
		)
		SELECT h.asset_id, a.ticker, a.name,
			COALESCE(lp.close * h.qty, 0) AS current_value,
			h.cost_basis AS total_invested
		FROM holdings h
		LEFT JOIN latest_prices lp ON lp.asset_id = h.asset_id
		JOIN assets a ON a.id = h.asset_id
		WHERE h.qty > 0
		ORDER BY current_value DESC
	`, portfolioID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rois []*model.AssetROI
	for rows.Next() {
		r := &model.AssetROI{}
		if err := rows.Scan(&r.AssetID, &r.Ticker, &r.Name, &r.CurrentValue, &r.TotalInvested); err != nil {
			return nil, err
		}
		if r.TotalInvested.IsPositive() {
			r.ROI = r.CurrentValue.Sub(r.TotalInvested).Div(r.TotalInvested).Mul(decimal.NewFromInt(100))
		}
		rois = append(rois, r)
	}
	return rois, nil
}

func (r *portfolioRepo) HeldAssets(ctx context.Context, portfolioID uuid.UUID) ([]*model.Asset, error) {
	rows, err := r.db.Query(ctx, `
		WITH buy AS (
			SELECT asset_id, SUM(quantity) AS qty
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'buy'
			GROUP BY asset_id
		),
		sell AS (
			SELECT asset_id, SUM(quantity) AS qty
			FROM transactions
			WHERE portfolio_id = $1 AND type = 'sell'
			GROUP BY asset_id
		)
		SELECT a.id, a.ticker, a.isin, a.name, a.type, a.category_id, a.country, a.currency, a.created_at, a.price_fetched_at
		FROM buy b
		LEFT JOIN sell s ON s.asset_id = b.asset_id
		JOIN assets a ON a.id = b.asset_id
		WHERE b.qty - COALESCE(s.qty, 0) > 0
	`, portfolioID)
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

func (r *portfolioRepo) HoldingsDetailed(ctx context.Context, portfolioIDs []uuid.UUID) ([]*model.Holding, error) {
	if len(portfolioIDs) == 0 {
		return nil, nil
	}
	rows, err := r.db.Query(ctx, `
		SELECT DISTINCT t.portfolio_id, t.asset_id, a.ticker, a.name, a.currency,
			COALESCE(p.close, 0) AS last_close, (p.close IS NOT NULL) AS has_price
		FROM transactions t
		JOIN assets a ON a.id = t.asset_id
		LEFT JOIN LATERAL (
			SELECT close FROM prices WHERE asset_id = t.asset_id ORDER BY date DESC LIMIT 1
		) p ON TRUE
		WHERE t.portfolio_id = ANY($1::uuid[])
		ORDER BY t.portfolio_id, a.ticker`, portfolioIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	base := make([]*model.Holding, 0)
	for rows.Next() {
		h := &model.Holding{}
		if err := rows.Scan(&h.PortfolioID, &h.AssetID, &h.Ticker, &h.Name, &h.Currency, &h.LastClose, &h.HasPrice); err != nil {
			return nil, err
		}
		base = append(base, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	txs, err := r.transactions.FindByPortfoliosAsc(ctx, portfolioIDs)
	if err != nil {
		return nil, err
	}

	assetIDs := make([]uuid.UUID, 0, len(txs))
	seen := map[uuid.UUID]bool{}
	for _, tx := range txs {
		if !seen[tx.AssetID] {
			seen[tx.AssetID] = true
			assetIDs = append(assetIDs, tx.AssetID)
		}
	}
	splitRows, err := r.splits.FindByAssets(ctx, assetIDs)
	if err != nil {
		return nil, err
	}
	splitEvents := make([]position.SplitEvent, 0, len(splitRows))
	for _, sp := range splitRows {
		splitEvents = append(splitEvents, position.SplitEvent{
			AssetID: sp.AssetID,
			Date:    sp.Date,
			Ratio:   sp.Numerator.Div(sp.Denominator),
		})
	}
	states := position.Walk(txs, splitEvents)

	for _, h := range base {
		if st, ok := states[h.PortfolioID+"|"+h.AssetID]; ok {
			h.Qty = st.Qty
			h.Cost = st.Cost
			h.CostCCY = st.CostCCY
			h.Realized = st.Realized
			h.RealizedCCY = st.RealizedCCY
			h.AvgCost = st.Avg
		}
	}
	return base, nil
}
