package mysql

import (
	"context"
	"database/sql"
	"time"

	"go-property-rental/internal/domain/entity"
)

type Reports struct{ DB *sql.DB }

func (r Reports) Occupancy(ctx context.Context, filter entity.ReportFilter) (entity.OccupancyReport, error) {
	var report entity.OccupancyReport
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status IN ('occupied','leased')),0),COALESCE(SUM(status IN ('vacant','available')),0),COALESCE(SUM(status='maintenance'),0) FROM properties WHERE (?='' OR id=?)`, filter.PropertyID, filter.PropertyID).Scan(&report.TotalUnits, &report.OccupiedUnits, &report.VacantUnits, &report.MaintenanceUnits)
	if err != nil {
		return report, err
	}
	if report.TotalUnits > 0 {
		report.OccupancyBasisPoints = int64(report.OccupiedUnits * 10000 / report.TotalUnits)
	}
	return report, nil
}

func (r Reports) RentRoll(ctx context.Context, filter entity.ReportFilter) ([]entity.RentRollRow, error) {
	query := `SELECT p.id,p.building,p.room,l.id,t.name,l.status,l.start_date,l.end_date,l.monthly_rent,l.deposit FROM leases l JOIN properties p ON p.id=l.property_id JOIN tenants t ON t.id=l.tenant_id WHERE (?='' OR p.id=?) AND (? IS NULL OR l.end_date>?) AND (? IS NULL OR l.start_date<=?) ORDER BY p.building,p.room,l.start_date`
	rows, err := r.DB.QueryContext(ctx, query, filter.PropertyID, filter.PropertyID, filter.StartDate, filter.StartDate, filter.EndDate, filter.EndDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.RentRollRow, 0)
	for rows.Next() {
		var item entity.RentRollRow
		if err := rows.Scan(&item.PropertyID, &item.Building, &item.Room, &item.LeaseID, &item.TenantName, &item.LeaseStatus, &item.StartDate, &item.EndDate, &item.MonthlyRent, &item.Deposit); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Reports) ReceivableAging(ctx context.Context, filter entity.ReportFilter, at time.Time) ([]entity.ReceivableAgingRow, error) {
	query := `SELECT l.id,t.name,CONCAT(p.building,' ',p.room),SUM(CASE WHEN DATEDIFF(?,b.due_date)<=0 THEN b.amount+b.penalty-b.discount-b.paid ELSE 0 END),SUM(CASE WHEN DATEDIFF(?,b.due_date) BETWEEN 1 AND 30 THEN b.amount+b.penalty-b.discount-b.paid ELSE 0 END),SUM(CASE WHEN DATEDIFF(?,b.due_date) BETWEEN 31 AND 60 THEN b.amount+b.penalty-b.discount-b.paid ELSE 0 END),SUM(CASE WHEN DATEDIFF(?,b.due_date) BETWEEN 61 AND 90 THEN b.amount+b.penalty-b.discount-b.paid ELSE 0 END),SUM(CASE WHEN DATEDIFF(?,b.due_date)>90 THEN b.amount+b.penalty-b.discount-b.paid ELSE 0 END) FROM bills b JOIN leases l ON l.id=b.lease_id JOIN tenants t ON t.id=l.tenant_id JOIN properties p ON p.id=l.property_id WHERE b.status IN ('open','partial') AND (?='' OR p.id=?) GROUP BY l.id,t.name,p.building,p.room ORDER BY l.id`
	rows, err := r.DB.QueryContext(ctx, query, at, at, at, at, at, filter.PropertyID, filter.PropertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.ReceivableAgingRow, 0)
	for rows.Next() {
		var item entity.ReceivableAgingRow
		if err := rows.Scan(&item.LeaseID, &item.TenantName, &item.PropertyLabel, &item.Current, &item.Days1To30, &item.Days31To60, &item.Days61To90, &item.DaysOver90); err != nil {
			return nil, err
		}
		item.Total = item.Current + item.Days1To30 + item.Days31To60 + item.Days61To90 + item.DaysOver90
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Reports) DepositReconciliation(ctx context.Context, filter entity.ReportFilter) ([]entity.DepositReconciliationRow, error) {
	query := `SELECT l.id,t.name,CONCAT(p.building,' ',p.room),l.deposit,COALESCE(SUM(CASE WHEN d.kind IN ('collect','top_up','refund') THEN d.amount ELSE -d.amount END),0) FROM leases l JOIN tenants t ON t.id=l.tenant_id JOIN properties p ON p.id=l.property_id LEFT JOIN deposit_ledgers d ON d.lease_id=l.id WHERE (?='' OR p.id=?) GROUP BY l.id,t.name,p.building,p.room,l.deposit ORDER BY p.building,p.room`
	rows, err := r.DB.QueryContext(ctx, query, filter.PropertyID, filter.PropertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.DepositReconciliationRow, 0)
	for rows.Next() {
		var item entity.DepositReconciliationRow
		if err := rows.Scan(&item.LeaseID, &item.TenantName, &item.PropertyLabel, &item.ContractDeposit, &item.LedgerBalance); err != nil {
			return nil, err
		}
		item.Difference = item.ContractDeposit - item.LedgerBalance
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r Reports) MaintenanceSLA(ctx context.Context, filter entity.ReportFilter, at time.Time) (entity.MaintenanceSLAReport, error) {
	var report entity.MaintenanceSLAReport
	err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*),COALESCE(SUM(status IN ('reported','assigned','repairing')),0),COALESCE(SUM(status='closed'),0),COALESCE(SUM(status IN ('reported','assigned','repairing') AND created_at<?),0),CAST(COALESCE(AVG(CASE WHEN status='closed' THEN TIMESTAMPDIFF(HOUR,created_at,updated_at) END),0) AS SIGNED) FROM work_orders WHERE (?='' OR property_id=?)`, at.Add(-24*time.Hour), filter.PropertyID, filter.PropertyID).Scan(&report.Total, &report.Open, &report.Completed, &report.Overdue, &report.AverageResolutionHours)
	return report, err
}
