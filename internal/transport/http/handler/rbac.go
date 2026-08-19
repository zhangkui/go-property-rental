package handler

import (
	"net/http"

	"go-property-rental/internal/domain/entity"
)

func (h Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, size := pagination(r)
	users, total, err := h.RBAC.ListUsers(r.Context(), page, size, r.URL.Query().Get("status"), r.URL.Query().Get("keyword"))
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, map[string]any{"items": users, "total": total, "page": page, "page_size": size})
}

func (h Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username    string   `json:"username"`
		Password    string   `json:"password"`
		DisplayName string   `json:"display_name"`
		Email       string   `json:"email"`
		RoleIDs     []string `json:"role_ids"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	user, err := h.RBAC.CreateUser(r.Context(), identity(r), input.Username, input.Password, input.DisplayName, input.Email, input.RoleIDs)
	if err != nil {
		fail(w, http.StatusConflict, err)
		return
	}
	write(w, http.StatusCreated, publicUser(user))
}

func (h Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.RBAC.UpdateProfile(r.Context(), identity(r), r.PathValue("id"), input.DisplayName, input.Email); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, input)
}

func (h Handler) SetUserStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.RBAC.SetUserStatus(r.Context(), identity(r), r.PathValue("id"), input.Status); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, input)
}

func (h Handler) ResetUserPassword(w http.ResponseWriter, r *http.Request) {
	var input struct {
		NewPassword string `json:"new_password"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.RBAC.ResetPassword(r.Context(), identity(r), r.PathValue("id"), input.NewPassword); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"status": "password_reset", "message": "all user sessions have been revoked"})
}

func (h Handler) ReplaceUserRoles(w http.ResponseWriter, r *http.Request) {
	var input struct {
		RoleIDs []string `json:"role_ids"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.RBAC.ReplaceUserRoles(r.Context(), identity(r), r.PathValue("id"), input.RoleIDs); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, input)
}

func (h Handler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.RBAC.ListRoles(r.Context())
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, map[string]any{"items": roles})
}

func (h Handler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name          string   `json:"name"`
		Description   string   `json:"description"`
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	role, err := h.RBAC.CreateRole(r.Context(), identity(r), input.Name, input.Description, input.PermissionIDs)
	if err != nil {
		fail(w, http.StatusConflict, err)
		return
	}
	write(w, http.StatusCreated, role)
}

func (h Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.RBAC.UpdateRole(r.Context(), identity(r), r.PathValue("id"), input.Name, input.Description); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, input)
}

func (h Handler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	if err := h.RBAC.DeleteRole(r.Context(), identity(r), r.PathValue("id")); err != nil {
		fail(w, http.StatusConflict, err)
		return
	}
	write(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h Handler) ReplaceRolePermissions(w http.ResponseWriter, r *http.Request) {
	var input struct {
		PermissionIDs []string `json:"permission_ids"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	if err := h.RBAC.ReplaceRolePermissions(r.Context(), identity(r), r.PathValue("id"), input.PermissionIDs); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, input)
}

func (h Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	groups, err := h.RBAC.ListPermissionGroups(r.Context())
	if err != nil {
		fail(w, http.StatusInternalServerError, err)
		return
	}
	write(w, http.StatusOK, map[string]any{"groups": groups})
}

func (h Handler) CreatePermission(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Code        string `json:"code"`
		Module      string `json:"module"`
		Description string `json:"description"`
	}
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	permission, err := h.RBAC.CreatePermission(r.Context(), identity(r), input.Code, input.Module, input.Description)
	if err != nil {
		fail(w, http.StatusConflict, err)
		return
	}
	write(w, http.StatusCreated, permission)
}

func (h Handler) UpdatePermission(w http.ResponseWriter, r *http.Request) {
	var input entity.Permission
	if err := decode(r, &input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	input.ID = r.PathValue("id")
	if err := h.RBAC.UpdatePermission(r.Context(), identity(r), input); err != nil {
		fail(w, http.StatusBadRequest, err)
		return
	}
	write(w, http.StatusOK, input)
}
