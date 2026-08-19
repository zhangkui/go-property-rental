package bootstrap

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type permissionSeed struct {
	Code        string
	Module      string
	Description string
}

type roleSeed struct {
	Name        string
	Description string
	Permissions []string
}

func Admin(ctx context.Context, db *sql.DB, username, password string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	permissions := defaultPermissions()
	for _, permission := range permissions {
		if _, err = tx.ExecContext(ctx, `INSERT INTO permissions(id,code,module,description) VALUES(UUID(),?,?,?) ON DUPLICATE KEY UPDATE module=VALUES(module),description=VALUES(description)`, permission.Code, permission.Module, permission.Description); err != nil {
			return fmt.Errorf("seed permission %s: %w", permission.Code, err)
		}
	}
	roles := defaultRoles(permissions)
	for _, role := range roles {
		if _, err = tx.ExecContext(ctx, `INSERT INTO roles(id,name,description,built_in) VALUES(UUID(),?,?,TRUE) ON DUPLICATE KEY UPDATE description=VALUES(description),built_in=TRUE`, role.Name, role.Description); err != nil {
			return fmt.Errorf("seed role %s: %w", role.Name, err)
		}
		var roleID string
		if err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name=?`, role.Name).Scan(&roleID); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, roleID); err != nil {
			return err
		}
		for _, code := range role.Permissions {
			if _, err = tx.ExecContext(ctx, `INSERT INTO role_permissions(role_id,permission_id) SELECT ?,id FROM permissions WHERE code=?`, roleID, code); err != nil {
				return fmt.Errorf("assign permission %s to %s: %w", code, role.Name, err)
			}
		}
	}
	var userID string
	err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE username=?`, username).Scan(&userID)
	if err == sql.ErrNoRows {
		hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			return hashErr
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO users(id,username,password_hash,display_name,email,status,created_at,updated_at) VALUES(UUID(),?,?,?,'','active',UTC_TIMESTAMP(),UTC_TIMESTAMP())`, username, string(hash), "Local administrator"); err != nil {
			return err
		}
		if err = tx.QueryRowContext(ctx, `SELECT id FROM users WHERE username=?`, username).Scan(&userID); err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	var adminRoleID string
	if err = tx.QueryRowContext(ctx, `SELECT id FROM roles WHERE name='business_admin'`).Scan(&adminRoleID); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `INSERT IGNORE INTO user_roles(user_id,role_id) VALUES(?,?)`, userID, adminRoleID); err != nil {
		return err
	}
	return tx.Commit()
}

func defaultPermissions() []permissionSeed {
	return []permissionSeed{
		{"dashboard:read", "dashboard", "View operational dashboard"},
		{"property:read", "property", "View properties"}, {"property:create", "property", "Create properties"}, {"property:update", "property", "Update properties"}, {"property:status", "property", "Change property status"},
		{"facility:read", "facility", "View facilities"}, {"facility:write", "facility", "Manage facilities"},
		{"tenant:read", "tenant", "View tenants"}, {"tenant:create", "tenant", "Create tenants"}, {"tenant:update", "tenant", "Update tenants"}, {"tenant:status", "tenant", "Change tenant status"},
		{"lease:read", "lease", "View leases"}, {"lease:create", "lease", "Create leases"}, {"lease:submit", "lease", "Submit lease changes"}, {"lease:approve", "lease", "Approve leases"}, {"lease:renew", "lease", "Renew leases"}, {"lease:terminate", "lease", "Terminate leases"},
		{"billing:read", "billing", "View bills"}, {"billing:generate", "billing", "Generate bills"}, {"billing:adjust", "billing", "Adjust bills"}, {"payment:post", "billing", "Post payments"},
		{"deposit:read", "deposit", "View deposits"}, {"deposit:collect", "deposit", "Collect deposits"}, {"deposit:deduct", "deposit", "Deduct deposits"}, {"deposit:refund", "deposit", "Refund deposits"},
		{"maintenance:read", "maintenance", "View work orders"}, {"maintenance:create", "maintenance", "Create work orders"}, {"maintenance:assign", "maintenance", "Assign work orders"}, {"maintenance:work", "maintenance", "Process work orders"}, {"maintenance:confirm", "maintenance", "Confirm work orders"},
		{"settlement:read", "settlement", "View settlements"}, {"settlement:create", "settlement", "Create settlements"}, {"settlement:approve", "settlement", "Approve settlements"},
		{"audit:read", "audit", "View audit logs"}, {"report:read", "report", "View reports"}, {"report:export", "report", "Export reports"},
		{"approval:read", "approval", "View approvals"}, {"approval:submit", "approval", "Submit approvals"}, {"approval:review", "approval", "Review approvals"},
		{"notification:read", "notification", "View notifications"}, {"notification:manage", "notification", "Manage reminders"},
		{"user:read", "security", "View users"}, {"user:create", "security", "Create users"}, {"user:update", "security", "Update users"}, {"user:status", "security", "Change user status"}, {"user:reset_password", "security", "Reset passwords"}, {"user:assign_role", "security", "Assign user roles"},
		{"role:read", "security", "View roles"}, {"role:write", "security", "Manage roles"}, {"permission:read", "security", "View permissions"}, {"permission:write", "security", "Manage permissions"},
	}
}

func defaultRoles(permissions []permissionSeed) []roleSeed {
	all := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		all = append(all, permission.Code)
	}
	operator := []string{"dashboard:read", "property:read", "property:create", "property:update", "property:status", "facility:read", "facility:write", "tenant:read", "tenant:create", "tenant:update", "tenant:status", "lease:read", "lease:create", "lease:submit", "lease:renew", "lease:terminate", "billing:read", "billing:generate", "payment:post", "deposit:read", "deposit:collect", "maintenance:read", "maintenance:create", "maintenance:work", "settlement:read", "settlement:create", "notification:read"}
	reviewer := []string{"dashboard:read", "property:read", "tenant:read", "lease:read", "lease:approve", "billing:read", "billing:adjust", "deposit:read", "deposit:deduct", "deposit:refund", "maintenance:read", "maintenance:assign", "maintenance:confirm", "settlement:read", "settlement:approve", "approval:read", "approval:review", "audit:read", "report:read"}
	auditor := []string{"dashboard:read", "property:read", "facility:read", "tenant:read", "lease:read", "billing:read", "deposit:read", "maintenance:read", "settlement:read", "audit:read", "report:read", "notification:read"}
	return []roleSeed{{"business_admin", "Full operational administrator", all}, {"daily_operator", "Daily operations staff", operator}, {"approval_reviewer", "Approval and review staff", reviewer}, {"readonly_auditor", "Read-only audit staff", auditor}}
}
