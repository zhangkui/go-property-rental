package verify

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
	"testing"
)

type bug012DepositStore struct{ items []entity.DepositLedger }

func (s *bug012DepositStore) Append(_ context.Context, x entity.DepositLedger) error {
	s.items = append(s.items, x)
	return nil
}
func (*bug012DepositStore) List(context.Context, string) ([]entity.DepositLedger, error) {
	return nil, nil
}
func (*bug012DepositStore) Balance(context.Context, string) (entity.DepositBalance, error) {
	return entity.DepositBalance{}, nil
}
func TestBug012_BusinessRegression(t *testing.T) {
	s := &bug012DepositStore{}
	svc := service.DepositService{Repo: s}
	for i := 0; i < 2; i++ {
		x, err := svc.Transaction(context.Background(), "lease-1", "collect", "client-ref", "deposit", "admin", 5000)
		if err != nil {
			t.Fatal(err)
		}
		if x.Reference != "client-ref" {
			t.Fatalf("reference rewritten to %q", x.Reference)
		}
	}
	if len(s.items) != 2 || s.items[0].Reference != s.items[1].Reference {
		t.Fatalf("retries did not preserve reference: %#v", s.items)
	}
	if _, err := svc.Transaction(context.Background(), "lease-1", "collect", "client-ref", "deposit", "admin", 0); err == nil {
		t.Fatal("zero deposit must be rejected")
	}
}
