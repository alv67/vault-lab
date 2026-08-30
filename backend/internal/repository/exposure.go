package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/amelamela/vault-lab/internal/model"
)

type ExposureRepository interface {
	FindRegions(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error)
	FindSectors(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error)
	ReplaceRegions(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error
	ReplaceSectors(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error
	FindRegionsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error)
	FindSectorsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error)
}

type exposureRepo struct {
	db DBTX
}

func (r *exposureRepo) FindRegions(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT region, weight FROM asset_region_weights WHERE asset_id = $1 ORDER BY region`,
		assetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ExposureRow
	for rows.Next() {
		var row model.ExposureRow
		if err := rows.Scan(&row.Name, &row.Weight); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *exposureRepo) FindSectors(ctx context.Context, assetID uuid.UUID) ([]model.ExposureRow, error) {
	rows, err := r.db.Query(ctx,
		`SELECT sector, weight FROM asset_sector_weights WHERE asset_id = $1 ORDER BY sector`,
		assetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.ExposureRow
	for rows.Next() {
		var row model.ExposureRow
		if err := rows.Scan(&row.Name, &row.Weight); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *exposureRepo) FindRegionsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return r.findByAssets(ctx, assetIDs, "asset_region_weights", "region")
}

func (r *exposureRepo) FindSectorsByAssets(ctx context.Context, assetIDs []uuid.UUID) (map[string][]model.ExposureRow, error) {
	return r.findByAssets(ctx, assetIDs, "asset_sector_weights", "sector")
}

// findByAssets fetches the stored exposure rows for a set of assets in a single
// batch query, keyed by asset id. Rows with an empty name or a non-positive
// weight are skipped, mirroring replace's semantics. Table and column are
// internal constants, so the formatted names are safe.
func (r *exposureRepo) findByAssets(ctx context.Context, assetIDs []uuid.UUID, table, column string) (map[string][]model.ExposureRow, error) {
	if len(assetIDs) == 0 {
		return map[string][]model.ExposureRow{}, nil
	}
	query := fmt.Sprintf(`SELECT asset_id, %s, weight FROM %s WHERE asset_id = ANY($1::uuid[]) ORDER BY asset_id, %s`, column, table, column)
	rows, err := r.db.Query(ctx, query, assetIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[string][]model.ExposureRow{}
	for rows.Next() {
		var assetID uuid.UUID
		var row model.ExposureRow
		if err := rows.Scan(&assetID, &row.Name, &row.Weight); err != nil {
			return nil, err
		}
		if row.Name == "" || !row.Weight.IsPositive() {
			continue
		}
		key := assetID.String()
		out[key] = append(out[key], row)
	}
	return out, rows.Err()
}

func (r *exposureRepo) ReplaceRegions(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return r.replace(ctx, assetID, rows, "asset_region_weights", "region")
}

func (r *exposureRepo) ReplaceSectors(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow) error {
	return r.replace(ctx, assetID, rows, "asset_sector_weights", "sector")
}

// replace deletes the stored rows for an asset and inserts the given ones,
// skipping rows with an empty name or a non-positive weight. Table and column
// are internal constants, so the formatted names are safe.
func (r *exposureRepo) replace(ctx context.Context, assetID uuid.UUID, rows []model.ExposureRow, table, column string) error {
	if _, err := r.db.Exec(ctx, fmt.Sprintf(`DELETE FROM %s WHERE asset_id = $1`, table), assetID); err != nil {
		return err
	}
	for _, row := range rows {
		if row.Name == "" || !row.Weight.IsPositive() {
			continue
		}
		if _, err := r.db.Exec(ctx,
			fmt.Sprintf(`INSERT INTO %s (asset_id, %s, weight) VALUES ($1, $2, $3)`, table, column),
			assetID, row.Name, row.Weight,
		); err != nil {
			return err
		}
	}
	return nil
}
