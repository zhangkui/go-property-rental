package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-property-rental/internal/domain/entity"
)

type BillingPlans struct{ DB *sql.DB }

func (r BillingPlans) ListPlans(ctx context.Context, limit, offset int, status string) ([]entity.BillingPlanBundle, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT bp.id,bp.lease_id,bp.status,bp.billing_day,bp.due_days,bp.advance_days,bp.next_period_start,bp.created_at,bp.updated_at,l.property_id,l.tenant_id,l.status,l.start_date,l.end_date,l.monthly_rent,l.deposit,t.name,CONCAT(p.building,' ',p.room) FROM billing_plans bp JOIN leases l ON l.id=bp.lease_id JOIN tenants t ON t.id=l.tenant_id JOIN properties p ON p.id=l.property_id WHERE (?='' OR bp.status=?) ORDER BY bp.updated_at DESC LIMIT ? OFFSET ?`, status, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.BillingPlanBundle, 0)
	for rows.Next() {
		bundle, err := scanBillingPlanBundle(rows)
		if err != nil {
			return nil, err
		}
		bundle.Items, err = r.listItems(ctx, bundle.Plan.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, bundle)
	}
	return items, rows.Err()
}

func (r BillingPlans) GetPlan(ctx context.Context, planID string) (entity.BillingPlanBundle, error) {
	bundle, err := scanBillingPlanBundle(r.DB.QueryRowContext(ctx, `SELECT bp.id,bp.lease_id,bp.status,bp.billing_day,bp.due_days,bp.advance_days,bp.next_period_start,bp.created_at,bp.updated_at,l.property_id,l.tenant_id,l.status,l.start_date,l.end_date,l.monthly_rent,l.deposit,t.name,CONCAT(p.building,' ',p.room) FROM billing_plans bp JOIN leases l ON l.id=bp.lease_id JOIN tenants t ON t.id=l.tenant_id JOIN properties p ON p.id=l.property_id WHERE bp.id=?`, planID))
	if err != nil {
		return bundle, err
	}
	bundle.Items, err = r.listItems(ctx, planID)
	return bundle, err
}

func scanBillingPlanBundle(scanner interface{ Scan(...any) error }) (entity.BillingPlanBundle, error) {
	var bundle entity.BillingPlanBundle
	err := scanner.Scan(&bundle.Plan.ID, &bundle.Plan.LeaseID, &bundle.Plan.Status, &bundle.Plan.BillingDay, &bundle.Plan.DueDays, &bundle.Plan.AdvanceDays, &bundle.Plan.NextPeriodStart, &bundle.Plan.CreatedAt, &bundle.Plan.UpdatedAt, &bundle.Lease.PropertyID, &bundle.Lease.TenantID, &bundle.Lease.Status, &bundle.Lease.StartDate, &bundle.Lease.EndDate, &bundle.Lease.MonthlyRent, &bundle.Lease.Deposit, &bundle.TenantName, &bundle.PropertyLabel)
	bundle.Lease.ID = bundle.Plan.LeaseID
	return bundle, err
}

func (r BillingPlans) listItems(ctx context.Context, planID string) ([]entity.BillingPlanItem, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,plan_id,kind,name,amount,enabled FROM billing_plan_items WHERE plan_id=? ORDER BY kind,name`, planID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.BillingPlanItem, 0)
	for rows.Next() {
		var item entity.BillingPlanItem
		if err := rows.Scan(&item.ID, &item.PlanID, &item.Kind, &item.Name, &item.Amount, &item.Enabled); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r BillingPlans) SavePlan(ctx context.Context, plan entity.BillingPlan, items []entity.BillingPlanItem) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var existing int
	if err = tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM billing_plans WHERE id=?`, plan.ID).Scan(&existing); err != nil {
		return err
	}
	if existing == 0 {
		_, err = tx.ExecContext(ctx, `INSERT INTO billing_plans(id,lease_id,status,billing_day,due_days,advance_days,next_period_start) VALUES(?,?,?,?,?,?,?)`, plan.ID, plan.LeaseID, plan.Status, plan.BillingDay, plan.DueDays, plan.AdvanceDays, plan.NextPeriodStart)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE billing_plans SET lease_id=?,status=?,billing_day=?,due_days=?,advance_days=?,next_period_start=? WHERE id=?`, plan.LeaseID, plan.Status, plan.BillingDay, plan.DueDays, plan.AdvanceDays, plan.NextPeriodStart, plan.ID)
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM billing_plan_items WHERE plan_id=?`, plan.ID); err != nil {
		return err
	}
	for _, item := range items {
		if _, err = tx.ExecContext(ctx, `INSERT INTO billing_plan_items(id,plan_id,kind,name,amount,enabled) VALUES(?,?,?,?,?,?)`, item.ID, plan.ID, item.Kind, item.Name, item.Amount, item.Enabled); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r BillingPlans) SetPlanStatus(ctx context.Context, planID, status string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE billing_plans SET status=? WHERE id=?`, status, planID)
	return err
}

func (r BillingPlans) DuePlans(ctx context.Context, at time.Time) ([]entity.BillingPlanBundle, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT bp.id,bp.lease_id,bp.status,bp.billing_day,bp.due_days,bp.advance_days,bp.next_period_start,bp.created_at,bp.updated_at,l.property_id,l.tenant_id,l.status,l.start_date,l.end_date,l.monthly_rent,l.deposit,t.name,CONCAT(p.building,' ',p.room) FROM billing_plans bp JOIN leases l ON l.id=bp.lease_id JOIN tenants t ON t.id=l.tenant_id JOIN properties p ON p.id=l.property_id WHERE bp.status='active' AND l.status='active' AND DATE_SUB(DATE_ADD(LAST_DAY(DATE_SUB(bp.next_period_start,INTERVAL 1 MONTH)),INTERVAL bp.billing_day DAY),INTERVAL bp.advance_days DAY)<=? AND bp.next_period_start<=l.end_date ORDER BY bp.next_period_start,bp.id`, at)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.BillingPlanBundle, 0)
	for rows.Next() {
		bundle, err := scanBillingPlanBundle(rows)
		if err != nil {
			return nil, err
		}
		bundle.Items, err = r.listItems(ctx, bundle.Plan.ID)
		if err != nil {
			return nil, err
		}
		items = append(items, bundle)
	}
	return items, rows.Err()
}

func (r BillingPlans) CreateRun(ctx context.Context, run entity.BillingRun) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO billing_runs(id,trigger_type,period_key,status,actor_id,total,succeeded,failed,started_at) VALUES(?,?,?,?,?,?,?,?,?)`, run.ID, run.TriggerType, run.PeriodKey, run.Status, nullableString(run.ActorID), run.Total, run.Succeeded, run.Failed, run.StartedAt)
	return err
}

func (r BillingPlans) ListRuns(ctx context.Context, limit, offset int) ([]entity.BillingRun, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,trigger_type,period_key,status,COALESCE(actor_id,''),total,succeeded,failed,started_at,finished_at FROM billing_runs ORDER BY started_at DESC LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.BillingRun, 0)
	for rows.Next() {
		var item entity.BillingRun
		if err := rows.Scan(&item.ID, &item.TriggerType, &item.PeriodKey, &item.Status, &item.ActorID, &item.Total, &item.Succeeded, &item.Failed, &item.StartedAt, &item.FinishedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r BillingPlans) GetRun(ctx context.Context, runID string) (entity.BillingRun, []entity.BillingRunItem, error) {
	var run entity.BillingRun
	err := r.DB.QueryRowContext(ctx, `SELECT id,trigger_type,period_key,status,COALESCE(actor_id,''),total,succeeded,failed,started_at,finished_at FROM billing_runs WHERE id=?`, runID).Scan(&run.ID, &run.TriggerType, &run.PeriodKey, &run.Status, &run.ActorID, &run.Total, &run.Succeeded, &run.Failed, &run.StartedAt, &run.FinishedAt)
	if err != nil {
		return run, nil, err
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT id,run_id,plan_id,lease_id,COALESCE(bill_id,''),period_start,period_end,due_date,amount,status,error_message,attempts,created_at,updated_at FROM billing_run_items WHERE run_id=? ORDER BY created_at,id`, runID)
	if err != nil {
		return run, nil, err
	}
	defer rows.Close()
	items := make([]entity.BillingRunItem, 0)
	for rows.Next() {
		item, err := scanBillingRunItem(rows)
		if err != nil {
			return run, nil, err
		}
		items = append(items, item)
	}
	return run, items, rows.Err()
}

func (r BillingPlans) GetRunItem(ctx context.Context, itemID string) (entity.BillingRunItem, error) {
	return scanBillingRunItem(r.DB.QueryRowContext(ctx, `SELECT id,run_id,plan_id,lease_id,COALESCE(bill_id,''),period_start,period_end,due_date,amount,status,error_message,attempts,created_at,updated_at FROM billing_run_items WHERE id=?`, itemID))
}

func scanBillingRunItem(scanner interface{ Scan(...any) error }) (entity.BillingRunItem, error) {
	var item entity.BillingRunItem
	err := scanner.Scan(&item.ID, &item.RunID, &item.PlanID, &item.LeaseID, &item.BillID, &item.PeriodStart, &item.PeriodEnd, &item.DueDate, &item.Amount, &item.Status, &item.ErrorMessage, &item.Attempts, &item.CreatedAt, &item.UpdatedAt)
	return item, err
}

func (r BillingPlans) Generate(ctx context.Context, runItem entity.BillingRunItem, bill entity.Bill, items []entity.BillItem, key string) (string, bool, error) {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()
	var existing string
	err = tx.QueryRowContext(ctx, `SELECT resource_id FROM idempotency_keys WHERE scope='billing.plan' AND request_key=? FOR UPDATE`, key).Scan(&existing)
	if err == nil {
		if _, err = tx.ExecContext(ctx, `INSERT INTO billing_run_items(id,run_id,plan_id,lease_id,bill_id,period_start,period_end,due_date,amount,status,error_message,attempts) VALUES(?,?,?,?,?,?,?,?,?,'succeeded','',?) ON DUPLICATE KEY UPDATE bill_id=VALUES(bill_id),status='succeeded',error_message='',attempts=attempts+1`, runItem.ID, runItem.RunID, runItem.PlanID, runItem.LeaseID, existing, runItem.PeriodStart, runItem.PeriodEnd, runItem.DueDate, runItem.Amount, runItem.Attempts); err != nil {
			return "", false, err
		}
		if _, err = tx.ExecContext(ctx, `UPDATE billing_plans SET next_period_start=DATE_ADD(next_period_start,INTERVAL 2 MONTH) WHERE id=? AND next_period_start=?`, runItem.PlanID, runItem.PeriodStart); err != nil {
			return "", false, err
		}
		return existing, false, tx.Commit()
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}
	var status string
	if err = tx.QueryRowContext(ctx, `SELECT status FROM billing_plans WHERE id=? FOR UPDATE`, runItem.PlanID).Scan(&status); err != nil {
		return "", false, err
	}
	if status != "active" {
		return "", false, errors.New("billing plan is inactive")
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO bills(id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid) VALUES(?,?,?,?,?,'unpaid',?,0,0,0)`, bill.ID, bill.LeaseID, bill.PeriodStart, bill.PeriodEnd, bill.DueDate, bill.Amount); err != nil {
		return "", false, err
	}
	for _, item := range items {
		if _, err = tx.ExecContext(ctx, `INSERT INTO bill_items(id,bill_id,kind,name,amount) VALUES(?,?,?,?,?)`, item.ID, bill.ID, item.Kind, item.Name, item.Amount); err != nil {
			return "", false, err
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO idempotency_keys(scope,request_key,resource_id) VALUES('billing.plan',?,?)`, key, bill.ID); err != nil {
		return "", false, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO billing_run_items(id,run_id,plan_id,lease_id,bill_id,period_start,period_end,due_date,amount,status,error_message,attempts) VALUES(?,?,?,?,?,?,?,?,?,'succeeded','',?) ON DUPLICATE KEY UPDATE bill_id=VALUES(bill_id),status='succeeded',error_message='',attempts=attempts+1`, runItem.ID, runItem.RunID, runItem.PlanID, runItem.LeaseID, bill.ID, runItem.PeriodStart, runItem.PeriodEnd, runItem.DueDate, runItem.Amount, runItem.Attempts); err != nil {
		return "", false, err
	}
	if _, err = tx.ExecContext(ctx, `UPDATE billing_plans SET next_period_start=DATE_ADD(next_period_start,INTERVAL 2 MONTH) WHERE id=?`, runItem.PlanID); err != nil {
		return "", false, err
	}
	if err = tx.Commit(); err != nil {
		return "", false, err
	}
	return bill.ID, true, nil
}

func (r BillingPlans) RecordFailure(ctx context.Context, item entity.BillingRunItem) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO billing_run_items(id,run_id,plan_id,lease_id,period_start,period_end,due_date,amount,status,error_message,attempts) VALUES(?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE status='failed',error_message=VALUES(error_message),attempts=attempts+1`, item.ID, item.RunID, item.PlanID, item.LeaseID, item.PeriodStart, item.PeriodEnd, item.DueDate, item.Amount, "failed", item.ErrorMessage, item.Attempts)
	return err
}

func (r BillingPlans) FinishRun(ctx context.Context, runID string, total, succeeded, failed int, finishedAt time.Time) error {
	status := "completed"
	if failed > 0 {
		status = "completed_with_errors"
	}
	_, err := r.DB.ExecContext(ctx, `UPDATE billing_runs SET status=?,total=?,succeeded=?,failed=?,finished_at=? WHERE id=?`, status, total, succeeded, failed, finishedAt, runID)
	return err
}
