package mysql

import (
	"context"
	"database/sql"
	"errors"
	"go-property-rental/internal/domain/entity"
)

type WorkOrders struct{ DB *sql.DB }

func (r WorkOrders) List(c context.Context, limit, offset int, propertyID, status string) (out []entity.WorkOrder, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,property_id,COALESCE(tenant_id,''),COALESCE(assignee_id,''),description,status,material_cost,tenant_confirmed,created_at,updated_at FROM work_orders WHERE (?='' OR property_id=?) AND (?='' OR status=?) ORDER BY created_at DESC LIMIT ? OFFSET ?", propertyID, propertyID, status, status, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.WorkOrder
		if err = rows.Scan(&x.ID, &x.PropertyID, &x.TenantID, &x.AssigneeID, &x.Description, &x.Status, &x.MaterialCost, &x.TenantConfirmed, &x.CreatedAt, &x.UpdatedAt); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r WorkOrders) Get(c context.Context, id string) (x entity.WorkOrder, materials []entity.WorkOrderMaterial, err error) {
	err = r.DB.QueryRowContext(c, "SELECT id,property_id,COALESCE(tenant_id,''),COALESCE(assignee_id,''),description,status,material_cost,tenant_confirmed,created_at,updated_at FROM work_orders WHERE id=?", id).Scan(&x.ID, &x.PropertyID, &x.TenantID, &x.AssigneeID, &x.Description, &x.Status, &x.MaterialCost, &x.TenantConfirmed, &x.CreatedAt, &x.UpdatedAt)
	if err != nil {
		return
	}
	rows, err := r.DB.QueryContext(c, "SELECT id,work_order_id,name,quantity,unit_cost FROM work_order_materials WHERE work_order_id=?", id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var m entity.WorkOrderMaterial
		if err = rows.Scan(&m.ID, &m.WorkOrderID, &m.Name, &m.Quantity, &m.UnitCost); err != nil {
			return
		}
		materials = append(materials, m)
	}
	return x, materials, rows.Err()
}
func (r WorkOrders) Create(c context.Context, x entity.WorkOrder) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(c, "INSERT INTO work_orders(id,property_id,tenant_id,assignee_id,description,status,material_cost,tenant_confirmed,created_at,updated_at) VALUES(?,?,?,?,?,?,0,FALSE,UTC_TIMESTAMP(),UTC_TIMESTAMP())", x.ID, x.PropertyID, nullString(x.TenantID), nullString(x.AssigneeID), x.Description, x.Status); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO work_order_state_history(id,work_order_id,from_status,to_status,reason,actor_id) VALUES(UUID(),?,'',?,'created',NULL)", x.ID, x.Status); e != nil {
		return e
	}
	return tx.Commit()
}
func (r WorkOrders) Assign(c context.Context, id, assignee, actor string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var status string
	if e = tx.QueryRowContext(c, "SELECT status FROM work_orders WHERE id=? FOR UPDATE", id).Scan(&status); e != nil {
		return e
	}
	if status != "reported" && status != "assigned" {
		return errors.New("work order cannot be assigned")
	}
	if _, e = tx.ExecContext(c, "UPDATE work_orders SET assignee_id=?,status='assigned',updated_at=UTC_TIMESTAMP() WHERE id=?", assignee, id); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO work_order_state_history(id,work_order_id,from_status,to_status,reason,actor_id) VALUES(UUID(),?,?,'assigned','assigned',?)", id, status, actor); e != nil {
		return e
	}
	return tx.Commit()
}
func (r WorkOrders) AddMaterial(c context.Context, m entity.WorkOrderMaterial) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var status string
	if e = tx.QueryRowContext(c, "SELECT status FROM work_orders WHERE id=? FOR UPDATE", m.WorkOrderID).Scan(&status); e != nil {
		return e
	}
	if status != "assigned" && status != "repairing" {
		return errors.New("materials can only be added during repair")
	}
	if m.Quantity <= 0 || m.UnitCost < 0 {
		return errors.New("invalid material quantity or cost")
	}
	if _, e = tx.ExecContext(c, "INSERT INTO work_order_materials(id,work_order_id,name,quantity,unit_cost) VALUES(?,?,?,?,?)", m.ID, m.WorkOrderID, m.Name, m.Quantity, m.UnitCost); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "UPDATE work_orders SET material_cost=material_cost+?,updated_at=UTC_TIMESTAMP() WHERE id=?", m.Total(), m.WorkOrderID); e != nil {
		return e
	}
	return tx.Commit()
}
func (r WorkOrders) Transition(c context.Context, id, to, reason, actor string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	var from string
	if e = tx.QueryRowContext(c, "SELECT status FROM work_orders WHERE id=? FOR UPDATE", id).Scan(&from); e != nil {
		return e
	}
	if !validWorkTransition(from, to) {
		return errors.New("illegal work order transition")
	}
	if _, e = tx.ExecContext(c, "UPDATE work_orders SET status=?,updated_at=UTC_TIMESTAMP() WHERE id=?", to, id); e != nil {
		return e
	}
	if _, e = tx.ExecContext(c, "INSERT INTO work_order_state_history(id,work_order_id,from_status,to_status,reason,actor_id) VALUES(UUID(),?,?,?,?,?)", id, from, to, reason, actor); e != nil {
		return e
	}
	return tx.Commit()
}
func (r WorkOrders) Confirm(c context.Context, id string) error {
	res, e := r.DB.ExecContext(c, "UPDATE work_orders SET tenant_confirmed=TRUE,status='closed',updated_at=UTC_TIMESTAMP() WHERE id=? AND status='completed'", id)
	if e != nil {
		return e
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("only completed work order can be confirmed")
	}
	return nil
}
func validWorkTransition(from, to string) bool {
	m := map[string]map[string]bool{"reported": {"assigned": true, "cancelled": true}, "assigned": {"repairing": true, "cancelled": true}, "repairing": {"completed": true}, "completed": {"closed": true}}
	return m[from][to]
}
func nullString(v string) any {
	if v == "" {
		return nil
	}
	return v
}
