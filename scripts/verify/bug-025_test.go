package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug025AuditStore struct{ resource, actor string }

func (*bug025AuditStore) Append(context.Context, entity.AuditLog) error { return nil }
func (s *bug025AuditStore) List(_ context.Context, _ int, _ int, resource, actor string) ([]entity.AuditLog, error) {
	s.resource = resource
	s.actor = actor
	return nil, nil
}
func TestBug025_BusinessRegression(t *testing.T) {
	s := &bug025AuditStore{}
	svc := service.AuditService{Repo: s}
	_, _ = svc.List(context.Background(), 1, 20, "property", "admin")
	if s.resource != "property" || s.actor != "admin" {
		t.Fatalf("audit filters changed: resource=%q actor=%q", s.resource, s.actor)
	}
	_, _ = svc.List(context.Background(), 1, 20, "", "")
	if s.resource != "" {
		t.Fatalf("empty resource changed to %q", s.resource)
	}
}
