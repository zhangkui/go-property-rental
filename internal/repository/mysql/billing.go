package mysql

import (
	"context"
	"database/sql"
	"errors"
	"go-property-rental/internal/domain/entity"
)

type Billing struct{ DB *sql.DB }

func (r Billing) Generate(c context.Context, b entity.Bill, key string) (entity.Bill, bool, error) {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return b, false, e
	}
	defer tx.Rollback()
	var existing string
	e = tx.QueryRowContext(c, "SELECT resource_id FROM idempotency_keys WHERE scope='bill.generate.manual' AND request_key=? FOR UPDATE", key).Scan(&existing)
	if e == nil {
		found, ge := scanBill(tx.QueryRowContext(c, "SELECT id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid FROM bills WHERE id=?", existing))
		return found, false, ge
	}
	if !errors.Is(e, sql.ErrNoRows) {
		return b, false, e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO bills(id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid) VALUES(?,?,?,?,?,?,?,?,?,?)", b.ID, b.LeaseID, b.PeriodStart, b.PeriodEnd, b.DueDate, b.Status, b.Amount, b.Penalty, b.Discount, b.Paid); e != nil {
		return b, false, e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO idempotency_keys(scope,request_key,resource_id) VALUES('bill.generate',?,?)", key, b.ID); e != nil {
		return b, false, e
	}
	if e = tx.Commit(); e != nil {
		return b, false, e
	}
	return b, true, nil
}
func (r Billing) List(c context.Context, limit, offset int, leaseID, status string) (out []entity.Bill, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid FROM bills WHERE (?='' OR lease_id=?) AND (?='' OR status=?) ORDER BY period_start DESC LIMIT ? OFFSET ?", leaseID, leaseID, status, status, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		x, e := scanBill(rows)
		if e != nil {
			return nil, e
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r Billing) Get(c context.Context, id string) (entity.Bill, error) {
	return scanBill(r.DB.QueryRowContext(c, "SELECT id,lease_id,period_start,period_end,due_date,status,amount,penalty,discount,paid FROM bills WHERE id=?", id))
}

func (r Billing) Items(c context.Context, billID string) ([]entity.BillItem, error) {
	rows, err := r.DB.QueryContext(c, `SELECT id,bill_id,kind,name,amount FROM bill_items WHERE bill_id=? ORDER BY kind,name`, billID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]entity.BillItem, 0)
	for rows.Next() {
		var item entity.BillItem
		if err := rows.Scan(&item.ID, &item.BillID, &item.Kind, &item.Name, &item.Amount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

type rowScanner interface{ Scan(...any) error }

func scanBill(s rowScanner) (x entity.Bill, err error) {
	err = s.Scan(&x.ID, &x.LeaseID, &x.PeriodStart, &x.PeriodEnd, &x.DueDate, &x.Status, &x.Amount, &x.Penalty, &x.Discount, &x.Paid)
	return
}
func (r Billing) Adjust(c context.Context, a entity.BillAdjustment) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var status string
	if e = tx.QueryRowContext(c, "SELECT status FROM bills WHERE id=? FOR UPDATE", a.BillID).Scan(&status); e != nil {
		return e
	}
	if status == "paid" || status == "void" {
		return errors.New("closed bill cannot be adjusted")
	}
	column := "discount"
	if a.Kind == "penalty" {
		column = "penalty"
	} else if a.Kind != "discount" {
		return errors.New("unsupported adjustment kind")
	}
	if _, e = tx.ExecContext(c, "INSERT INTO bill_adjustments(id,bill_id,kind,amount,reason,actor_id) VALUES(?,?,?,?,?,?)", a.ID, a.BillID, a.Kind, a.Amount, a.Reason, a.ActorID); e != nil {
		return e
	}
	q := "UPDATE bills SET " + column + "=" + column + "+? WHERE id=?"
	if _, e = tx.ExecContext(c, q, a.Amount, a.BillID); e != nil {
		return e
	}
	return tx.Commit()
}
func (r Billing) RecordPayment(c context.Context, p entity.Payment, allocations []entity.PaymentAllocation) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var duplicate int
	if e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM payment_receipts WHERE reference=?", p.Reference).Scan(&duplicate); e != nil {
		return e
	}
	if duplicate > 0 {
		return nil
	}
	var sum int64
	for _, a := range allocations {
		sum += int64(a.Amount)
	}
	if sum != int64(p.Amount) {
		return errors.New("allocation total must equal payment amount")
	}
	if _, e = tx.ExecContext(c, "INSERT INTO payment_receipts(id,reference,amount,payer,paid_at,status) VALUES(?,?,?,?,?,?)", p.ID, p.Reference, p.Amount, p.Payer, p.PaidAt, p.Status); e != nil {
		return e
	}
	for _, a := range allocations {
		var amount, penalty, discount, paid int64
		var status string
		if e = tx.QueryRowContext(c, "SELECT amount,penalty,discount,paid,status FROM bills WHERE id=? FOR UPDATE", a.BillID).Scan(&amount, &penalty, &discount, &paid, &status); e != nil {
			return e
		}
		outstanding := amount + penalty - discount - paid
		if int64(a.Amount) <= 0 || int64(a.Amount) > outstanding {
			return errors.New("allocation exceeds bill outstanding")
		}
		if _, e = tx.ExecContext(c, "INSERT INTO payment_allocations(id,payment_id,bill_id,amount) VALUES(?,?,?,?)", a.ID, p.ID, a.BillID, a.Amount); e != nil {
			return e
		}
		newPaid := paid + int64(a.Amount)
		newStatus := "partial"
		if newPaid == amount+penalty-discount {
			newStatus = "paid"
		}
		if _, e = tx.ExecContext(c, "UPDATE bills SET paid=?,status=? WHERE id=?", newPaid, newStatus, a.BillID); e != nil {
			return e
		}
	}
	return tx.Commit()
}
