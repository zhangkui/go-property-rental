package mysql

import (
	"context"
	"database/sql"
	"errors"
	"go-property-rental/internal/domain/entity"
)

type Deposits struct{ DB *sql.DB }

func (r Deposits) Append(c context.Context, x entity.DepositLedger) error {
	if x.Amount <= 0 {
		return errors.New("deposit amount must be positive")
	}
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var duplicate int
	if e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM deposit_ledgers WHERE lease_id=? AND reference=?", x.LeaseID, x.Reference).Scan(&duplicate); e != nil {
		return e
	}
	if duplicate > 0 {
		return nil
	}
	var received, deducted, refunded int64
	if e = tx.QueryRowContext(c, "SELECT COALESCE(SUM(CASE WHEN kind IN ('collect','topup') THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='deduct' THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='refund' THEN amount ELSE 0 END),0) FROM deposit_ledgers WHERE lease_id=? FOR UPDATE", x.LeaseID).Scan(&received, &deducted, &refunded); e != nil {
		return e
	}
	available := received - deducted - refunded
	if (x.Kind == "deduct" || x.Kind == "refund") && int64(x.Amount) > available {
		return errors.New("insufficient deposit balance")
	}
	if x.Kind != "collect" && x.Kind != "topup" && x.Kind != "deduct" && x.Kind != "refund" {
		return errors.New("unsupported deposit transaction")
	}
	_, e = tx.ExecContext(c, "INSERT INTO deposit_ledgers(id,lease_id,kind,reference,amount,reason,actor_id) VALUES(?,?,?,?,?,?,?)", x.ID, x.LeaseID, x.Kind, x.Reference, x.Amount, x.Reason, x.ActorID)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (r Deposits) List(c context.Context, leaseID string) (out []entity.DepositLedger, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,lease_id,kind,reference,amount,COALESCE(reason,''),COALESCE(actor_id,''),created_at FROM deposit_ledgers WHERE lease_id=? ORDER BY created_at", leaseID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.DepositLedger
		if err = rows.Scan(&x.ID, &x.LeaseID, &x.Kind, &x.Reference, &x.Amount, &x.Reason, &x.ActorID, &x.CreatedAt); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r Deposits) Balance(c context.Context, leaseID string) (x entity.DepositBalance, err error) {
	x.LeaseID = leaseID
	err = r.DB.QueryRowContext(c, "SELECT COALESCE(SUM(CASE WHEN kind IN ('collect','topup') THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='deduct' THEN amount ELSE 0 END),0),COALESCE(SUM(CASE WHEN kind='refund' THEN amount ELSE 0 END),0) FROM deposit_ledgers WHERE lease_id=?", leaseID).Scan(&x.Received, &x.Deducted, &x.Refunded)
	return
}
