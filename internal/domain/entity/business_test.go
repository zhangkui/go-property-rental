package entity

import (
	"go-property-rental/internal/domain/valueobject"
	"testing"
)

func TestBillOutstanding(t *testing.T) {
	b := Bill{Amount: 10000, Penalty: 500, Discount: 200, Paid: 3000}
	if got := b.Outstanding(); got != 7300 {
		t.Fatalf("outstanding=%d", got)
	}
}
func TestDepositAvailable(t *testing.T) {
	b := DepositBalance{Received: 10000, Deducted: 2500, Refunded: 1000}
	if b.Available() != 6500 {
		t.Fatalf("available=%d", b.Available())
	}
}
func TestMaterialTotal(t *testing.T) {
	m := WorkOrderMaterial{Quantity: 3, UnitCost: 1200}
	if m.Total() != 3600 {
		t.Fatalf("total=%d", m.Total())
	}
}
func TestSettlementNetPayable(t *testing.T) {
	s := Settlement{Total: valueobject.Money(9000), DepositDeduction: 4000}
	if s.NetPayable() != 5000 {
		t.Fatalf("net=%d", s.NetPayable())
	}
}
