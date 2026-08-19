package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
	"io"
	"strings"
	"sync"
	"testing"
	"time"
)

type bug005LeaseStore struct {
	lease   entity.Lease
	version entity.LeaseVersion
	end     time.Time
	calls   int
}

type bug005SQLState struct {
	currentEnd     time.Time
	versionRent    int64
	versionDeposit int64
	leaseRent      int64
	leaseDeposit   int64
}

type bug005Driver struct{}
type bug005Conn struct{ state *bug005SQLState }
type bug005Tx struct{}
type bug005Rows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

var (
	bug005RegisterDriver sync.Once
	bug005ActiveSQLState *bug005SQLState
)

func (*bug005LeaseStore) List(context.Context, int, int, string, string) ([]entity.Lease, error) {
	return nil, nil
}
func (s *bug005LeaseStore) Get(context.Context, string) (entity.Lease, error) { return s.lease, nil }
func (*bug005LeaseStore) Create(context.Context, entity.Lease, entity.LeaseVersion, []entity.LeaseOccupant, string) error {
	return nil
}
func (s *bug005LeaseStore) Renew(_ context.Context, v entity.LeaseVersion, end time.Time, _ string) error {
	s.calls++
	s.version = v
	s.end = end
	return nil
}
func (*bug005LeaseStore) Versions(context.Context, string) ([]entity.LeaseVersion, error) {
	return nil, nil
}
func (*bug005LeaseStore) ChangeStatus(context.Context, string, string, string, string) error {
	return nil
}

func (bug005Driver) Open(string) (driver.Conn, error) {
	return &bug005Conn{state: bug005ActiveSQLState}, nil
}
func (*bug005Conn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("prepare is not supported")
}
func (*bug005Conn) Close() error              { return nil }
func (*bug005Conn) Begin() (driver.Tx, error) { return bug005Tx{}, nil }
func (*bug005Conn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return bug005Tx{}, nil
}
func (bug005Tx) Commit() error          { return nil }
func (bug005Tx) Rollback() error        { return nil }
func (r *bug005Rows) Columns() []string { return r.columns }
func (*bug005Rows) Close() error        { return nil }
func (r *bug005Rows) Next(destination []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(destination, r.values[r.index])
	r.index++
	return nil
}
func (c *bug005Conn) QueryContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Rows, error) {
	switch {
	case strings.Contains(query, "SELECT property_id,status,end_date"):
		return &bug005Rows{columns: []string{"property_id", "status", "end_date"}, values: [][]driver.Value{{"property-1", "active", c.state.currentEnd}}}, nil
	case strings.Contains(query, "SELECT COUNT(*)"):
		return &bug005Rows{columns: []string{"count"}, values: [][]driver.Value{{int64(0)}}}, nil
	case strings.Contains(query, "SELECT COALESCE(MAX(version_no),0)+1"):
		return &bug005Rows{columns: []string{"version_no"}, values: [][]driver.Value{{int64(2)}}}, nil
	default:
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
}
func (c *bug005Conn) ExecContext(_ context.Context, query string, arguments []driver.NamedValue) (driver.Result, error) {
	switch {
	case strings.Contains(query, "INSERT INTO lease_versions"):
		c.state.versionRent = arguments[3].Value.(int64)
		c.state.versionDeposit = arguments[4].Value.(int64)
	case strings.Contains(query, "UPDATE leases SET end_date"):
		c.state.leaseRent = arguments[1].Value.(int64)
		c.state.leaseDeposit = arguments[2].Value.(int64)
	case strings.Contains(query, "INSERT INTO lease_state_history"):
	default:
		return nil, fmt.Errorf("unexpected exec: %s", query)
	}
	return driver.RowsAffected(1), nil
}

func TestBug005_BusinessRegression(t *testing.T) {
	old := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
	next := time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC)
	store := &bug005LeaseStore{lease: entity.Lease{ID: "lease-1", EndDate: old}}
	if err := (service.LeaseService{Repo: store}).Renew(context.Background(), "lease-1", next, 120000, 50000, "admin"); err != nil {
		t.Fatal(err)
	}
	if int64(store.version.MonthlyRent) != 120000 || int64(store.version.Deposit) != 50000 || !store.end.Equal(next) {
		t.Fatalf("renewal values crossed: %#v", store.version)
	}
	if err := (service.LeaseService{Repo: store}).Renew(context.Background(), "lease-1", old, 120000, 50000, "admin"); err == nil {
		t.Fatal("expected non-increasing end date to fail")
	}
	if store.calls != 1 {
		t.Fatalf("invalid renewal reached repository, calls=%d", store.calls)
	}

	bug005RegisterDriver.Do(func() { sql.Register("bug005-runtime", bug005Driver{}) })
	state := &bug005SQLState{currentEnd: old}
	bug005ActiveSQLState = state
	database, err := sql.Open("bug005-runtime", "")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	version := entity.LeaseVersion{ID: "version-2", LeaseID: "lease-1", MonthlyRent: 120000, Deposit: 50000}
	if err := (mysqlrepo.LeaseStore{DB: database}).Renew(context.Background(), version, next, "admin"); err != nil {
		t.Fatal(err)
	}
	if state.versionRent != 120000 || state.versionDeposit != 50000 {
		t.Fatalf("version values crossed: rent=%d deposit=%d", state.versionRent, state.versionDeposit)
	}
	if state.leaseRent != 120000 || state.leaseDeposit != 50000 {
		t.Fatalf("lease values crossed: rent=%d deposit=%d", state.leaseRent, state.leaseDeposit)
	}
}
