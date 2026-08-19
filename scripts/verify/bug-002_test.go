package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

type facilityStoreStub struct {
	current    []entity.Facility
	replaced   []string
	propertyID string
}

func (s *facilityStoreStub) List(context.Context) ([]entity.Facility, error) { return nil, nil }
func (s *facilityStoreStub) Create(context.Context, entity.Facility) error   { return nil }
func (s *facilityStoreStub) ForProperty(context.Context, string) ([]entity.Facility, error) {
	return s.current, nil
}
func (s *facilityStoreStub) ReplacePropertyFacilities(_ context.Context, id string, ids []string) error {
	s.propertyID = id
	s.replaced = append([]string(nil), ids...)
	return nil
}

type facilitySQLState struct {
	mu           sync.Mutex
	execQueries  []string
	preparedSQL  string
	preparedArgs [][]driver.Value
	committed    bool
}

type facilityDriver struct{ state *facilitySQLState }
type facilityConn struct{ state *facilitySQLState }
type facilityTx struct{ state *facilitySQLState }
type facilityStmt struct{ state *facilitySQLState }

var facilityDriverSequence atomic.Uint64

func (d facilityDriver) Open(string) (driver.Conn, error) { return &facilityConn{state: d.state}, nil }
func (c *facilityConn) Prepare(query string) (driver.Stmt, error) {
	c.state.mu.Lock()
	c.state.preparedSQL = normalizeFacilitySQL(query)
	c.state.mu.Unlock()
	return &facilityStmt{state: c.state}, nil
}
func (c *facilityConn) Close() error              { return nil }
func (c *facilityConn) Begin() (driver.Tx, error) { return &facilityTx{state: c.state}, nil }
func (c *facilityConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	return &facilityTx{state: c.state}, nil
}
func (c *facilityConn) ExecContext(_ context.Context, query string, _ []driver.NamedValue) (driver.Result, error) {
	c.state.mu.Lock()
	c.state.execQueries = append(c.state.execQueries, normalizeFacilitySQL(query))
	c.state.mu.Unlock()
	return driver.RowsAffected(1), nil
}
func (t *facilityTx) Commit() error {
	t.state.mu.Lock()
	t.state.committed = true
	t.state.mu.Unlock()
	return nil
}
func (t *facilityTx) Rollback() error { return nil }
func (s *facilityStmt) Close() error  { return nil }
func (s *facilityStmt) NumInput() int { return 2 }
func (s *facilityStmt) Exec(args []driver.Value) (driver.Result, error) {
	s.state.mu.Lock()
	s.state.preparedArgs = append(s.state.preparedArgs, append([]driver.Value(nil), args...))
	s.state.mu.Unlock()
	return driver.RowsAffected(1), nil
}
func (s *facilityStmt) Query([]driver.Value) (driver.Rows, error) {
	return nil, fmt.Errorf("unexpected query")
}

func normalizeFacilitySQL(query string) string { return strings.Join(strings.Fields(query), " ") }

func openFacilityDB(t *testing.T) (*sql.DB, *facilitySQLState) {
	t.Helper()
	state := &facilitySQLState{}
	driverName := fmt.Sprintf("bug002-facility-%d", facilityDriverSequence.Add(1))
	sql.Register(driverName, facilityDriver{state: state})
	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db, state
}

func TestBug002_BusinessRegression(t *testing.T) {
	t.Run("replacement discards stale facilities", func(t *testing.T) {
		s := &facilityStoreStub{current: []entity.Facility{{ID: "old-a"}, {ID: "old-b"}}}
		err := (service.FacilityService{Repo: s}).Replace(context.Background(), "property-1", []string{"new-c"})
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(s.replaced, []string{"new-c"}) {
			t.Fatalf("replacement retained stale facilities: %#v", s.replaced)
		}
	})

	t.Run("repository deletes old links and inserts exact replacement", func(t *testing.T) {
		db, state := openFacilityDB(t)
		err := (mysqlrepo.Facilities{DB: db}).ReplacePropertyFacilities(context.Background(), "property-1", []string{"new-c"})
		if err != nil {
			t.Fatal(err)
		}
		state.mu.Lock()
		defer state.mu.Unlock()
		if !reflect.DeepEqual(state.execQueries, []string{"DELETE FROM property_facilities WHERE property_id=?"}) {
			t.Fatalf("replacement did not delete existing links: %#v", state.execQueries)
		}
		if state.preparedSQL != "INSERT INTO property_facilities(property_id,facility_id) VALUES(?,?)" {
			t.Fatalf("replacement used non-strict insert: %q", state.preparedSQL)
		}
		if !reflect.DeepEqual(state.preparedArgs, [][]driver.Value{{"property-1", "new-c"}}) {
			t.Fatalf("replacement inserted unexpected links: %#v", state.preparedArgs)
		}
		if !state.committed {
			t.Fatal("replacement transaction was not committed")
		}
	})

	t.Run("empty replacement clears all facilities", func(t *testing.T) {
		db, state := openFacilityDB(t)
		err := (mysqlrepo.Facilities{DB: db}).ReplacePropertyFacilities(context.Background(), "property-1", nil)
		if err != nil {
			t.Fatal(err)
		}
		state.mu.Lock()
		defer state.mu.Unlock()
		if !reflect.DeepEqual(state.execQueries, []string{"DELETE FROM property_facilities WHERE property_id=?"}) {
			t.Fatalf("empty replacement did not clear existing links: %#v", state.execQueries)
		}
		if len(state.preparedArgs) != 0 {
			t.Fatalf("empty replacement inserted facilities: %#v", state.preparedArgs)
		}
		if !state.committed {
			t.Fatal("empty replacement transaction was not committed")
		}
	})
}
