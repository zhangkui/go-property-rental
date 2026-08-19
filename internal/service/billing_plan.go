package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/platform/id"
	"go-property-rental/internal/repository"
)

type BillingPlanService struct {
	Repo   repository.BillingPlanStore
	Lock   repository.BillingRunLock
	Leases repository.LeaseStore
	Now    func() time.Time
}

func (s BillingPlanService) ListPlans(ctx context.Context, page, size int, status string) ([]entity.BillingPlanBundle, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.ListPlans(ctx, limit, offset, status)
}
func (s BillingPlanService) GetPlan(ctx context.Context, id string) (entity.BillingPlanBundle, error) {
	return s.Repo.GetPlan(ctx, id)
}
func (s BillingPlanService) ListRuns(ctx context.Context, page, size int) ([]entity.BillingRun, error) {
	limit, offset := pageValues(page, size)
	return s.Repo.ListRuns(ctx, limit, offset)
}
func (s BillingPlanService) GetRun(ctx context.Context, id string) (entity.BillingRun, []entity.BillingRunItem, error) {
	return s.Repo.GetRun(ctx, id)
}

func (s BillingPlanService) SavePlan(ctx context.Context, plan entity.BillingPlan, items []entity.BillingPlanItem) (entity.BillingPlan, error) {
	if err := required(plan.LeaseID); err != nil {
		return plan, err
	}
	lease, err := s.Leases.Get(ctx, plan.LeaseID)
	if err != nil {
		return plan, err
	}
	if lease.Status != "active" {
		return plan, errors.New("billing plan requires an active lease")
	}
	if plan.BillingDay < 1 || plan.BillingDay > 28 {
		return plan, errors.New("billing day must be between 1 and 28")
	}
	if plan.DueDays < 0 || plan.DueDays > 60 {
		return plan, errors.New("due days must be between 0 and 60")
	}
	if plan.AdvanceDays < 0 || plan.AdvanceDays > 28 {
		return plan, errors.New("advance days must be between 0 and 28")
	}
	if plan.ID == "" {
		plan.ID = id.New()
	}
	if plan.Status == "" {
		plan.Status = "active"
	}
	if plan.Status != "active" && plan.Status != "paused" {
		return plan, errors.New("unsupported billing plan status")
	}
	if plan.NextPeriodStart.IsZero() {
		now := s.now()
		plan.NextPeriodStart = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
		if lease.StartDate.After(plan.NextPeriodStart) {
			plan.NextPeriodStart = day(lease.StartDate)
		}
	} else {
		plan.NextPeriodStart = day(plan.NextPeriodStart)
	}
	if plan.NextPeriodStart.Before(day(lease.StartDate)) || plan.NextPeriodStart.After(day(lease.EndDate)) {
		return plan, errors.New("next billing period must be within lease range")
	}
	allowed := map[string]bool{"water": true, "electricity": true, "property_fee": true, "service_fee": true, "other": true}
	for index := range items {
		if !allowed[items[index].Kind] {
			return plan, fmt.Errorf("unsupported charge kind: %s", items[index].Kind)
		}
		if items[index].Name == "" || items[index].Amount < 0 {
			return plan, errors.New("charge name and non-negative amount are required")
		}
		if items[index].ID == "" {
			items[index].ID = id.New()
		}
		items[index].PlanID = plan.ID
	}
	return plan, s.Repo.SavePlan(ctx, plan, items)
}

func (s BillingPlanService) SetStatus(ctx context.Context, planID, status string) error {
	if status != "active" && status != "paused" {
		return errors.New("unsupported billing plan status")
	}
	return s.Repo.SetPlanStatus(ctx, planID, status)
}

func (s BillingPlanService) Preview(ctx context.Context, at time.Time) (entity.BillingPreview, error) {
	if at.IsZero() {
		at = s.now()
	}
	bundles, err := s.Repo.DuePlans(ctx, at.UTC())
	if err != nil {
		return entity.BillingPreview{}, err
	}
	preview := entity.BillingPreview{GeneratedAt: at.UTC(), Items: make([]entity.BillingPreviewItem, 0, len(bundles))}
	for _, bundle := range bundles {
		item := previewBundle(bundle)
		preview.Items = append(preview.Items, item)
		preview.TotalAmount += item.TotalAmount
	}
	preview.Count = len(preview.Items)
	return preview, nil
}

func previewBundle(bundle entity.BillingPlanBundle) entity.BillingPreviewItem {
	periodStart := day(bundle.Plan.NextPeriodStart)
	periodEnd := periodStart.AddDate(0, 1, 0)
	if bundle.Lease.EndDate.Before(periodEnd) {
		periodEnd = day(bundle.Lease.EndDate)
	}
	additional := int64(0)
	for _, item := range bundle.Items {
		if item.Enabled {
			additional += item.Amount
		}
	}
	rent := int64(bundle.Lease.MonthlyRent)
	return entity.BillingPreviewItem{PlanID: bundle.Plan.ID, LeaseID: bundle.Lease.ID, TenantName: bundle.TenantName, PropertyLabel: bundle.PropertyLabel, PeriodStart: periodStart, PeriodEnd: periodEnd, DueDate: periodStart.AddDate(0, 0, bundle.Plan.DueDays), RentAmount: rent, AdditionalAmount: additional, TotalAmount: rent + additional, Items: bundle.Items}
}

func (s BillingPlanService) RunDue(ctx context.Context, triggerType, actor string, at time.Time) (entity.BillingRun, error) {
	if at.IsZero() {
		at = s.now()
	}
	lockKey := at.UTC().Format("2006-01-02")
	if s.Lock != nil {
		acquired, err := s.Lock.Acquire(ctx, lockKey, 15*time.Minute)
		if err != nil {
			return entity.BillingRun{}, err
		}
		if !acquired {
			return entity.BillingRun{}, errors.New("billing run is already in progress")
		}
		defer s.Lock.Release(context.Background(), lockKey)
	}
	bundles, err := s.Repo.DuePlans(ctx, at.UTC())
	if err != nil {
		return entity.BillingRun{}, err
	}
	run := entity.BillingRun{ID: id.New(), TriggerType: triggerType, PeriodKey: lockKey, Status: "running", ActorID: actor, Total: len(bundles), StartedAt: at.UTC()}
	if err = s.Repo.CreateRun(ctx, run); err != nil {
		return run, err
	}
	for _, bundle := range bundles {
		if err = s.executeBundle(ctx, run.ID, bundle, ""); err != nil {
			run.Failed++
		} else {
			run.Succeeded++
		}
	}
	finished := s.now()
	run.FinishedAt = &finished
	if err = s.Repo.FinishRun(ctx, run.ID, run.Total, run.Succeeded, run.Failed, finished); err != nil {
		return run, err
	}
	if run.Failed > 0 {
		run.Status = "completed_with_errors"
	} else {
		run.Status = "completed"
	}
	return run, nil
}

func (s BillingPlanService) executeBundle(ctx context.Context, runID string, bundle entity.BillingPlanBundle, runItemID string) error {
	preview := previewBundle(bundle)
	if runItemID == "" {
		runItemID = id.New()
	}
	runItem := entity.BillingRunItem{ID: runItemID, RunID: runID, PlanID: bundle.Plan.ID, LeaseID: bundle.Lease.ID, PeriodStart: preview.PeriodStart, PeriodEnd: preview.PeriodEnd, DueDate: preview.DueDate, Amount: preview.TotalAmount, Status: "running", Attempts: 1}
	bill := entity.Bill{ID: id.New(), LeaseID: bundle.Lease.ID, PeriodStart: preview.PeriodStart, PeriodEnd: preview.PeriodEnd, DueDate: preview.DueDate, Status: "unpaid", Amount: money(preview.TotalAmount)}
	items := []entity.BillItem{{ID: id.New(), BillID: bill.ID, Kind: "rent", Name: "月租金", Amount: preview.RentAmount}}
	for _, planItem := range bundle.Items {
		if planItem.Enabled {
			items = append(items, entity.BillItem{ID: id.New(), BillID: bill.ID, Kind: planItem.Kind, Name: planItem.Name, Amount: planItem.Amount})
		}
	}
	_, _, err := s.Repo.Generate(ctx, runItem, bill, items, bundle.Plan.ID+":"+preview.PeriodStart.Format("2006-01-02"))
	if err != nil {
		runItem.Status = "failed"
		runItem.ErrorMessage = err.Error()
		_ = s.Repo.RecordFailure(ctx, runItem)
		return err
	}
	return nil
}

func (s BillingPlanService) Retry(ctx context.Context, itemID string) error {
	item, err := s.Repo.GetRunItem(ctx, itemID)
	if err != nil {
		return err
	}
	if item.Status != "failed" {
		return errors.New("only failed billing item can be retried")
	}
	bundle, err := s.Repo.GetPlan(ctx, item.PlanID)
	if err != nil {
		return err
	}
	bundle.Plan.NextPeriodStart = item.PeriodStart.AddDate(0, 1, 0)
	return s.executeBundle(ctx, item.RunID, bundle, item.ID)
}
func (s BillingPlanService) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}
