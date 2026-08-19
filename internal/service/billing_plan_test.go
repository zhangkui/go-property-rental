package service

import (
	"context"
	"testing"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/domain/valueobject"
)

type billingPlanStoreStub struct {
	duePlans         []entity.BillingPlanBundle
	plan             entity.BillingPlanBundle
	runItem          entity.BillingRunItem
	generatedRunItem entity.BillingRunItem
	saved            entity.BillingPlan
}

func (s *billingPlanStoreStub) ListPlans(context.Context, int, int, string) ([]entity.BillingPlanBundle, error) {
	return nil, nil
}
func (s *billingPlanStoreStub) GetPlan(context.Context, string) (entity.BillingPlanBundle, error) {
	return s.plan, nil
}
func (s *billingPlanStoreStub) SavePlan(_ context.Context, plan entity.BillingPlan, _ []entity.BillingPlanItem) error {
	s.saved = plan
	return nil
}
func (s *billingPlanStoreStub) SetPlanStatus(context.Context, string, string) error { return nil }
func (s *billingPlanStoreStub) DuePlans(context.Context, time.Time) ([]entity.BillingPlanBundle, error) {
	return s.duePlans, nil
}
func (s *billingPlanStoreStub) CreateRun(context.Context, entity.BillingRun) error { return nil }
func (s *billingPlanStoreStub) ListRuns(context.Context, int, int) ([]entity.BillingRun, error) {
	return nil, nil
}
func (s *billingPlanStoreStub) GetRun(context.Context, string) (entity.BillingRun, []entity.BillingRunItem, error) {
	return entity.BillingRun{}, nil, nil
}
func (s *billingPlanStoreStub) GetRunItem(context.Context, string) (entity.BillingRunItem, error) {
	return s.runItem, nil
}
func (s *billingPlanStoreStub) Generate(_ context.Context, item entity.BillingRunItem, _ entity.Bill, _ []entity.BillItem, _ string) (string, bool, error) {
	s.generatedRunItem = item
	return "", false, nil
}
func (s *billingPlanStoreStub) RecordFailure(context.Context, entity.BillingRunItem) error {
	return nil
}
func (s *billingPlanStoreStub) FinishRun(context.Context, string, int, int, int, time.Time) error {
	return nil
}

type leaseStoreStub struct{ lease entity.Lease }

func (s leaseStoreStub) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (s leaseStoreStub) Get(context.Context, string) (entity.Lease, error) { return s.lease, nil }
func (s leaseStoreStub) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	return nil
}
func (s leaseStoreStub) Renew(context.Context, entity.LeaseVersion, time.Time, string) error {
	return nil
}
func (s leaseStoreStub) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (s leaseStoreStub) ChangeStatus(context.Context, string, string, string, string) error {
	return nil
}

func TestBillingPlanPreviewIncludesEnabledChargesAndCapsLeaseEnd(t *testing.T) {
	periodStart := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	store := &billingPlanStoreStub{duePlans: []entity.BillingPlanBundle{{
		Plan: entity.BillingPlan{ID: "plan-1", LeaseID: "lease-1", DueDays: 3, NextPeriodStart: periodStart},
		Lease: entity.Lease{
			ID:          "lease-1",
			Status:      "active",
			EndDate:     time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC),
			MonthlyRent: valueobject.Money(350000),
		},
		Items: []entity.BillingPlanItem{
			{Kind: "water", Amount: 1800, Enabled: true},
			{Kind: "electricity", Amount: 4200, Enabled: true},
			{Kind: "service_fee", Amount: 1500, Enabled: false},
		},
	}}}

	preview, err := (BillingPlanService{Repo: store}).Preview(context.Background(), periodStart)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Count != 1 || preview.TotalAmount != 356000 {
		t.Fatalf("unexpected preview totals: %#v", preview)
	}
	item := preview.Items[0]
	if item.AdditionalAmount != 6000 || !item.PeriodEnd.Equal(store.duePlans[0].Lease.EndDate) {
		t.Fatalf("unexpected preview item: %#v", item)
	}
}

func TestBillingPlanSaveRejectsPeriodOutsideLease(t *testing.T) {
	lease := entity.Lease{
		ID:        "lease-1",
		Status:    "active",
		StartDate: time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC),
	}
	store := &billingPlanStoreStub{}
	service := BillingPlanService{Repo: store, Leases: leaseStoreStub{lease: lease}}
	plan := entity.BillingPlan{
		LeaseID:         lease.ID,
		BillingDay:      18,
		DueDays:         3,
		AdvanceDays:     2,
		NextPeriodStart: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}

	if _, err := service.SavePlan(context.Background(), plan, nil); err == nil {
		t.Fatal("expected period outside lease range to be rejected")
	}
	if store.saved.ID != "" {
		t.Fatal("invalid plan must not be persisted")
	}
}

func TestBillingPlanRetryReusesFailedRunItem(t *testing.T) {
	periodStart := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	store := &billingPlanStoreStub{
		runItem: entity.BillingRunItem{ID: "item-1", RunID: "run-1", PlanID: "plan-1", Status: "failed", PeriodStart: periodStart},
		plan: entity.BillingPlanBundle{
			Plan:  entity.BillingPlan{ID: "plan-1", LeaseID: "lease-1"},
			Lease: entity.Lease{ID: "lease-1", EndDate: periodStart.AddDate(0, 1, 0), MonthlyRent: valueobject.Money(100000)},
		},
	}

	if err := (BillingPlanService{Repo: store}).Retry(context.Background(), "item-1"); err != nil {
		t.Fatal(err)
	}
	if store.generatedRunItem.ID != "item-1" || store.generatedRunItem.RunID != "run-1" {
		t.Fatalf("retry must update the original run item: %#v", store.generatedRunItem)
	}
}
