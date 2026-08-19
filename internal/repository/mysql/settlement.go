package mysql

import (
	"context"
	"database/sql"
	"errors"
	"go-property-rental/internal/domain/entity"
)

type Settlements struct{ DB *sql.DB }

func (r Settlements) Create(c context.Context, s entity.Settlement, items []entity.SettlementItem, readings []entity.MeterReading, actor string) (entity.Settlement, error) {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return s, e
	}
	defer tx.Rollback()
	var leaseStatus, propertyID string
	if e = tx.QueryRowContext(c, "SELECT status,property_id FROM leases WHERE id=? FOR UPDATE", s.LeaseID).Scan(&leaseStatus, &propertyID); e != nil {
		return s, e
	}
	if leaseStatus != "active" && leaseStatus != "terminating" && leaseStatus != "expired" {
		return s, errors.New("lease cannot be settled")
	}
	var exists int
	if e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM settlements WHERE lease_id=?", s.LeaseID).Scan(&exists); e != nil {
		return s, e
	}
	if exists > 0 {
		return s, errors.New("settlement already exists")
	}
	var total int64
	for _, item := range items {
		if item.Amount < 0 {
			return s, errors.New("settlement item must not be negative")
		}
		total += int64(item.Amount)
	}
	s.Total = entityMoney(total)
	var received, deducted, refunded int64
	if e = tx.QueryRowContext(c, "SELECT COALESCE(SUM(CASE WHEN kind IN ('collect','topup') THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='deduct' THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='refund' THEN amount ELSE 0 END),0) FROM deposit_ledgers WHERE lease_id=? FOR UPDATE", s.LeaseID).Scan(&received, &deducted, &refunded); e != nil {
		return s, e
	}
	available := received - deducted - refunded
	if int64(s.DepositDeduction) > available {
		return s, errors.New("deposit deduction exceeds balance")
	}
	refund := available - int64(s.DepositDeduction)
	s.Refund = entityMoney(refund)
	if _, e = tx.ExecContext(c, "INSERT INTO settlements(id,lease_id,status,total,deposit_deduction,refund) VALUES(?,?,?,?,?,?)", s.ID, s.LeaseID, s.Status, s.Total, s.DepositDeduction, s.Refund); e != nil {
		return s, e
	}
	for _, item := range items {
		if _, e = tx.ExecContext(c, "INSERT INTO settlement_items(id,settlement_id,kind,description,amount) VALUES(?,?,?,?,?)", item.ID, s.ID, item.Kind, item.Description, item.Amount); e != nil {
			return s, e
		}
	}
	for _, reading := range readings {
		if _, e = tx.ExecContext(c, "INSERT INTO meter_readings(id,lease_id,kind,reading,read_at) VALUES(?,?,?,?,?)", reading.ID, s.LeaseID, reading.Kind, reading.Reading, reading.ReadAt); e != nil {
			return s, e
		}
	}
	if s.DepositDeduction > 0 {
		if _, e = tx.ExecContext(c, "INSERT INTO deposit_ledgers(id,lease_id,kind,reference,amount,reason,actor_id) VALUES(UUID(),?,'deduct',?,?,?,?)", s.LeaseID, "settlement-deduct-"+s.ID, s.DepositDeduction, "move-out settlement", actor); e != nil {
			return s, e
		}
	}
	if s.Refund > 0 {
		if _, e = tx.ExecContext(c, "INSERT INTO deposit_ledgers(id,lease_id,kind,reference,amount,reason,actor_id) VALUES(UUID(),?,'refund',?,?,?,?)", s.LeaseID, "settlement-refund-"+s.ID, s.Refund, "move-out refund", actor); e != nil {
			return s, e
		}
	}
	return s, tx.Commit()
}
func (r Settlements) Get(c context.Context, id string) (s entity.Settlement, items []entity.SettlementItem, readings []entity.MeterReading, err error) {
	err = r.DB.QueryRowContext(c, "SELECT id,lease_id,status,total,deposit_deduction,refund,settled_at,created_at FROM settlements WHERE id=?", id).Scan(&s.ID, &s.LeaseID, &s.Status, &s.Total, &s.DepositDeduction, &s.Refund, &s.SettledAt, &s.CreatedAt)
	if err != nil {
		return
	}
	rows, err := r.DB.QueryContext(c, "SELECT id,settlement_id,kind,description,amount FROM settlement_items WHERE settlement_id=?", id)
	if err != nil {
		return
	}
	for rows.Next() {
		var x entity.SettlementItem
		if err = rows.Scan(&x.ID, &x.SettlementID, &x.Kind, &x.Description, &x.Amount); err != nil {
			rows.Close()
			return
		}
		items = append(items, x)
	}
	rows.Close()
	rows, err = r.DB.QueryContext(c, "SELECT id,lease_id,kind,reading,read_at FROM meter_readings WHERE lease_id=? ORDER BY read_at", s.LeaseID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.MeterReading
		if err = rows.Scan(&x.ID, &x.LeaseID, &x.Kind, &x.Reading, &x.ReadAt); err != nil {
			return
		}
		readings = append(readings, x)
	}
	return s, items, readings, rows.Err()
}
func (r Settlements) Complete(c context.Context, id, actor string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var leaseID, status string
	if e = tx.QueryRowContext(c, "SELECT lease_id,status FROM settlements WHERE id=? FOR UPDATE", id).Scan(&leaseID, &status); e != nil {
		return e
	}
	if status != "draft" && status != "reviewed" {
		return errors.New("settlement cannot be completed")
	}
	var propertyID string
	if e = tx.QueryRowContext(c, "SELECT property_id FROM leases WHERE id=? FOR UPDATE", leaseID).Scan(&propertyID); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "UPDATE settlements SET status='completed',settled_at=UTC_TIMESTAMP() WHERE id=?", id); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "UPDATE leases SET status='closed' WHERE id=?", leaseID); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "UPDATE properties SET status=?,available_from=UTC_TIMESTAMP() WHERE id=?", actor, propertyID); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO lease_state_history(id,lease_id,from_status,to_status,reason,actor_id) SELECT UUID(),id,status,'closed','move-out settlement',? FROM leases WHERE id=?", actor, leaseID); e != nil {
		return e
	}
	return tx.Commit()
}
