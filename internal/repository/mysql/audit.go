package mysql

import (
	"context"
	"database/sql"
	"go-property-rental/internal/domain/entity"
)

type Audits struct{ DB *sql.DB }

func (r Audits) Append(c context.Context, x entity.AuditLog) error {
	_, e := r.DB.ExecContext(c, "INSERT INTO audit_logs(id,actor_id,action,resource,resource_id,detail) VALUES(?,?,?,?,?,?)", x.ID, x.ActorID, x.Action, x.Resource, x.ResourceID, x.Detail)
	return e
}
func (r Audits) List(c context.Context, limit, offset int, resource, actor string) (out []entity.AuditLog, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,COALESCE(actor_id,''),action,resource,COALESCE(resource_id,''),CAST(detail AS CHAR),created_at FROM audit_logs WHERE (?='' OR resource=?) AND (?='' OR actor_id=?) ORDER BY created_at DESC LIMIT ? OFFSET ?", resource, resource, actor, actor, limit, offset)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.AuditLog
		if err = rows.Scan(&x.ID, &x.ActorID, &x.Action, &x.Resource, &x.ResourceID, &x.Detail, &x.CreatedAt); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
