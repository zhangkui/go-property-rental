package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var bug006Seq atomic.Uint64

type bug006Driver struct{}

func (bug006Driver) Open(string) (driver.Conn, error) { return &bug006Conn{}, nil }

type bug006Conn struct{}

func (*bug006Conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (*bug006Conn) Close() error                        { return nil }
func (*bug006Conn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (*bug006Conn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	v1 := []driver.Value{"v1", "lease-1", int64(1), int64(200000), int64(200000), now, now.AddDate(1, 0, 0), now}
	v2 := []driver.Value{"v2", "lease-1", int64(2), int64(250000), int64(300000), now.AddDate(1, 0, 0), now.AddDate(2, 0, 0), now.AddDate(1, 0, 0)}
	rows := [][]driver.Value{v1, v2}
	if strings.Contains(q, "DESC") {
		rows = [][]driver.Value{v2, v1}
	}
	return &bug006Rows{rows: rows}, nil
}

type bug006Rows struct {
	rows  [][]driver.Value
	index int
}

func (*bug006Rows) Columns() []string {
	return []string{"id", "lease_id", "version_no", "monthly_rent", "deposit", "start_date", "end_date", "created_at"}
}
func (*bug006Rows) Close() error { return nil }
func (r *bug006Rows) Next(dest []driver.Value) error {
	if r.index >= len(r.rows) {
		return io.EOF
	}
	copy(dest, r.rows[r.index])
	r.index++
	return nil
}
func bug006DB(t *testing.T) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("bug006-%d", bug006Seq.Add(1))
	sql.Register(name, bug006Driver{})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
func TestBug006_BusinessRegression(t *testing.T) {
	repo := mysqlrepo.LeaseStore{DB: bug006DB(t)}
	versions, err := (service.LeaseService{Repo: repo}).Versions(context.Background(), "lease-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 2 || versions[0].VersionNo != 2 || versions[1].VersionNo != 1 {
		t.Fatalf("latest version was not first: %#v", versions)
	}
	if int64(versions[0].MonthlyRent) != 250000 || int64(versions[0].Deposit) != 300000 {
		t.Fatalf("latest version money mismatch: %#v", versions[0])
	}
}

var _ driver.QueryerContext = (*bug006Conn)(nil)
