package repoImple

import (
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/errorx"
	"aurora/internal/model"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RBACRepoImple struct {
	db *pgxpool.Pool
}

func NewRBACRepoImple(db *pgxpool.Pool) domainrepo.RBACRepoInterface {
	return &RBACRepoImple{db: db}
}

// Roles
func (r *RBACRepoImple) CreateRole(ctx context.Context, role *entity.Role) error {
	if role == nil {
		return errorx.ErrEntityNil
	}
	m := model.RoleEntityToModel(*role)
	const q = `
		INSERT INTO roles (id, name, description, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT DO NOTHING
	`
	cmd, err := r.db.Exec(ctx, q, m.ID, m.Name, m.Description, m.CreatedAt)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrRoleAlreadyExist
	}
	return nil
}

func (r *RBACRepoImple) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*entity.Role, error) {
	const q = `
		SELECT id, name, description, created_at
		FROM roles
		WHERE id = $1
		LIMIT 1
	`
	var m model.Role
	if err := r.db.QueryRow(ctx, q, roleID).Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrRoleNotFound
		}
		return nil, err
	}
	role := model.RoleModelToEntity(m)
	return &role, nil
}

func (r *RBACRepoImple) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	const q = `
		SELECT id, name, description, created_at
		FROM roles
		WHERE name = $1
		LIMIT 1
	`
	var m model.Role
	if err := r.db.QueryRow(ctx, q, name).Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrRoleNotFound
		}
		return nil, err
	}
	role := model.RoleModelToEntity(m)
	return &role, nil
}

func (r *RBACRepoImple) ListRoles(ctx context.Context) ([]entity.Role, error) {
	const q = `
		SELECT id, name, description, created_at
		FROM roles
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Role
	for rows.Next() {
		var m model.Role
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, model.RoleModelToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *RBACRepoImple) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	const q = `DELETE FROM roles WHERE id = $1`
	cmd, err := r.db.Exec(ctx, q, roleID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrRoleNotFound
	}
	return nil
}

// Permissions (read-only)
func (r *RBACRepoImple) GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*entity.Permission, error) {
	const q = `
		SELECT id, name, description, created_at
		FROM permissions
		WHERE id = $1
		LIMIT 1
	`
	var m model.Permission
	if err := r.db.QueryRow(ctx, q, permissionID).Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrPermissionNotFound
		}
		return nil, err
	}
	perm := model.PermissionModelToEntity(m)
	return &perm, nil
}

func (r *RBACRepoImple) GetPermissionByName(ctx context.Context, name string) (*entity.Permission, error) {
	const q = `
		SELECT id, name, description, created_at
		FROM permissions
		WHERE name = $1
		LIMIT 1
	`
	var m model.Permission
	if err := r.db.QueryRow(ctx, q, name).Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorx.ErrPermissionNotFound
		}
		return nil, err
	}
	perm := model.PermissionModelToEntity(m)
	return &perm, nil
}

func (r *RBACRepoImple) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	const q = `
		SELECT id, name, description, created_at
		FROM permissions
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Permission
	for rows.Next() {
		var m model.Permission
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, model.PermissionModelToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// Role-Permission mapping
func (r *RBACRepoImple) AddPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	const q = `
		INSERT INTO role_permissions (id, role_id, permission_id, created_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`
	now := time.Now()
	_, err := r.db.Exec(ctx, q, uuid.New(), roleID, permissionID, &now)
	return err
}

func (r *RBACRepoImple) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	const q = `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`
	cmd, err := r.db.Exec(ctx, q, roleID, permissionID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrRolePermissionNotFound
	}
	return nil
}

func (r *RBACRepoImple) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error) {
	const q = `
		SELECT p.id, p.name, p.description, p.created_at
		FROM role_permissions rp
		JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role_id = $1
	`
	rows, err := r.db.Query(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Permission
	for rows.Next() {
		var m model.Permission
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, model.PermissionModelToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// User-Role mapping
func (r *RBACRepoImple) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	const existsQ = `
		SELECT 1 FROM user_roles
		WHERE user_id = $1 AND role_id = $2
		LIMIT 1
	`
	var exists int
	if err := r.db.QueryRow(ctx, existsQ, userID, roleID).Scan(&exists); err == nil {
		return nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}

	const q = `
		INSERT INTO user_roles (id, user_id, role_id, created_at)
		VALUES ($1,$2,$3,$4)
	`
	now := time.Now()
	_, err := r.db.Exec(ctx, q, uuid.New(), userID, roleID, &now)
	return err
}

func (r *RBACRepoImple) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	const q = `
		DELETE FROM user_roles
		WHERE user_id = $1 AND role_id = $2
	`
	cmd, err := r.db.Exec(ctx, q, userID, roleID)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errorx.ErrUserRoleNotFound
	}
	return nil
}

func (r *RBACRepoImple) ListUserIDsByRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	const q = `
		SELECT ur.user_id
		FROM user_roles ur
		WHERE ur.role_id = $1
	`
	rows, err := r.db.Query(ctx, q, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]uuid.UUID, 0)
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		out = append(out, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *RBACRepoImple) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error) {
	const q = `
		SELECT r.id, r.name, r.description, r.created_at
		FROM user_roles ur
		JOIN roles r ON r.id = ur.role_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Role
	for rows.Next() {
		var m model.Role
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, model.RoleModelToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *RBACRepoImple) ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]entity.Permission, error) {
	const q = `
		SELECT DISTINCT p.id, p.name, p.description, p.created_at
		FROM user_roles ur
		JOIN role_permissions rp ON rp.role_id = ur.role_id
		JOIN permissions p ON p.id = rp.permission_id
		WHERE ur.user_id = $1
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []entity.Permission
	for rows.Next() {
		var m model.Permission
		if err := rows.Scan(&m.ID, &m.Name, &m.Description, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, model.PermissionModelToEntity(m))
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}
