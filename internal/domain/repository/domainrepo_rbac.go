package domainrepo

import (
	"aurora/internal/domain/entity"
	"context"

	"github.com/google/uuid"
)

// RBACRepoInterface defines persistence operations for roles, permissions, and assignments.
type RBACRepoInterface interface {
	// Roles
	CreateRole(ctx context.Context, role *entity.Role) error
	GetRoleByID(ctx context.Context, roleID uuid.UUID) (*entity.Role, error)
	GetRoleByName(ctx context.Context, name string) (*entity.Role, error)
	ListRoles(ctx context.Context) ([]entity.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error

	// Permissions
	GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*entity.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*entity.Permission, error)
	ListPermissions(ctx context.Context) ([]entity.Permission, error)

	// Role-Permission mapping
	AddPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error
	ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error)

	// User-Role mapping
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	ListUserIDsByRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
	ListUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error)
	ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]entity.Permission, error)
}
