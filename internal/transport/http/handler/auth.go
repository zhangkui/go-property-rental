package handler

import (
	"net/http"

	"go-property-rental/internal/domain/entity"
	"go-property-rental/internal/service"
)

func (h Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.Auth.Register(r.Context(), input.Username, input.Password, input.DisplayName, input.Email)
	if err != nil {
		fail(w, http.StatusConflict, err)
		return
	}
	write(w, http.StatusCreated, map[string]any{"user": publicUser(user)})
}

func (h Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	user, pair, err := h.Auth.Login(r.Context(), input.Username, input.Password, service.LoginMetadata{IPAddress: clientIP(r), UserAgent: r.UserAgent()})
	if err != nil {
		fail(w, http.StatusUnauthorized, err)
		return
	}
	write(w, http.StatusOK, map[string]any{
		"token":              pair.AccessToken,
		"access_token":       pair.AccessToken,
		"refresh_token":      pair.RefreshToken,
		"access_expires_at":  pair.AccessExpiresAt,
		"refresh_expires_at": pair.RefreshExpiresAt,
		"user":               publicUser(user),
	})
}

func (h Handler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	pair, err := h.Auth.Refresh(r.Context(), input.RefreshToken)
	if err != nil {
		fail(w, http.StatusUnauthorized, err)
		return
	}
	write(w, http.StatusOK, map[string]any{
		"token":              pair.AccessToken,
		"access_token":       pair.AccessToken,
		"refresh_token":      pair.RefreshToken,
		"access_expires_at":  pair.AccessExpiresAt,
		"refresh_expires_at": pair.RefreshExpiresAt,
	})
}

func (h Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if err := h.Auth.Logout(r.Context(), identity(r)); err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"status": "logged_out"})
}

func (h Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.Auth.CurrentUser(r.Context(), identity(r))
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	permissions := make([]string, 0, len(user.Permissions))
	for _, permission := range user.Permissions {
		permissions = append(permissions, permission.Code)
	}
	write(w, http.StatusOK, map[string]any{
		"id":            user.ID,
		"username":      user.Username,
		"display_name":  user.DisplayName,
		"email":         user.Email,
		"status":        user.Status,
		"roles":         user.Roles,
		"permissions":   permissions,
		"created_at":    user.CreatedAt,
		"last_login_at": user.LastLoginAt,
	})
}

func (h Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.Auth.ChangePassword(r.Context(), identity(r), input.CurrentPassword, input.NewPassword); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"status": "password_changed", "message": "all sessions have been revoked"})
}

func publicUser(user entity.User) map[string]any {
	return map[string]any{
		"id":            user.ID,
		"username":      user.Username,
		"display_name":  user.DisplayName,
		"email":         user.Email,
		"status":        user.Status,
		"created_at":    user.CreatedAt,
		"updated_at":    user.UpdatedAt,
		"last_login_at": user.LastLoginAt,
	}
}
