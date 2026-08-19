package repository

import (
	"context"
	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/domain/valueobject"
	"time"
)

type TenantStore interface {
	List(context.Context, int, int, string) ([]entity.Tenant, error)
	Get(context.Context, string) (entity.Tenant, error)
	Create(context.Context, entity.Tenant) error
	Update(context.Context, entity.Tenant) error
	SetStatus(context.Context, string, string) error
}
type FacilityStore interface {
	List(context.Context) ([]entity.Facility, error)
	Create(context.Context, entity.Facility) error
	ForProperty(context.Context, string) ([]entity.Facility, error)
	ReplacePropertyFacilities(context.Context, string, []string) error
}
type LeaseStore interface {
	List(context.Context, int, int, string, string) ([]entity.Lease, error)
	Get(context.Context, string) (entity.Lease, error)
	Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error
	Renew(context.Context, entity.LeaseVersion, time.Time, string) error
	Versions(context.Context, string) ([]entity.LeaseVersion, error)
	ChangeStatus(context.Context, string, string, string, string) error
}
type BillingStore interface {
	Generate(context.Context, entity.Bill, string) (entity.Bill, bool, error)
	List(context.Context, int, int, string, string) ([]entity.Bill, error)
	Get(context.Context, string) (entity.Bill, error)
	Items(context.Context, string) ([]entity.BillItem, error)
	Adjust(context.Context, entity.BillAdjustment) error
	RecordPayment(context.Context, entity.Payment, []entity.PaymentAllocation) error
}

type BillingPlanStore interface {
	ListPlans(context.Context, int, int, string) ([]entity.BillingPlanBundle, error)
	GetPlan(context.Context, string) (entity.BillingPlanBundle, error)
	SavePlan(context.Context, entity.BillingPlan, []entity.BillingPlanItem) error
	SetPlanStatus(context.Context, string, string) error
	DuePlans(context.Context, time.Time) ([]entity.BillingPlanBundle, error)
	CreateRun(context.Context, entity.BillingRun) error
	ListRuns(context.Context, int, int) ([]entity.BillingRun, error)
	GetRun(context.Context, string) (entity.BillingRun, []entity.BillingRunItem, error)
	GetRunItem(context.Context, string) (entity.BillingRunItem, error)
	Generate(context.Context, entity.BillingRunItem, entity.Bill, []entity.BillItem, string) (string, bool, error)
	RecordFailure(context.Context, entity.BillingRunItem) error
	FinishRun(context.Context, string, int, int, int, time.Time) error
}

type BillingRunLock interface {
	Acquire(context.Context, string, time.Duration) (bool, error)
	Release(context.Context, string) error
}
type DepositStore interface {
	Append(context.Context, entity.DepositLedger) error
	List(context.Context, string) ([]entity.DepositLedger, error)
	Balance(context.Context, string) (entity.DepositBalance, error)
}
type WorkOrderStore interface {
	List(context.Context, int, int, string, string) ([]entity.WorkOrder, error)
	Get(context.Context, string) (entity.WorkOrder, []entity.WorkOrderMaterial, error)
	Create(context.Context, entity.WorkOrder) error
	Assign(context.Context, string, string, string) error
	AddMaterial(context.Context, entity.WorkOrderMaterial) error
	Transition(context.Context, string, string, string, string) error
	Confirm(context.Context, string) error
}
type SettlementStore interface {
	Create(context.Context, entity.Settlement, []entity.SettlementItem, []entity.MeterReading, string) (entity.Settlement, error)
	Get(context.Context, string) (entity.Settlement, []entity.SettlementItem, []entity.MeterReading, error)
	Complete(context.Context, string, string) error
}
type AuditStore interface {
	Append(context.Context, entity.AuditLog) error
	List(context.Context, int, int, string, string) ([]entity.AuditLog, error)
}
type BalanceReader interface {
	Outstanding(context.Context, string) (valueobject.Money, error)
}

type ApprovalStore interface {
	Create(context.Context, entity.ApprovalRequest, []entity.ApprovalStep) error
	List(context.Context, int, int, string, string, string) ([]entity.ApprovalRequest, error)
	Get(context.Context, string) (entity.ApprovalRequest, []entity.ApprovalStep, []entity.ApprovalHistory, error)
	Decide(context.Context, string, string, string, string) error
	Cancel(context.Context, string, string, string) error
}

type NotificationStore interface {
	List(context.Context, string, int, int, string, string) ([]entity.Notification, error)
	UnreadCount(context.Context, string) (int, error)
	CreateMany(context.Context, []entity.Notification) error
	MarkRead(context.Context, string, string) error
	MarkAllRead(context.Context, string) error
	ListRules(context.Context) ([]entity.ReminderRule, error)
	SaveRule(context.Context, entity.ReminderRule) error
	SetRuleEnabled(context.Context, string, bool) error
	StartRun(context.Context, entity.ReminderRun) (bool, error)
	FinishRun(context.Context, string, string, int, int, time.Time) error
	RecipientIDsByRole(context.Context, string) ([]string, error)
	LeaseExpiryCandidates(context.Context, int) ([]entity.Lease, error)
	OverdueBillCandidates(context.Context, int) ([]entity.Bill, error)
	OverdueWorkOrderCandidates(context.Context, int) ([]entity.WorkOrder, error)
}

type ReportStore interface {
	Occupancy(context.Context, entity.ReportFilter) (entity.OccupancyReport, error)
	RentRoll(context.Context, entity.ReportFilter) ([]entity.RentRollRow, error)
	ReceivableAging(context.Context, entity.ReportFilter, time.Time) ([]entity.ReceivableAgingRow, error)
	DepositReconciliation(context.Context, entity.ReportFilter) ([]entity.DepositReconciliationRow, error)
	MaintenanceSLA(context.Context, entity.ReportFilter, time.Time) (entity.MaintenanceSLAReport, error)
}

type DashboardStore interface {
	Summary(context.Context, string, time.Time) (entity.DashboardSummary, error)
}

type DashboardCache interface {
	Get(context.Context, string) (entity.DashboardSummary, bool, error)
	Set(context.Context, string, entity.DashboardSummary, time.Duration) error
}
