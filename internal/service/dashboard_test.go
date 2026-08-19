package service

import (
	"context"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
)

type dashboardStoreStub struct{ calls int }

func (s *dashboardStoreStub) Summary(context.Context, string, time.Time) (entity.DashboardSummary, error) {
	s.calls++
	return entity.DashboardSummary{Properties: entity.DashboardPropertyMetrics{Total: 12}}, nil
}

type dashboardCacheStub struct {
	value entity.DashboardSummary
	found bool
	sets  int
}

func (c *dashboardCacheStub) Get(context.Context, string) (entity.DashboardSummary, bool, error) {
	return c.value, c.found, nil
}
func (c *dashboardCacheStub) Set(_ context.Context, _ string, value entity.DashboardSummary, _ time.Duration) error {
	c.value = value
	c.found = true
	c.sets++
	return nil
}

func TestDashboardServiceCachesSummary(t *testing.T) {
	store := &dashboardStoreStub{}
	cache := &dashboardCacheStub{}
	service := DashboardService{Repo: store, Cache: cache, Now: func() time.Time { return time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC) }}
	first, err := service.Summary(context.Background(), "user-1")
	if err != nil || first.Properties.Total != 12 {
		t.Fatalf("unexpected first result: %#v %v", first, err)
	}
	second, err := service.Summary(context.Background(), "user-1")
	if err != nil || second.Properties.Total != 12 {
		t.Fatalf("unexpected cached result: %#v %v", second, err)
	}
	if store.calls != 1 || cache.sets != 1 {
		t.Fatalf("expected one database call and one cache set, got %d and %d", store.calls, cache.sets)
	}
}
