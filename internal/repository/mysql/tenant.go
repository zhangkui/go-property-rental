package mysql

import (
	"context"
	"database/sql"
	"go-property-rental/internal/domain/entity"
)

type Tenants struct{ DB *sql.DB }

func (r Tenants) List(c context.Context, limit, offset int, status string) (out []entity.Tenant, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,name,phone,email,identity_no,status,created_at,updated_at FROM tenants WHERE (?='' OR status=?) ORDER BY created_at DESC LIMIT ? OFFSET ?", status, status, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.Tenant
		if err = rows.Scan(&x.ID, &x.Name, &x.Phone, &x.Email, &x.IdentityNo, &x.Status, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r Tenants) Get(c context.Context, id string) (x entity.Tenant, err error) {
	err = r.DB.QueryRowContext(c, "SELECT id,name,phone,email,identity_no,CASE WHEN status='disabled' THEN 'suspended' ELSE status END,created_at,updated_at FROM tenants WHERE id=?", id).Scan(&x.ID, &x.Name, &x.Phone, &x.Email, &x.IdentityNo, &x.Status, &x.CreatedAt, &x.UpdatedAt)
	return
}
func (r Tenants) Create(c context.Context, x entity.Tenant) error {
	_, e := r.DB.ExecContext(c, "INSERT INTO tenants(id,name,phone,email,identity_no,status,created_at,updated_at) VALUES(?,?,?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP())", x.ID, x.Name, x.Phone, x.Email, x.IdentityNo, x.Status)
	return e
}
func (r Tenants) Update(c context.Context, x entity.Tenant) error {
	_, e := r.DB.ExecContext(c, "UPDATE tenants SET name=?,phone=?,email=?,identity_no=?,updated_at=UTC_TIMESTAMP() WHERE id=?", x.Name, x.Phone, x.Email, x.IdentityNo, x.ID)
	return e
}
func (r Tenants) SetStatus(c context.Context, id, status string) error {
	_, e := r.DB.ExecContext(c, "UPDATE tenants SET status=?,updated_at=UTC_TIMESTAMP() WHERE id=?", status, id)
	return e
}
