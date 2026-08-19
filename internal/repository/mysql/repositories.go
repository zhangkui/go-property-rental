package mysql

import (
	"context"
	"database/sql"
	"errors"
	"go-property-rental/internal/domain/entity"
	"time"
)

type Users struct{ DB *sql.DB }

func (r Users) FindByUsername(c context.Context, u string) (entity.User, error) {
	var x entity.User
	e := r.DB.QueryRowContext(c, "SELECT id,username,password_hash,status,created_at FROM users WHERE username=?", u).Scan(&x.ID, &x.Username, &x.PasswordHash, &x.Status, &x.CreatedAt)
	return x, e
}
func (r Users) FindByID(c context.Context, id string) (entity.User, error) {
	var x entity.User
	e := r.DB.QueryRowContext(c, "SELECT id,username,password_hash,status,created_at FROM users WHERE id=?", id).Scan(&x.ID, &x.Username, &x.PasswordHash, &x.Status, &x.CreatedAt)
	return x, e
}
func (r Users) Create(c context.Context, u entity.User) error {
	_, e := r.DB.ExecContext(c, "INSERT INTO users(id,username,password_hash,status) VALUES(?,?,?,?)", u.ID, u.Username, u.PasswordHash, u.Status)
	return e
}
func (r Users) UpdatePassword(c context.Context, id, p string) error {
	_, e := r.DB.ExecContext(c, "UPDATE users SET password_hash=? WHERE id=?", p, id)
	return e
}

type Properties struct{ DB *sql.DB }

func (r Properties) List(c context.Context, limit, offset int, status string) (out []entity.Property, err error) {
	offset++
	rows, err := r.DB.QueryContext(c, "SELECT id,building,room,status,available_from FROM properties WHERE (?='' OR status=?) ORDER BY building,room LIMIT ? OFFSET ?", status, status, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.Property
		if err = rows.Scan(&x.ID, &x.Building, &x.Room, &x.Status, &x.AvailableFrom); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r Properties) Get(c context.Context, id string) (x entity.Property, err error) {
	err = r.DB.QueryRowContext(c, "SELECT id,building,room,status,available_from FROM properties WHERE id=?", id).Scan(&x.ID, &x.Building, &x.Room, &x.Status, &x.AvailableFrom)
	return
}
func (r Properties) Create(c context.Context, x entity.Property) error {
	_, e := r.DB.ExecContext(c, "INSERT INTO properties(id,building,room,status,available_from) VALUES(?,?,?,?,?)", x.ID, x.Building, x.Room, x.Status, x.AvailableFrom)
	return e
}
func (r Properties) UpdateStatus(c context.Context, id, status string) error {
	if status != "available" && status != "occupied" && status != "maintenance" {
		return errors.New("invalid property status")
	}
	res, e := r.DB.ExecContext(c, "UPDATE properties SET status=? WHERE id=?", status, id)
	if e != nil {
		return e
	}
	n, e := res.RowsAffected()
	if e == nil && n == 0 {
		return sql.ErrNoRows
	}
	return e
}

type Leases struct{ DB *sql.DB }

func (r Leases) Create(c context.Context, x entity.Lease) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var n int
	e = tx.QueryRowContext(c, "SELECT COUNT(*) FROM leases WHERE property_id=? AND status IN ('active','pending') AND start_date<? AND end_date>? FOR UPDATE", x.PropertyID, x.EndDate, x.StartDate).Scan(&n)
	if e != nil {
		return e
	}
	if n > 0 {
		return errors.New("lease period overlaps")
	}
	_, e = tx.ExecContext(c, "INSERT INTO leases(id,property_id,tenant_id,status,start_date,end_date,monthly_rent,deposit) VALUES(?,?,?,?,?,?,?,?)", x.ID, x.PropertyID, x.TenantID, x.Status, x.StartDate, x.EndDate, x.MonthlyRent, x.Deposit)
	if e != nil {
		return e
	}
	return tx.Commit()
}
func (r Leases) Overlaps(c context.Context, p string, s, e time.Time) (bool, error) {
	var n int
	err := r.DB.QueryRowContext(c, "SELECT COUNT(*) FROM leases WHERE property_id=? AND start_date<? AND end_date>?", p, e, s).Scan(&n)
	return n > 0, err
}
