package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

var bug003Seq atomic.Uint64

type bug003Driver struct{ status string }

func (d bug003Driver) Open(string) (driver.Conn, error) { return &bug003Conn{status: d.status}, nil }

type bug003Conn struct{ status string }

func (*bug003Conn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (*bug003Conn) Close() error                        { return nil }
func (*bug003Conn) Begin() (driver.Tx, error)           { return nil, errors.New("unsupported") }
func (c *bug003Conn) QueryContext(_ context.Context, q string, _ []driver.NamedValue) (driver.Rows, error) {
	status := c.status
	if status == "disabled" && strings.Contains(q, "CASE WHEN status='disabled'") {
		status = "suspended"
	}
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &bug003Rows{values: []driver.Value{"tenant-1", "Tenant", "", "", "", status, now, now}}, nil
}

type bug003Rows struct {
	values []driver.Value
	done   bool
}

func (*bug003Rows) Columns() []string {
	return []string{"id", "name", "phone", "email", "identity_no", "status", "created_at", "updated_at"}
}
func (*bug003Rows) Close() error { return nil }
func (r *bug003Rows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	copy(dest, r.values)
	r.done = true
	return nil
}

type bug003LeaseStore struct{ creates int }

func (*bug003LeaseStore) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (*bug003LeaseStore) Get(context.Context, string) (entity.Lease, error) {
	return entity.Lease{}, nil
}
func (s *bug003LeaseStore) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	s.creates++
	return nil
}
func (*bug003LeaseStore) Renew(context.Context, entity.LeaseVersion, time.Time, string) error {
	return nil
}
func (*bug003LeaseStore) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (*bug003LeaseStore) ChangeStatus(context.Context, string, string, string, string) error {
	return nil
}
func bug003DB(t *testing.T, status string) *sql.DB {
	t.Helper()
	name := fmt.Sprintf("bug003-%d", bug003Seq.Add(1))
	sql.Register(name, bug003Driver{status: status})
	db, err := sql.Open(name, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}
func TestBug003_BusinessRegression(t *testing.T) {
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	lease := entity.Lease{PropertyID: "property-1", TenantID: "tenant-1", StartDate: start, EndDate: start.AddDate(1, 0, 0), MonthlyRent: 200000, Deposit: 200000}
	disabledStore := &bug003LeaseStore{}
	disabledService := service.LeaseService{Repo: disabledStore, Tenants: mysqlrepo.Tenants{DB: bug003DB(t, "disabled")}}
	if _, err := disabledService.Create(context.Background(), lease, nil, "admin"); err == nil {
		t.Fatal("disabled tenant was allowed to create a lease")
	}
	if disabledStore.creates != 0 {
		t.Fatalf("disabled tenant created %d lease rows", disabledStore.creates)
	}
	activeStore := &bug003LeaseStore{}
	activeService := service.LeaseService{Repo: activeStore, Tenants: mysqlrepo.Tenants{DB: bug003DB(t, "active")}}
	if _, err := activeService.Create(context.Background(), lease, nil, "admin"); err != nil {
		t.Fatalf("active tenant was rejected: %v", err)
	}
	if activeStore.creates != 1 {
		t.Fatalf("active tenant expected one lease create, got %d", activeStore.creates)
	}
}

var _ driver.QueryerContext = (*bug003Conn)(nil)
