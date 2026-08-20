package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug020PropertyRepo struct {
	limit, offset int
	status        string
}

func (s *bug020PropertyRepo) List(_ context.Context, limit, offset int, status string) ([]entity.Property, error) {
	s.limit = limit
	s.offset = offset
	s.status = status
	return nil, nil
}
func (*bug020PropertyRepo) Get(context.Context, string) (entity.Property, error) {
	return entity.Property{}, nil
}
func (*bug020PropertyRepo) Create(context.Context, entity.Property) error      { return nil }
func (*bug020PropertyRepo) UpdateStatus(context.Context, string, string) error { return nil }
func TestBug020_BusinessRegression(t *testing.T) {
	s := &bug020PropertyRepo{}
	svc := service.PropertyService{Repo: s}
	_, _ = svc.List(context.Background(), 2, 2, "maintenance")
	if s.limit != 2 || s.offset != 2 {
		t.Fatalf("page 2 expected limit=2 offset=2, got %d/%d", s.limit, s.offset)
	}
	_, _ = svc.List(context.Background(), 1, 2, "maintenance")
	if s.offset != 0 {
		t.Fatalf("page 1 expected offset 0, got %d", s.offset)
	}
	_, _ = svc.List(context.Background(), 0, 0, "")
	if s.limit != 20 || s.offset != 0 {
		t.Fatalf("invalid pagination defaults expected 20/0, got %d/%d", s.limit, s.offset)
	}
}
