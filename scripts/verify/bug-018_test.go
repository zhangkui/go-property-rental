package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug018TenantStore struct{ updated entity.Tenant }

func (*bug018TenantStore) List(context.Context, int, int, string) ([]entity.Tenant, error) {
	return nil, nil
}
func (*bug018TenantStore) Get(context.Context, string) (entity.Tenant, error) {
	return entity.Tenant{}, nil
}
func (*bug018TenantStore) Create(context.Context, entity.Tenant) error { return nil }
func (s *bug018TenantStore) Update(_ context.Context, x entity.Tenant) error {
	s.updated = x
	return nil
}
func (*bug018TenantStore) SetStatus(context.Context, string, string) error { return nil }
func TestBug018_BusinessRegression(t *testing.T) {
	s := &bug018TenantStore{}
	input := entity.Tenant{ID: "tenant-1", Name: "Updated Tenant", Phone: "13912345678", Email: "updated@example.com", IdentityNo: "ID-2026-018", Status: "active"}
	if err := (service.TenantService{Repo: s}).Update(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	if s.updated.Name != input.Name || s.updated.Phone != input.Phone || s.updated.Email != input.Email || s.updated.IdentityNo != input.IdentityNo || s.updated.Status != "active" {
		t.Fatalf("tenant fields crossed: %#v", s.updated)
	}
}
