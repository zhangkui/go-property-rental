package httptransport

import (
	"go-property-rental/internal/transport/http/handler"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	r := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	Router(handler.Handler{}).ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("status=%d", w.Code)
	}
}
func TestProtectedRoute(t *testing.T) {
	r := httptest.NewRequest("GET", "/api/tenants", nil)
	w := httptest.NewRecorder()
	Router(handler.Handler{}).ServeHTTP(w, r)
	if w.Code != 401 {
		t.Fatalf("status=%d", w.Code)
	}
}
