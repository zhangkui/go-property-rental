package service

import (
	"context"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/repository"
)

type DashboardService struct {
	Repo     repository.DashboardStore
	Cache    repository.DashboardCache
	CacheTTL time.Duration
	Now      func() time.Time
}

func (s DashboardService) Summary(ctx context.Context, userID string) (entity.DashboardSummary, error) {
	cacheKey := "summary:" + userID
	if s.Cache != nil {
		summary, found, err := s.Cache.Get(ctx, cacheKey)
		if err == nil && found {
			return summary, nil
		}
		// A corrupt cache entry returns found=true with an error; fall back to the
		// database rather than returning the bad value, then overwrite it below.
	}
	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	summary, err := s.Repo.Summary(ctx, userID, now)
	if err != nil {
		return entity.DashboardSummary{}, err
	}
	ttl := s.CacheTTL
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	if s.Cache != nil {
		_ = s.Cache.Set(ctx, cacheKey, summary, ttl)
	}
	return summary, nil
}
