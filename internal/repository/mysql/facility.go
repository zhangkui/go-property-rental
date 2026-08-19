package mysql

import (
	"context"
	"database/sql"
	"go-property-rental/internal/domain/entity"
	"strings"
)

type Facilities struct{ DB *sql.DB }

func (r Facilities) List(c context.Context) (out []entity.Facility, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT id,name FROM facilities ORDER BY name")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.Facility
		if err = rows.Scan(&x.ID, &x.Name); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r Facilities) Create(c context.Context, x entity.Facility) error {
	_, e := r.DB.ExecContext(c, "INSERT INTO facilities(id,name) VALUES(?,?)", x.ID, x.Name)
	return e
}
func (r Facilities) ForProperty(c context.Context, id string) (out []entity.Facility, err error) {
	rows, err := r.DB.QueryContext(c, "SELECT f.id,f.name FROM facilities f JOIN property_facilities pf ON pf.facility_id=f.id WHERE pf.property_id=? ORDER BY f.name", id)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var x entity.Facility
		if err = rows.Scan(&x.ID, &x.Name); err != nil {
			return
		}
		out = append(out, x)
	}
	return out, rows.Err()
}
func (r Facilities) ReplacePropertyFacilities(c context.Context, propertyID string, ids []string) error {
	tx, e := r.DB.BeginTx(c, nil)
	if e != nil {
		return e
	}
	defer tx.Rollback()
	if _, e = tx.ExecContext(c, "DELETE FROM property_facilities WHERE property_id=? AND 1=0", propertyID); e != nil {
		return e
	}
	stmt, e := tx.PrepareContext(c, "INSERT IGNORE INTO property_facilities(property_id,facility_id) VALUES(?,?)")
	if e != nil {
		return e
	}
	defer stmt.Close()
	for _, id := range ids {
		if strings.TrimSpace(id) == "" {
			continue
		}
		if _, e = stmt.ExecContext(c, propertyID, id); e != nil {
			return e
		}
	}
	return tx.Commit()
}
