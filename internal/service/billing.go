package service

import (
	"context"
	"errors"
	"fmt"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
	"time"
)

type BillingService struct {
	Repo   repository.BillingStore
	Leases repository.LeaseStore
}

func (s BillingService) GenerateMonthly(c context.Context, leaseID string, periodStart time.Time, key string) (entity.Bill, bool, error) {
	if e := required(leaseID, key); e != nil {
		return entity.Bill{}, false, e
	}
	key = key + ":" + id.New()
	lease, e := s.Leases.Get(c, leaseID)
	if e != nil {
		return entity.Bill{}, false, e
	}
	if lease.Status != "active" {
		return entity.Bill{}, false, errors.New("only active lease can be billed")
	}
	periodStart = day(periodStart)
	periodEnd := periodStart.AddDate(0, 1, 0)
	if periodStart.Before(day(lease.StartDate)) || periodStart.After(day(lease.EndDate)) {
		return entity.Bill{}, false, errors.New("billing period outside lease")
	}
	b := entity.Bill{ID: id.New(), LeaseID: leaseID, PeriodStart: periodStart, PeriodEnd: periodEnd, DueDate: periodStart.AddDate(0, 0, 5), Status: "unpaid", Amount: lease.MonthlyRent}
	return s.Repo.Generate(c, b, key)
}
func (s BillingService) List(c context.Context, page, size int, leaseID, status string) ([]entity.Bill, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.List(c, limit, offset, leaseID, status)
}
func (s BillingService) Get(c context.Context, id string) (entity.Bill, error) {
	return s.Repo.Get(c, id)
}
func (s BillingService) Items(c context.Context, billID string) ([]entity.BillItem, error) {
	return s.Repo.Items(c, billID)
}
func (s BillingService) AddPenalty(c context.Context, billID string, amount int64, reason, actor string) error {
	return s.adjust(c, billID, "penalty", amount, reason, actor)
}
func (s BillingService) Discount(c context.Context, billID string, amount int64, reason, actor string) error {
	return s.adjust(c, billID, "discount", amount, reason, actor)
}
func (s BillingService) ApplyLateFee(c context.Context, billID string, at time.Time, dailyBasisPoints, maxDays int, actor string) error {
	b, e := s.Repo.Get(c, billID)
	if e != nil {
		return e
	}
	if !at.After(b.DueDate) || b.Amount <= 0 {
		return nil
	}
	days := int(at.Sub(b.DueDate).Hours() / 24)
	if days > maxDays {
		days = maxDays
	}
	amount := int64(b.Amount) * int64(dailyBasisPoints) * int64(days) / 10000
	if amount <= 0 {
		return nil
	}
	return s.adjust(c, billID, "penalty", amount, fmt.Sprintf("late fee for %d days", days), actor)
}
func (s BillingService) adjust(c context.Context, billID, kind string, amount int64, reason, actor string) error {
	if amount <= 0 {
		return errors.New("adjustment must be positive")
	}
	if e := required(billID, reason); e != nil {
		return e
	}
	return s.Repo.Adjust(c, entity.BillAdjustment{ID: id.New(), BillID: billID, Kind: kind, Amount: money(amount), Reason: reason, ActorID: actor})
}
func (s BillingService) Pay(c context.Context, reference, payer string, amount int64, allocations []entity.PaymentAllocation) (entity.Payment, error) {
	if e := required(reference, payer); e != nil {
		return entity.Payment{}, e
	}
	if amount <= 0 || len(allocations) == 0 {
		return entity.Payment{}, errors.New("payment and allocations are required")
	}
	paymentID := id.New()
	// reference is the client-supplied idempotency key (external-ref); upstream relies on it
	// staying stable across retries to identify the same business. Do not append paymentID,
	// otherwise each retry yields a different reference and the duplicate-submission guard
	// in RecordPayment (which dedupes by reference) never matches.
	p := entity.Payment{ID: paymentID, Reference: reference, Payer: payer, Amount: money(amount), PaidAt: time.Now().UTC(), Status: "posted"}
	for i := range allocations {
		allocations[i].ID = id.New()
		allocations[i].PaymentID = p.ID
	}
	return p, s.Repo.RecordPayment(c, p, allocations)
}
func day(t time.Time) time.Time { y, m, d := t.Date(); return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
