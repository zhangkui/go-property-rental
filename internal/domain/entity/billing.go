package entity

import (
	"go-property-rental/internal/domain/valueobject"
	"time"
)

type Bill struct {
	ID, LeaseID                     string
	PeriodStart, PeriodEnd          time.Time
	Status                          string
	Amount, Penalty, Discount, Paid valueobject.Money
	DueDate                         time.Time
}

func (b Bill) Outstanding() valueobject.Money {
	return b.Amount.Add(b.Penalty).Sub(b.Discount).Sub(b.Paid)
}

type BillAdjustment struct {
	ID, BillID, Kind, Reason, ActorID string
	Amount                            valueobject.Money
	CreatedAt                         time.Time
}
type Payment struct {
	ID, Reference, Status string
	Amount                valueobject.Money
	PaidAt                time.Time
	Payer                 string
}
type PaymentAllocation struct {
	ID, PaymentID, BillID string
	Amount                valueobject.Money
	CreatedAt             time.Time
}
