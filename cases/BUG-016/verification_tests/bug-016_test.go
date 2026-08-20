package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug016WorkStore struct{ material entity.WorkOrderMaterial }

func (*bug016WorkStore) List(context.Context, int, int, string, string) ([]entity.WorkOrder, error) {
	return nil, nil
}
func (*bug016WorkStore) Get(context.Context, string) (entity.WorkOrder, []entity.WorkOrderMaterial, error) {
	return entity.WorkOrder{}, nil, nil
}
func (*bug016WorkStore) Create(context.Context, entity.WorkOrder) error       { return nil }
func (*bug016WorkStore) Assign(context.Context, string, string, string) error { return nil }
func (s *bug016WorkStore) AddMaterial(_ context.Context, m entity.WorkOrderMaterial) error {
	s.material = m
	return nil
}
func (*bug016WorkStore) Transition(context.Context, string, string, string, string) error { return nil }
func (*bug016WorkStore) Confirm(context.Context, string) error                            { return nil }
func TestBug016_BusinessRegression(t *testing.T) {
	s := &bug016WorkStore{}
	svc := service.WorkOrderService{Repo: s}
	if err := svc.AddMaterial(context.Background(), "work-1", "filter", 3, 2500); err != nil {
		t.Fatal(err)
	}
	if s.material.Quantity != 3 || int64(s.material.UnitCost) != 2500 || int64(s.material.Total()) != 7500 {
		t.Fatalf("material values crossed: %#v total=%d", s.material, s.material.Total())
	}
	if err := svc.AddMaterial(context.Background(), "work-1", "filter", 0, 2500); err == nil {
		t.Fatal("zero quantity must be rejected")
	}
}
