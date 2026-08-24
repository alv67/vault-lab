package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/amelamela/vault-lab/internal/model"
)

type HealthRepository interface {
	RecordEvent(ctx context.Context, event *model.HealthEvent) error
	GetLatestEvents(ctx context.Context, limit int) ([]*model.HealthEvent, error)
}

type healthRepo struct {
	db *pgxpool.Pool
}

func NewHealthRepository(db *pgxpool.Pool) HealthRepository {
	return &healthRepo{db}
}

func (r *healthRepo) RecordEvent(ctx context.Context, event *model.HealthEvent) error {
	_, err := r.db.Exec(ctx, 
		`INSERT INTO health_events (id, asset_id, event_type, status, code, message, duration_ms, error_code, created_at) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		event.ID, event.AssetID, event.EventType, event.Status, event.Code, event.Message, event.DurationMs, event.ErrorCode, event.CreatedAt)
	return err
}

func (r *healthRepo) GetLatestEvents(ctx context.Context, limit int) ([]*model.HealthEvent, error) {
	rows, err := r.db.Query(ctx, 
		`SELECT id, asset_id, event_type, status, code, message, duration_ms, error_code, created_at 
		 FROM health_events 
		 ORDER BY created_at DESC 
		 LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*model.HealthEvent
	for rows.Next() {
		e := &model.HealthEvent{}
		var assetID *uuid.UUID
		err := rows.Scan(&e.ID, &assetID, &e.EventType, &e.Status, &e.Code, &e.Message, &e.DurationMs, &e.ErrorCode, &e.CreatedAt)
		if err != nil {
			return nil, err
		}
		e.AssetID = assetID
		events = append(events, e)
	}
	return events, nil
}
