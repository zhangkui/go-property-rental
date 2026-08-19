package entity

import "time"

type BillingPlan struct {
	ID, LeaseID, Status  string
	BillingDay, DueDays  int
	AdvanceDays          int
	NextPeriodStart      time.Time
	CreatedAt, UpdatedAt time.Time
}

type BillingPlanItem struct {
	ID, PlanID, Kind, Name string
	Amount                 int64
	Enabled                bool
}

type BillItem struct {
	ID, BillID, Kind, Name string
	Amount                 int64
}

type BillingPlanBundle struct {
	Plan          BillingPlan
	Lease         Lease
	TenantName    string
	PropertyLabel string
	Items         []BillingPlanItem
}

type BillingPreviewItem struct {
	PlanID, LeaseID, TenantName, PropertyLabel string
	PeriodStart, PeriodEnd, DueDate            time.Time
	RentAmount, AdditionalAmount, TotalAmount  int64
	Items                                      []BillingPlanItem
}

type BillingPreview struct {
	GeneratedAt time.Time
	Count       int
	TotalAmount int64
	Items       []BillingPreviewItem
}

type BillingRun struct {
	ID, TriggerType, Status, ActorID string
	PeriodKey                        string
	Total, Succeeded, Failed         int
	StartedAt                        time.Time
	FinishedAt                       *time.Time
}

type BillingRunItem struct {
	ID, RunID, PlanID, LeaseID, BillID string
	PeriodStart, PeriodEnd, DueDate    time.Time
	Amount                             int64
	Status, ErrorMessage               string
	Attempts                           int
	CreatedAt, UpdatedAt               time.Time
}
