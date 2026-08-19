package mysql

import (
	"context"
	"database/sql"
	"errors"
	"go-property-rental/internal/domain/entity"
	"time"
)

type LeaseStore struct{ DB *sql.DB }

func (r LeaseStore) List(c context.Context, limit, offset int, propertyID, status string) (out []entity.Lease, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit FROM leases WHERE (?='' OR property_id=?) AND (?='' OR status=?) ORDER BY start_date DESC LIMIT ? OFFSET ?", propertyID, propertyID, status, status, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.Lease
		if err = rows.Scan(&x.ID, &x.PropertyID, &x.TenantID, &x.Status, &x.StartDate, &x.EndDate, &x.MonthlyRent, &x.Deposit); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r LeaseStore) Get(c context.Context, id string) (x entity.Lease, err error) {
	err = r.DB.QueryRowContext(c, "SELECT id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit FROM leases WHERE id=?", id).Scan(&x.ID, &x.PropertyID, &x.TenantID, &x.Status, &x.StartDate, &x.EndDate, &x.MonthlyRent, &x.Deposit)
	return
}
func (r LeaseStore) Create(c context.Context, x entity.Lease, v entity.LeaseVersion, occupants []entity.LeaseOccupant, actor string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var n int
	if e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM leases WHERE property_id=? AND status IN ('draft','pending','active') AND start_date<? AND end_date>? FOR UPDATE", x.PropertyID, x.EndDate, x.StartDate).Scan(&n); e != nil {
		return e
	}
	if n > 0 {
		return errors.New("lease period overlaps")
	}
	if _, e = tx.ExecContext(c, "INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES(?,?,?,?,?,?,?,?)", x.ID, x.PropertyID, x.TenantID, x.Status, x.StartDate, x.EndDate, x.MonthlyRent, x.Deposit); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO lease_versions(id,lease_id,version_no,monthly_rent,deposit,start_date,end_date) VALUES(?,?,?,?,?,?,?)", v.ID, x.ID, v.VersionNo, v.MonthlyRent, v.Deposit, v.StartDate, v.EndDate); e != nil {
		return e
	}
	for _, o := range occupants {
		if _, e = tx.ExecContext(c, "INSERT INTO lease_occupants(id,lease_id,name,phone,identity_no,is_primary) VALUES(?,?,?,?,?,?)", o.ID, x.ID, o.Name, o.Phone, o.IdentityNo, o.Primary); e != nil {
			return e
		}
	}
	if _, e = tx.ExecContext(c, "INSERT INTO lease_state_history(id,lease_id,from_status,to_status,reason,actor_id) VALUES(UUID(),?,'',?,'created',?)", x.ID, x.Status, actor); e != nil {
		return e
	}
	return tx.Commit()
}
func (r LeaseStore) Renew(c context.Context, v entity.LeaseVersion, newEnd time.Time, actor string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var propertyID, status string
	var currentEnd time.Time
	if e = tx.QueryRowContext(c, "SELECT property_id,status,end_date FROM leases WHERE id=? FOR UPDATE", v.LeaseID).Scan(&propertyID, &status, &currentEnd); e != nil {
		return e
	}
	if status != "active" {
		return errors.New("only active lease can be renewed")
	}
	var n int
	if e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM leases WHERE property_id=? AND id<>? AND status IN ('draft','pending','active') AND start_date<? AND end_date>?", propertyID, v.LeaseID, newEnd, currentEnd).Scan(&n); e != nil {
		return e
	}
	if n > 0 {
		return errors.New("renewal period overlaps")
	}
	var version int
	if e = tx.QueryRowContext(c, "SELECT COALESCE(MAX(version_no),0)+1 FROM lease_versions WHERE lease_id=?", v.LeaseID).Scan(&version); e != nil {
		return e
	}
	v.VersionNo = version
	v.StartDate = currentEnd
	v.EndDate = newEnd
	if _, e = tx.ExecContext(c, "INSERT INTO lease_versions(id,lease_id,version_no,monthly_rent,deposit,start_date,end_date) VALUES(?,?,?,?,?,?,?)", v.ID, v.LeaseID, v.VersionNo, v.MonthlyRent, v.Deposit, v.StartDate, v.EndDate); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "UPDATE leases SET end_date=?,monthly_rent=?,deposit=? WHERE id=?", newEnd, v.Deposit, v.Deposit, v.LeaseID); e != nil {
		return e
	}
	_, e = tx.ExecContext(c, "INSERT INTO lease_state_history(id,lease_id,from_status,to_status,reason,actor_id) VALUES(UUID(),?,?,?,'renewed',?)", v.LeaseID, status, status, actor)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (r LeaseStore) Versions(c context.Context, id string) (out []entity.LeaseVersion, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,lease_id,version_no,monthly_rent,deposit,start_date,end_date,created_at FROM lease_versions WHERE lease_id=? ORDER BY version_no ASC", id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.LeaseVersion
		if err = rows.Scan(&x.ID, &x.LeaseID, &x.VersionNo, &x.MonthlyRent, &x.Deposit, &x.StartDate, &x.EndDate, &x.CreatedAt); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r LeaseStore) ChangeStatus(c context.Context, id, to, reason, actor string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var from string
	if e = tx.QueryRowContext(c, "SELECT status FROM leases WHERE id=? FOR UPDATE", id).Scan(&from); e != nil {
		return e
	}
	if !validLeaseTransition(from, to) {
		return errors.New("illegal lease status transition")
	}
	if _, e = tx.ExecContext(c, "UPDATE leases SET status=? WHERE id=?", to, id); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO lease_state_history(id,lease_id,from_status,to_status,reason,actor_id) VALUES(UUID(),?,?,?,?,?)", id, from, from, reason, actor); e != nil {
		return e
	}
	return tx.Commit()
}
func validLeaseTransition(from, to string) bool {
	allowed := map[string]map[string]bool{"draft": {"pending": true, "cancelled": true}, "pending": {"active": true, "cancelled": true}, "active": {"terminating": true, "expired": true}, "terminating": {"closed": true}, "expired": {"closed": true}}
	return allowed[from][to]
}
