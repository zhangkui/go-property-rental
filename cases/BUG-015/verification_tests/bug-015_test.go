package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug015WorkStore struct{ confirmed string }

func (*bug015WorkStore) List(context.Context, int, int, string, string) ([]entity.WorkOrder, error) {
	return nil, nil
}
func (*bug015WorkStore) Get(context.Context, string) (entity.WorkOrder, []entity.WorkOrderMaterial, error) {
	return entity.WorkOrder{}, nil, nil
}
func (*bug015WorkStore) Create(context.Context, entity.WorkOrder) error                   { return nil }
func (*bug015WorkStore) Assign(context.Context, string, string, string) error             { return nil }
func (*bug015WorkStore) AddMaterial(context.Context, entity.WorkOrderMaterial) error      { return nil }
func (*bug015WorkStore) Transition(context.Context, string, string, string, string) error { return nil }
func (s *bug015WorkStore) Confirm(_ context.Context, id string) error                     { s.confirmed = id; return nil }
func TestBug015_BusinessRegression(t *testing.T) {
	s := &bug015WorkStore{}
	if err := (service.WorkOrderService{Repo: s}).Confirm(context.Background(), "work-1"); err != nil {
		t.Fatal(err)
	}
	if s.confirmed != "work-1" {
		t.Fatalf("confirmation bypass marker added to id: %q", s.confirmed)
	}
}
