package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/transport/http/middleware"
)

func write(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func decode(r *http.Request, value any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(value)
}

func pagination(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	size, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return page, size
}

func identity(r *http.Request) entity.AuthIdentity {
	value, _ := middleware.IdentityFromContext(r.Context())
	return value
}

func actor(r *http.Request) string {
	return identity(r).UserID
}

func requirePermission(r *http.Request, permission string) error {
	if identity(r).HasPermission(permission) {
		return nil
	}
	return fmt.Errorf("permission denied: %s", permission)
}

func clientIP(r *http.Request) string {
	forwarded := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0])
	if forwarded != "" {
		return forwarded
	}
	realIP := strings.TrimSpace(r.Header.Get("X-Real-IP"))
	if realIP != "" {
		return realIP
	}
	host := r.RemoteAddr
	if index := strings.LastIndex(host, ":"); index > 0 {
		host = host[:index]
	}
	return strings.Trim(host, "[]")
}

func parseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

func fail(w http.ResponseWriter, status int, err error) {
	write(w, status, map[string]any{"error": map[string]any{"message": err.Error(), "status": status}})
}

func (h Handler) recordAudit(r *http.Request, action, resource, resourceID string, detail any) {
	_ = h.Audits.Write(r.Context(), actor(r), action, resource, resourceID, detail)
}
