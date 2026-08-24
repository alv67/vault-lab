package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/amelamela/vault-lab/internal/model"
	"github.com/amelamela/vault-lab/internal/repository"
)

type HealthService struct {
	repos  *repository.Repository
	redis  *redis.Client
}

func NewHealthService(repos *repository.Repository, redis *redis.Client) *HealthService {
	return &HealthService{repos: repos, redis: redis}
}

func (s *HealthService) RecordEvent(ctx context.Context, event *model.HealthEvent) error {
	// 1. Persist to DB
	if err := s.repos.Health.RecordEvent(ctx, event); err != nil {
		return err
	}

	// 2. Update Hybrid Cache in Redis (snapshot for fast access)
	// Store summary for the last hour
	key := "health:summary:hourly"
	if event.Status == "success" {
		s.redis.Incr(ctx, key+":success")
	} else {
		s.redis.Incr(ctx, key+":failure")
		if event.Code == "rate_limited" {
			s.redis.Incr(ctx, key+":rate_limited")
		}
	}
	s.redis.Expire(ctx, key+":success", time.Hour)
	s.redis.Expire(ctx, key+":failure", time.Hour)
	s.redis.Expire(ctx, key+":rate_limited", time.Hour)

	return nil
}

func (s *HealthService) GetPriceHealth(ctx context.Context) (*model.HealthSummary, []*model.HealthEvent, error) {
	// Use Redis for summary if available
	summary := &model.HealthSummary{}
	
	successes, _ := s.redis.Get(ctx, "health:summary:hourly:success").Int64()
	failures, _ := s.redis.Get(ctx, "health:summary:hourly:failure").Int64()
	rateLimited, _ := s.redis.Get(ctx, "health:summary:hourly:rate_limited").Int64()

	summary.Successes = int(successes)
	summary.Failures = int(failures)
	summary.RateLimited = int(rateLimited)
	
	total := summary.Successes + summary.Failures
	if total > 0 {
		summary.SuccessRate = float64(summary.Successes) / float64(total)
	}

	events, err := s.repos.Health.GetLatestEvents(ctx, 100)
	if err != nil {
		return nil, nil, err
	}

	return summary, events, nil
}
