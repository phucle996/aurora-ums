package model

import (
	"aurora/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

// Role mirrors roles table.
type Role struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
}

func RoleModelToEntity(m Role) entity.Role {
	return entity.Role{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

func RoleEntityToModel(e entity.Role) Role {
	return Role{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
	}
}

// Permission mirrors permissions table.
type Permission struct {
	ID          uuid.UUID  `db:"id"`
	Name        string     `db:"name"`
	Description *string    `db:"description"`
	CreatedAt   *time.Time `db:"created_at"`
}

func PermissionModelToEntity(m Permission) entity.Permission {
	return entity.Permission{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		CreatedAt:   m.CreatedAt,
	}
}

func PermissionEntityToModel(e entity.Permission) Permission {
	return Permission{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
		CreatedAt:   e.CreatedAt,
	}
}

// RolePermission mirrors role_permissions table.
type RolePermission struct {
	ID           uuid.UUID  `db:"id"`
	RoleID       uuid.UUID  `db:"role_id"`
	PermissionID uuid.UUID  `db:"permission_id"`
	CreatedAt    *time.Time `db:"created_at"`
}

func RolePermissionModelToEntity(m RolePermission) entity.RolePermission {
	return entity.RolePermission{
		ID:           m.ID,
		RoleID:       m.RoleID,
		PermissionID: m.PermissionID,
		CreatedAt:    m.CreatedAt,
	}
}

func RolePermissionEntityToModel(e entity.RolePermission) RolePermission {
	return RolePermission{
		ID:           e.ID,
		RoleID:       e.RoleID,
		PermissionID: e.PermissionID,
		CreatedAt:    e.CreatedAt,
	}
}

// UserRole mirrors user_roles table.
type UserRole struct {
	ID        uuid.UUID  `db:"id"`
	UserID    uuid.UUID  `db:"user_id"`
	RoleID    uuid.UUID  `db:"role_id"`
	CreatedAt *time.Time `db:"created_at"`
}

func UserRoleModelToEntity(m UserRole) entity.UserRole {
	return entity.UserRole{
		ID:        m.ID,
		UserID:    m.UserID,
		RoleID:    m.RoleID,
		CreatedAt: m.CreatedAt,
	}
}

func UserRoleEntityToModel(e entity.UserRole) UserRole {
	return UserRole{
		ID:        e.ID,
		UserID:    e.UserID,
		RoleID:    e.RoleID,
		CreatedAt: e.CreatedAt,
	}
}
