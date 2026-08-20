package verify

import (
	"context"
	"errors"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
	"time"
)

type bug027Repo struct {
	calls   int
	summary entity.DashboardSummary
}

func (s *bug027Repo) Summary(context.Context, string, time.Time) (entity.DashboardSummary, error) {
	s.calls++
	return s.summary, nil
}

type bug027Cache struct{}

func (*bug027Cache) Get(context.Context, string) (entity.DashboardSummary, bool, error) {
	return entity.DashboardSummary{}, true, errors.New("corrupt cache")
}
func (*bug027Cache) Set(context.Context, string, entity.DashboardSummary, time.Duration) error {
	return nil
}
func TestBug027_BusinessRegression(t *testing.T) {
	repo := &bug027Repo{summary: entity.DashboardSummary{GeneratedAt: time.Date(2026, 8, 19, 0, 0, 0, 0, time.UTC)}}
	result, err := (service.DashboardService{Repo: repo, Cache: &bug027Cache{}}).Summary(context.Background(), "user-1")
	if err != nil {
		t.Fatal(err)
	}
	if repo.calls != 1 || result.GeneratedAt.IsZero() {
		t.Fatalf("corrupt cache did not fall back: calls=%d result=%#v", repo.calls, result)
	}
}
