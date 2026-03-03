package entity

import (
	"time"

	"github.com/google/uuid"
)

// Role defines a collection of permissions.
type Role struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   *time.Time
}

// Permission represents an action that can be assigned to roles.
type Permission struct {
	ID          uuid.UUID
	Name        string
	Description *string
	CreatedAt   *time.Time
}

// RolePermission links roles to permissions.
type RolePermission struct {
	ID           uuid.UUID
	RoleID       uuid.UUID
	PermissionID uuid.UUID
	CreatedAt    *time.Time
}

// UserRole links users to roles.
type UserRole struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	RoleID    uuid.UUID
	CreatedAt *time.Time
}
