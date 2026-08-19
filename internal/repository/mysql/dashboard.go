package mysql

import (
	"context"
	"database/sql"
	"time"

	"go-property-rental/internal/domain/entity"
)

type Dashboard struct{ DB *sql.DB }

func (r Dashboard) Summary(ctx context.Context, userID string, now time.Time) (entity.DashboardSummary, error) {
	summary := entity.DashboardSummary{GeneratedAt: now.UTC()}
	loaders := []func(context.Context, *entity.DashboardSummary, string, time.Time) error{
		r.loadPropertyMetrics,
		r.loadLeaseMetrics,
		r.loadFinanceMetrics,
		r.loadOperationMetrics,
		r.loadCashflow,
		r.loadLeaseStatuses,
		r.loadUpcomingExpiries,
		r.loadTopOverdueBills,
		r.loadRecentAudits,
	}
	for _, load := range loaders {
		if err := load(ctx, &summary, userID, now.UTC()); err != nil {
			return entity.DashboardSummary{}, err
		}
	}
	return summary, nil
}

func (r Dashboard) loadPropertyMetrics(ctx context.Context, summary *entity.DashboardSummary, _ string, _ time.Time) error {
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status='available'),0),COALESCE(SUM(status='occupied'),0),COALESCE(SUM(status='maintenance'),0) FROM properties`).Scan(&summary.Properties.Total, &summary.Properties.Available, &summary.Properties.Occupied, &summary.Properties.Maintenance)
	if err == nil && summary.Properties.Total > 0 {
		summary.Properties.OccupancyBasisPoints = int64(summary.Properties.Occupied * 10000 / summary.Properties.Total)
	}
	return err
}

func (r Dashboard) loadLeaseMetrics(ctx context.Context, summary *entity.DashboardSummary, _ string, now time.Time) error {
	return r.DB.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status='draft'),0),COALESCE(SUM(status='active'),0),COALESCE(SUM(status='active' AND end_date BETWEEN ? AND DATE_ADD(?,INTERVAL 30 DAY)),0),COALESCE(SUM(EXISTS(SELECT 1 FROM lease_versions v WHERE v.lease_id=leases.id AND v.version_no>1)),0) FROM leases`, now, now).Scan(&summary.Leases.Total, &summary.Leases.Draft, &summary.Leases.Active, &summary.Leases.Expiring30Days, &summary.Leases.Renewed)
}

func (r Dashboard) loadFinanceMetrics(ctx context.Context, summary *entity.DashboardSummary, _ string, now time.Time) error {
	if err := r.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN status NOT IN ('paid','void') THEN amount+penalty-discount-paid ELSE 0 END),0),COALESCE(SUM(CASE WHEN status NOT IN ('paid','void') AND due_date<? THEN amount+penalty-discount-paid ELSE 0 END),0),COALESCE(SUM(status NOT IN ('paid','void') AND due_date<?),0),COALESCE(SUM(status NOT IN ('paid','void') AND due_date BETWEEN ? AND DATE_ADD(?,INTERVAL 7 DAY)),0) FROM bills`, now, now, now, now).Scan(&summary.Finance.Outstanding, &summary.Finance.OverdueOutstanding, &summary.Finance.OverdueBills, &summary.Finance.DueWithin7Days); err != nil {
		return err
	}
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	previousStart := monthStart.AddDate(0, -1, 0)
	if err := r.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN paid_at>=? AND paid_at<? THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN paid_at>=? AND paid_at<? THEN amount ELSE 0 END),0) FROM payment_receipts WHERE status='posted'`, monthStart, monthStart.AddDate(0, 1, 0), previousStart, monthStart).Scan(&summary.Finance.ReceivedThisMonth, &summary.Finance.ReceivedPreviousMonth); err != nil {
		return err
	}
	return r.DB.QueryRowContext(ctx, `SELECT COALESCE(SUM(CASE WHEN kind IN ('collect','topup') THEN amount WHEN kind IN ('deduct','refund') THEN -amount ELSE 0 END),0) FROM deposit_ledgers`).Scan(&summary.Finance.DepositBalance)
}

func (r Dashboard) loadOperationMetrics(ctx context.Context, summary *entity.DashboardSummary, userID string, now time.Time) error {
	return r.DB.QueryRowContext(ctx, `SELECT (SELECT COUNT(*) FROM tenants WHERE status='active'),(SELECT COUNT(*) FROM work_orders WHERE status NOT IN ('closed','cancelled')),(SELECT COUNT(*) FROM work_orders WHERE status NOT IN ('closed','cancelled') AND created_at<DATE_SUB(?,INTERVAL 3 DAY)),(SELECT COUNT(*) FROM settlements WHERE status IN ('draft','pending')),(SELECT COUNT(*) FROM approval_requests WHERE status='pending'),(SELECT COUNT(*) FROM notifications WHERE user_id=? AND status='unread')`, now, userID).Scan(&summary.Operations.ActiveTenants, &summary.Operations.OpenWorkOrders, &summary.Operations.OverdueWorkOrders, &summary.Operations.PendingSettlements, &summary.Operations.PendingApprovals, &summary.Operations.UnreadNotifications)
}

func (r Dashboard) loadCashflow(ctx context.Context, summary *entity.DashboardSummary, _ string, now time.Time) error {
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	points := make([]entity.DashboardCashflowPoint, 6)
	index := make(map[string]int, 6)
	for offset := 5; offset >= 0; offset-- {
		month := monthStart.AddDate(0, -offset, 0).Format("2006-01")
		position := 5 - offset
		points[position].Month = month
		index[month] = position
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT DATE_FORMAT(paid_at,'%Y-%m'),COALESCE(SUM(amount),0) FROM payment_receipts WHERE status='posted' AND paid_at>=? GROUP BY DATE_FORMAT(paid_at,'%Y-%m')`, monthStart.AddDate(0, -5, 0))
	if err != nil {
		return err
	}
	for rows.Next() {
		var month string
		var amount int64
		if err := rows.Scan(&month, &amount); err != nil {
			rows.Close()
			return err
		}
		if position, ok := index[month]; ok {
			points[position].Received = amount
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	rows, err = r.DB.QueryContext(ctx, `SELECT DATE_FORMAT(period_start,'%Y-%m'),COALESCE(SUM(amount+penalty-discount),0) FROM bills WHERE period_start>=? GROUP BY DATE_FORMAT(period_start,'%Y-%m')`, monthStart.AddDate(0, -5, 0))
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var month string
		var amount int64
		if err := rows.Scan(&month, &amount); err != nil {
			return err
		}
		if position, ok := index[month]; ok {
			points[position].Billed = amount
		}
	}
	summary.Cashflow = points
	return rows.Err()
}

func (r Dashboard) loadLeaseStatuses(ctx context.Context, summary *entity.DashboardSummary, _ string, _ time.Time) error {
	rows, err := r.DB.QueryContext(ctx, `SELECT status,COUNT(*) FROM leases GROUP BY status ORDER BY status`)
	if err != nil {
		return err
	}
	defer rows.Close()
	items := make([]entity.DashboardStatusPoint, 0)
	for rows.Next() {
		var item entity.DashboardStatusPoint
		if err := rows.Scan(&item.Name, &item.Value); err != nil {
			return err
		}
		items = append(items, item)
	}
	summary.LeaseStatuses = items
	return rows.Err()
}

func (r Dashboard) loadUpcomingExpiries(ctx context.Context, summary *entity.DashboardSummary, _ string, now time.Time) error {
	rows, err := r.DB.QueryContext(ctx, `SELECT l.id,CONCAT(p.building,' ',p.room),t.name,l.end_date,DATEDIFF(l.end_date,?) FROM leases l JOIN properties p ON p.id=l.property_id JOIN tenants t ON t.id=l.tenant_id WHERE l.status='active' AND l.end_date BETWEEN ? AND DATE_ADD(?,INTERVAL 30 DAY) ORDER BY l.end_date LIMIT 6`, now, now, now)
	if err != nil {
		return err
	}
	defer rows.Close()
	items := make([]entity.DashboardUpcomingLease, 0)
	for rows.Next() {
		var item entity.DashboardUpcomingLease
		if err := rows.Scan(&item.LeaseID, &item.PropertyLabel, &item.TenantName, &item.EndDate, &item.DaysRemaining); err != nil {
			return err
		}
		items = append(items, item)
	}
	summary.UpcomingExpiries = items
	return rows.Err()
}

func (r Dashboard) loadTopOverdueBills(ctx context.Context, summary *entity.DashboardSummary, _ string, now time.Time) error {
	rows, err := r.DB.QueryContext(ctx, `SELECT b.id,CONCAT(p.building,' ',p.room),t.name,b.due_date,b.amount+b.penalty-b.discount-b.paid,DATEDIFF(?,b.due_date) FROM bills b JOIN leases l ON l.id=b.lease_id JOIN properties p ON p.id=l.property_id JOIN tenants t ON t.id=l.tenant_id WHERE b.status NOT IN ('paid','void') AND b.due_date<? AND b.amount+b.penalty-b.discount-b.paid>0 ORDER BY b.amount+b.penalty-b.discount-b.paid DESC LIMIT 6`, now, now)
	if err != nil {
		return err
	}
	defer rows.Close()
	items := make([]entity.DashboardOverdueBill, 0)
	for rows.Next() {
		var item entity.DashboardOverdueBill
		if err := rows.Scan(&item.BillID, &item.PropertyLabel, &item.TenantName, &item.DueDate, &item.Outstanding, &item.OverdueDays); err != nil {
			return err
		}
		items = append(items, item)
	}
	summary.TopOverdueBills = items
	return rows.Err()
}

func (r Dashboard) loadRecentAudits(ctx context.Context, summary *entity.DashboardSummary, _ string, _ time.Time) error {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,COALESCE(actor_id,''),action,resource,COALESCE(resource_id,''),CAST(detail AS CHAR),created_at FROM audit_logs ORDER BY created_at DESC LIMIT 8`)
	if err != nil {
		return err
	}
	defer rows.Close()
	items := make([]entity.AuditLog, 0)
	for rows.Next() {
		var item entity.AuditLog
		if err := rows.Scan(&item.ID, &item.ActorID, &item.Action, &item.Resource, &item.ResourceID, &item.Detail, &item.CreatedAt); err != nil {
			return err
		}
		items = append(items, item)
	}
	summary.RecentAuditActions = items
	return rows.Err()
}
