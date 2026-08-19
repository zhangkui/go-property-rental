package mysql

import (
	"context"
	"database/sql"
	"time"

	"go-property-rental/internal/domain/entity"
)

type Notifications struct{ DB *sql.DB }

func (r Notifications) List(ctx context.Context, userID string, limit, offset int, status, category string) ([]entity.Notification, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,user_id,category,title,content,resource_type,resource_id,status,created_at,read_at FROM notifications WHERE user_id=? AND (?='' OR status=?) AND (?='' OR category=?) ORDER BY created_at DESC LIMIT ? OFFSET ?`, userID, status, status, category, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.Notification, 0)
	for rows.Next() {
		var item entity.Notification
		if err := rows.Scan(&item.ID, &item.UserID, &item.Category, &item.Title, &item.Content, &item.ResourceType, &item.ResourceID, &item.Status, &item.CreatedAt, &item.ReadAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Notifications) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM notifications WHERE user_id=? AND status='unread'`, userID).Scan(&count)
	return count, err
}

func (r Notifications) CreateMany(ctx context.Context, items []entity.Notification) error {
	if len(items) == 0 {
		return nil
	}
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, item := range items {
		if _, err = tx.ExecContext(ctx, `INSERT INTO notifications(id,user_id,category,title,content,resource_type,resource_id,status,created_at) VALUES(?,?,?,?,?,?,?,?,?)`, item.ID, item.UserID, item.Category, item.Title, item.Content, item.ResourceType, item.ResourceID, item.Status, item.CreatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r Notifications) MarkRead(ctx context.Context, userID, notificationID string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE notifications SET status='read',read_at=UTC_TIMESTAMP() WHERE id=?`, notificationID)
	return err
}

func (r Notifications) MarkAllRead(ctx context.Context, userID string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE notifications SET status='read',read_at=UTC_TIMESTAMP() WHERE user_id=? AND status='unread'`, userID)
	return err
}

func (r Notifications) ListRules(ctx context.Context) ([]entity.ReminderRule, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,name,category,recipient_role,days_before,days_after,enabled,created_at,updated_at FROM reminder_rules ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.ReminderRule, 0)
	for rows.Next() {
		var item entity.ReminderRule
		if err := rows.Scan(&item.ID, &item.Name, &item.Category, &item.RecipientRole, &item.DaysBefore, &item.DaysAfter, &item.Enabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Notifications) SaveRule(ctx context.Context, rule entity.ReminderRule) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO reminder_rules(id,name,category,recipient_role,days_before,days_after,enabled) VALUES(?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE category=VALUES(category),recipient_role=VALUES(recipient_role),days_before=VALUES(days_before),days_after=VALUES(days_after),enabled=VALUES(enabled)`, rule.ID, rule.Name, rule.Category, rule.RecipientRole, rule.DaysBefore, rule.DaysAfter, rule.Enabled)
	return err
}

func (r Notifications) SetRuleEnabled(ctx context.Context, ruleID string, enabled bool) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE reminder_rules SET enabled=? WHERE id=?`, enabled, ruleID)
	return err
}

func (r Notifications) StartRun(ctx context.Context, run entity.ReminderRun) (bool, error) {
	result, err := r.DB.ExecContext(ctx, `INSERT INTO reminder_runs(id,rule_id,run_key,status,started_at) VALUES(?,?,?,?,?)`, run.ID, run.RuleID, run.RunKey, "running", run.StartedAt)
	if err != nil {
		return false, err
	}
	n, err := result.RowsAffected()
	return n == 1, err
}

func (r Notifications) FinishRun(ctx context.Context, runID, status string, generated, failed int, finishedAt time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE reminder_runs SET status=?,generated_count=?,failed_count=?,finished_at=? WHERE id=?`, status, generated, failed, finishedAt, runID)
	return err
}

func (r Notifications) RecipientIDsByRole(ctx context.Context, role string) ([]string, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT ur.user_id FROM user_roles ur JOIN roles r ON r.id=ur.role_id JOIN users u ON u.id=ur.user_id WHERE r.name=? AND u.status='active'`, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]string, 0)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	return items, rows.Err()
}

func (r Notifications) LeaseExpiryCandidates(ctx context.Context, days int) ([]entity.Lease, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit FROM leases WHERE status IN ('active','approved') AND end_date BETWEEN CURDATE() AND DATE_ADD(CURDATE(),INTERVAL ? DAY)`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.Lease, 0)
	for rows.Next() {
		var item entity.Lease
		if err := rows.Scan(&item.ID, &item.PropertyID, &item.TenantID, &item.Status, &item.StartDate, &item.EndDate, &item.MonthlyRent, &item.Deposit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Notifications) OverdueBillCandidates(ctx context.Context, days int) ([]entity.Bill, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid FROM bills WHERE status IN ('open','partial') AND due_date < DATE_SUB(CURDATE(),INTERVAL ? DAY)`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.Bill, 0)
	for rows.Next() {
		item, err := scanBill(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Notifications) OverdueWorkOrderCandidates(ctx context.Context, days int) ([]entity.WorkOrder, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,property_id,COALESCE(tenant_id,''),COALESCE(assignee_id,''),description,status,material_cost,tenant_confirmed,created_at,updated_at FROM work_orders WHERE status IN ('reported','assigned','repairing') AND created_at < DATE_SUB(UTC_TIMESTAMP(),INTERVAL ? DAY)`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.WorkOrder, 0)
	for rows.Next() {
		var item entity.WorkOrder
		if err := rows.Scan(&item.ID, &item.PropertyID, &item.TenantID, &item.AssigneeID, &item.Description, &item.Status, &item.MaterialCost, &item.TenantConfirmed, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
