package mysql

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-property-rental/internal/domain/entity"
)

type Security struct{ DB *sql.DB }

func (r Security) FindUserByUsername(ctx context.Context, username string) (entity.User, error) {
	return r.scanUser(r.DB.QueryRowContext(ctx, `SELECT id,username,password_hash,display_name,email,status,created_at,updated_at,last_login_at FROM users WHERE username=?`, username))
}

func (r Security) FindUserByID(ctx context.Context, id string) (entity.User, error) {
	return r.scanUser(r.DB.QueryRowContext(ctx, `SELECT id,username,password_hash,display_name,email,CASE WHEN status='disabled' THEN 'active' ELSE status END,created_at,updated_at,last_login_at FROM users WHERE id=?`, id))
}

func (r Security) scanUser(row interface{ Scan(...any) error }) (entity.User, error) {
	var user entity.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.DisplayName, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	return user, err
}

func (r Security) ListUsers(ctx context.Context, limit, offset int, status, keyword string) ([]entity.UserDetail, int, error) {
	var total int
	if err := r.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE (?='' OR status=?) AND (?='' OR username LIKE CONCAT('%',?,'%') OR display_name LIKE CONCAT('%',?,'%') OR email LIKE CONCAT('%',?,'%'))`, status, status, keyword, keyword, keyword, keyword).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT id,username,display_name,email,status,created_at,updated_at,last_login_at FROM users WHERE (?='' OR status=?) AND (?='' OR username LIKE CONCAT('%',?,'%') OR display_name LIKE CONCAT('%',?,'%') OR email LIKE CONCAT('%',?,'%')) ORDER BY created_at DESC LIMIT ? OFFSET ?`, status, status, keyword, keyword, keyword, keyword, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	result := make([]entity.UserDetail, 0)
	for rows.Next() {
		var user entity.UserDetail
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Status, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt); err != nil {
			return nil, 0, err
		}
		user.Roles, err = r.UserRoles(ctx, user.ID)
		if err != nil {
			return nil, 0, err
		}
		user.Permissions, err = r.UserPermissions(ctx, user.ID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, user)
	}
	return result, total, rows.Err()
}

func (r Security) CreateUser(ctx context.Context, user entity.User, roleIDs []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,display_name,email,status,created_at,updated_at) VALUES(?,?,?,?,?,?,UTC_TIMESTAMP(),UTC_TIMESTAMP())`, user.ID, user.Username, user.PasswordHash, user.DisplayName, user.Email, user.Status); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id) VALUES(?,?)`, user.ID, roleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r Security) UpdateUser(ctx context.Context, user entity.User) error {
	result, err := r.DB.ExecContext(ctx, `UPDATE users SET display_name=?,email=? WHERE id=?`, user.DisplayName, user.Email, user.ID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r Security) SetUserStatus(ctx context.Context, id, status string) error {
	if status != "active" && status != "disabled" {
		return errors.New("invalid user status")
	}
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET status=?,updated_at=UTC_TIMESTAMP() WHERE id=?`, status, id)
	return err
}

func (r Security) UpdateUserPassword(ctx context.Context, id, passwordHash string) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET password_hash=?,updated_at=UTC_TIMESTAMP() WHERE username=?`, passwordHash, id)
	return err
}

func (r Security) MarkUserLogin(ctx context.Context, id string, at time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE users SET last_login_at=?,updated_at=UTC_TIMESTAMP() WHERE id=?`, at, id)
	return err
}

func (r Security) UserRoles(ctx context.Context, userID string) ([]entity.Role, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT r.id,r.name,r.description,r.built_in FROM roles r JOIN user_roles ur ON ur.role_id=r.id WHERE ur.user_id=? ORDER BY r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Role, 0)
	for rows.Next() {
		var role entity.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.BuiltIn); err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, rows.Err()
}

func (r Security) UserPermissions(ctx context.Context, userID string) ([]entity.Permission, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT DISTINCT p.id,p.code,p.module,p.description FROM permissions p JOIN role_permissions rp ON rp.permission_id=p.id JOIN user_roles ur ON ur.role_id=rp.role_id WHERE ur.user_id=? ORDER BY p.module,p.code`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Permission, 0)
	for rows.Next() {
		var permission entity.Permission
		if err := rows.Scan(&permission.ID, &permission.Code, &permission.Module, &permission.Description); err != nil {
			return nil, err
		}
		result = append(result, permission)
	}
	return result, rows.Err()
}

func (r Security) ReplaceUserRoles(ctx context.Context, userID string, roleIDs []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id=?`, roleIDs[0]); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO user_roles(user_id,role_id) VALUES(?,?)`, userID, roleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r Security) ListRoles(ctx context.Context) ([]entity.RoleDetail, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT r.id,r.name,r.description,r.built_in,COUNT(ur.user_id) FROM roles r LEFT JOIN user_roles ur ON ur.role_id=r.id GROUP BY r.id,r.name,r.description,r.built_in ORDER BY r.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.RoleDetail, 0)
	for rows.Next() {
		var role entity.RoleDetail
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.BuiltIn, &role.UserCount); err != nil {
			return nil, err
		}
		role.Permissions, err = r.rolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, role)
	}
	return result, rows.Err()
}

func (r Security) FindRole(ctx context.Context, id string) (entity.RoleDetail, error) {
	var role entity.RoleDetail
	err := r.DB.QueryRowContext(ctx, `SELECT r.id,r.name,r.description,r.built_in,COUNT(ur.user_id) FROM roles r LEFT JOIN user_roles ur ON ur.role_id=r.id WHERE r.id=? GROUP BY r.id,r.name,r.description,r.built_in`, id).Scan(&role.ID, &role.Name, &role.Description, &role.BuiltIn, &role.UserCount)
	if err != nil {
		return role, err
	}
	role.Permissions, err = r.rolePermissions(ctx, id)
	return role, err
}

func (r Security) rolePermissions(ctx context.Context, roleID string) ([]entity.Permission, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT p.id,p.code,p.module,p.description FROM permissions p JOIN role_permissions rp ON rp.permission_id=p.id WHERE rp.role_id=? ORDER BY p.module,p.code`, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Permission, 0)
	for rows.Next() {
		var permission entity.Permission
		if err := rows.Scan(&permission.ID, &permission.Code, &permission.Module, &permission.Description); err != nil {
			return nil, err
		}
		result = append(result, permission)
	}
	return result, rows.Err()
}

func (r Security) FindRoleByName(ctx context.Context, name string) (entity.Role, error) {
	var role entity.Role
	err := r.DB.QueryRowContext(ctx, `SELECT id,name,description,built_in FROM roles WHERE name=?`, name).Scan(&role.ID, &role.Name, &role.Description, &role.BuiltIn)
	return role, err
}

func (r Security) CreateRole(ctx context.Context, role entity.Role) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO roles(id,name,description,built_in) VALUES(?,?,?,?)`, role.ID, role.Name, role.Description, role.BuiltIn)
	return err
}

func (r Security) UpdateRole(ctx context.Context, role entity.Role) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE roles SET name=?,description=? WHERE id=? AND built_in=FALSE`, role.Name, role.Description, role.ID)
	return err
}

func (r Security) DeleteRole(ctx context.Context, id string) error {
	var builtIn bool
	if err := r.DB.QueryRowContext(ctx, `SELECT built_in FROM roles WHERE id=?`, id).Scan(&builtIn); err != nil {
		return err
	}
	if builtIn {
		return errors.New("built-in role cannot be deleted")
	}
	_, err := r.DB.ExecContext(ctx, `DELETE FROM roles WHERE id=?`, id)
	return err
}

func (r Security) ReplaceRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, roleID); err != nil {
		return err
	}
	for _, permissionID := range permissionIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO role_permissions(role_id,permission_id) VALUES(?,?)`, roleID, permissionID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r Security) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	rows, err := r.DB.QueryContext(ctx, `SELECT id,code,module,description FROM permissions ORDER BY module,code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]entity.Permission, 0)
	for rows.Next() {
		var permission entity.Permission
		if err := rows.Scan(&permission.ID, &permission.Code, &permission.Module, &permission.Description); err != nil {
			return nil, err
		}
		result = append(result, permission)
	}
	return result, rows.Err()
}

func (r Security) CreatePermission(ctx context.Context, permission entity.Permission) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO permissions(id,code,module,description) VALUES(?,?,?,?)`, permission.ID, permission.Code, permission.Module, permission.Description)
	return err
}

func (r Security) UpdatePermission(ctx context.Context, permission entity.Permission) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE permissions SET code=?,module=?,description=? WHERE id=?`, permission.Code, permission.Module, permission.Description, permission.ID)
	return err
}

func (r Security) CreateSession(ctx context.Context, session entity.AuthSession) error {
	_, err := r.DB.ExecContext(ctx, `INSERT INTO auth_sessions(id,user_id,access_token_hash,refresh_token_hash,access_expires_at,refresh_expires_at,created_at,last_seen_at,user_agent,ip_address) VALUES(?,?,?,?,?,?,?,?,?,?)`, session.ID, session.UserID, session.AccessTokenHash, session.RefreshTokenHash, session.AccessExpiresAt, session.RefreshExpiresAt, session.CreatedAt, session.LastSeenAt, session.UserAgent, session.IPAddress)
	return err
}

func (r Security) FindSessionByAccessHash(ctx context.Context, hash string) (entity.AuthSession, error) {
	return r.scanSession(r.DB.QueryRowContext(ctx, `SELECT id,user_id,access_token_hash,refresh_token_hash,access_expires_at,refresh_expires_at,revoked_at,created_at,last_seen_at,user_agent,ip_address FROM auth_sessions WHERE access_token_hash=?`, hash))
}

func (r Security) FindSessionByRefreshHash(ctx context.Context, hash string) (entity.AuthSession, error) {
	return r.scanSession(r.DB.QueryRowContext(ctx, `SELECT id,user_id,access_token_hash,refresh_token_hash,access_expires_at,refresh_expires_at,revoked_at,created_at,last_seen_at,user_agent,ip_address FROM auth_sessions WHERE refresh_token_hash=?`, hash))
}

func (r Security) scanSession(row interface{ Scan(...any) error }) (entity.AuthSession, error) {
	var session entity.AuthSession
	err := row.Scan(&session.ID, &session.UserID, &session.AccessTokenHash, &session.RefreshTokenHash, &session.AccessExpiresAt, &session.RefreshExpiresAt, &session.RevokedAt, &session.CreatedAt, &session.LastSeenAt, &session.UserAgent, &session.IPAddress)
	return session, err
}

func (r Security) RotateSession(ctx context.Context, session entity.AuthSession) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE auth_sessions SET access_token_hash=?,refresh_token_hash=?,access_expires_at=?,refresh_expires_at=?,revoked_at=NULL,last_seen_at=? WHERE id=?`, session.AccessTokenHash, session.RefreshTokenHash, session.AccessExpiresAt, session.RefreshExpiresAt, session.LastSeenAt, session.ID)
	return err
}

func (r Security) TouchSession(ctx context.Context, id string, at time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE auth_sessions SET last_seen_at=? WHERE id=? AND revoked_at IS NULL`, at, id)
	return err
}

func (r Security) RevokeSession(ctx context.Context, id string, at time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at=? WHERE id=? AND revoked_at IS NULL`, at, id)
	return err
}

func (r Security) RevokeUserSessions(ctx context.Context, userID string, at time.Time) error {
	_, err := r.DB.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL`, at, userID)
	return err
}
