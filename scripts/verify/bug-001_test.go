package verify

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"go-property-rental/internal/domain/entity"
	mysqlrepo "go-property-rental/internal/repository/mysql"
	"go-property-rental/internal/service"
	"go-property-rental/internal/transport/http/handler"
)

var (
	errInvalidPropertyStatus = errors.New("invalid property status")
	driverSequence           atomic.Uint64
)

type propertyRepositoryStub struct {
	status      string
	updateCalls int
}

func (r *propertyRepositoryStub) List(context.Context, int, int, string) ([]entity.Property, error) {
	return nil, nil
}

func (r *propertyRepositoryStub) Get(context.Context, string) (entity.Property, error) {
	return entity.Property{}, nil
}

func (r *propertyRepositoryStub) Create(context.Context, entity.Property) error {
	return nil
}

func (r *propertyRepositoryStub) UpdateStatus(_ context.Context, _ string, status string) error {
	r.updateCalls++
	if status != "available" && status != "occupied" && status != "maintenance" {
		return errInvalidPropertyStatus
	}
	r.status = status
	return nil
}

type auditStoreStub struct {
	entries []entity.AuditLog
}

func (s *auditStoreStub) Append(_ context.Context, entry entity.AuditLog) error {
	s.entries = append(s.entries, entry)
	return nil
}

func (s *auditStoreStub) List(context.Context, int, int, string, string) ([]entity.AuditLog, error) {
	return nil, nil
}

type recordingSQLState struct {
	execCalls int
	query     string
	args      []driver.NamedValue
}

type recordingDriver struct {
	state *recordingSQLState
}

func (d recordingDriver) Open(string) (driver.Conn, error) {
	return &recordingConn{state: d.state}, nil
}

type recordingConn struct {
	state *recordingSQLState
}

func (c *recordingConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare is not supported")
}

func (c *recordingConn) Close() error {
	return nil
}

func (c *recordingConn) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not supported")
}

func (c *recordingConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	c.state.execCalls++
	c.state.query = query
	c.state.args = append([]driver.NamedValue(nil), args...)
	return driver.RowsAffected(1), nil
}

func TestBug001_BusinessRegression(t *testing.T) {
	t.Run("service does not normalize an invalid status", func(t *testing.T) {
		repository := &propertyRepositoryStub{status: "occupied"}
		propertyService := service.PropertyService{Repo: repository}

		err := propertyService.ChangeStatus(context.Background(), "property-1", "retired")

		if !errors.Is(err, errInvalidPropertyStatus) {
			t.Fatalf("expected invalid status error, got %v", err)
		}
		if repository.status != "occupied" {
			t.Fatalf("invalid status changed stored state to %q", repository.status)
		}
		if repository.updateCalls != 1 {
			t.Fatalf("expected one repository validation call, got %d", repository.updateCalls)
		}
	})

	t.Run("repository rejects invalid status before SQL", func(t *testing.T) {
		state := &recordingSQLState{}
		repository := mysqlrepo.Properties{DB: openRecordingDB(t, state)}

		err := repository.UpdateStatus(context.Background(), "property-1", "retired")

		if err == nil {
			t.Fatal("expected repository to reject invalid status")
		}
		if state.execCalls != 0 {
			t.Fatalf("invalid status executed %d SQL updates", state.execCalls)
		}
	})

	t.Run("handler rejects invalid status without success audit", func(t *testing.T) {
		repository := &propertyRepositoryStub{status: "occupied"}
		audits := &auditStoreStub{}
		h := handler.Handler{
			Properties: service.PropertyService{Repo: repository},
			Audits:     service.AuditService{Repo: audits},
		}
		request := statusRequest(t, "retired")
		response := httptest.NewRecorder()

		h.ChangePropertyStatus(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected HTTP 400, got %d body=%s", response.Code, response.Body.String())
		}
		if repository.status != "occupied" {
			t.Fatalf("invalid request changed stored status to %q", repository.status)
		}
		if len(audits.entries) != 0 {
			t.Fatalf("invalid request created %d success audits", len(audits.entries))
		}
	})

	t.Run("maintenance remains valid and audited", func(t *testing.T) {
		state := &recordingSQLState{}
		repository := mysqlrepo.Properties{DB: openRecordingDB(t, state)}
		if err := repository.UpdateStatus(context.Background(), "property-1", "maintenance"); err != nil {
			t.Fatalf("valid repository update failed: %v", err)
		}
		if state.execCalls != 1 {
			t.Fatalf("expected one SQL update, got %d", state.execCalls)
		}
		if state.query != "UPDATE properties SET status=? WHERE id=?" {
			t.Fatalf("unexpected update query %q", state.query)
		}
		if len(state.args) != 2 || state.args[0].Value != "maintenance" || state.args[1].Value != "property-1" {
			t.Fatalf("unexpected SQL arguments: %#v", state.args)
		}

		serviceRepository := &propertyRepositoryStub{status: "occupied"}
		audits := &auditStoreStub{}
		h := handler.Handler{
			Properties: service.PropertyService{Repo: serviceRepository},
			Audits:     service.AuditService{Repo: audits},
		}
		response := httptest.NewRecorder()
		h.ChangePropertyStatus(response, statusRequest(t, "maintenance"))

		if response.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200, got %d body=%s", response.Code, response.Body.String())
		}
		if serviceRepository.status != "maintenance" {
			t.Fatalf("valid status was not persisted: %q", serviceRepository.status)
		}
		if len(audits.entries) != 1 || audits.entries[0].Action != "property.status_changed" {
			t.Fatalf("expected one property.status_changed audit, got %#v", audits.entries)
		}
	})
}

func statusRequest(t *testing.T, status string) *http.Request {
	t.Helper()
	request := httptest.NewRequest(http.MethodPatch, "/api/properties/property-1/status", strings.NewReader(`{"status":"`+status+`"}`))
	request.SetPathValue("id", "property-1")
	return request
}

func openRecordingDB(t *testing.T, state *recordingSQLState) *sql.DB {
	t.Helper()
	driverName := fmt.Sprintf("bug001-%d", driverSequence.Add(1))
	sql.Register(driverName, recordingDriver{state: state})
	db, err := sql.Open(driverName, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

var _ driver.ExecerContext = (*recordingConn)(nil)
