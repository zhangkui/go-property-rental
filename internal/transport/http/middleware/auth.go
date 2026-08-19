package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"go-property-rental/internal/domain/entity"
)

type ContextKey string

const IdentityKey ContextKey = "auth_identity"
const UserKey ContextKey = "user_id"

type Authenticator interface {
	Authenticate(context.Context, string) (entity.AuthIdentity, error)
}

func Auth(authenticator Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authorization := strings.TrimSpace(r.Header.Get("Authorization"))
			if !strings.HasPrefix(authorization, "Bearer ") {
				unauthorized(w, "missing bearer token")
				return
			}
			token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
			identity, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				unauthorized(w, err.Error())
				return
			}
			ctx := context.WithValue(r.Context(), IdentityKey, identity)
			ctx = context.WithValue(ctx, UserKey, identity.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Require(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			identity, ok := IdentityFromContext(r.Context())
			if !ok {
				unauthorized(w, "authentication is required")
				return
			}
			if permission != "" && !identity.HasPermission(permission) {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"status": http.StatusForbidden, "message": "permission denied", "permission": permission}})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func IdentityFromContext(ctx context.Context) (entity.AuthIdentity, bool) {
	identity, ok := ctx.Value(IdentityKey).(entity.AuthIdentity)
	return identity, ok
}

func unauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"status": http.StatusUnauthorized, "message": message}})
}
