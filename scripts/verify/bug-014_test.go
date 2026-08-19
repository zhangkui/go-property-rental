package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/domain/valueobject"
	"go-property-rental/internal/service"
	"testing"
)

type bug014SettlementStore struct{ input entity.Settlement }

func (s *bug014SettlementStore) Create(_ context.Context, x entity.Settlement, _ []entity.SettlementItem, _ []entity.MeterReading, _ string) (entity.Settlement, error) {
	s.input = x
	x.Refund = valueobject.Money(5000 - int64(x.DepositDeduction))
	return x, nil
}
func (*bug014SettlementStore) Get(context.Context, string) (entity.Settlement, []entity.SettlementItem, []entity.MeterReading, error) {
	return entity.Settlement{}, nil, nil, nil
}
func (*bug014SettlementStore) Complete(context.Context, string, string) error { return nil }
func TestBug014_BusinessRegression(t *testing.T) {
	s := &bug014SettlementStore{}
	svc := service.SettlementService{Repo: s}
	items := []entity.SettlementItem{{Kind: "fee", Description: "cleaning", Amount: 2800}}
	result, err := svc.Create(context.Background(), "lease-1", 2000, items, nil, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if int64(s.input.DepositDeduction) != 2000 || int64(result.Refund) != 3000 {
		t.Fatalf("deposit conservation failed: deduction=%d refund=%d", s.input.DepositDeduction, result.Refund)
	}
	result, err = svc.Create(context.Background(), "lease-1", 0, items, nil, "admin")
	if err != nil {
		t.Fatal(err)
	}
	if int64(result.Refund) != 5000 {
		t.Fatalf("zero deduction expected full refund, got %d", result.Refund)
	}
}
