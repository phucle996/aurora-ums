package svcImple

import (
	"aurora/internal/cache"
	"aurora/internal/domain/entity"
	domainrepo "aurora/internal/domain/repository"
	domainsvc "aurora/internal/domain/service"
	"aurora/internal/errorx"
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

type RBACSvcImple struct {
	repo      domainrepo.RBACRepoInterface
	permCache *cache.UserPermissionCache
}

func NewRBACSvcImple(repo domainrepo.RBACRepoInterface, permCache *cache.UserPermissionCache) domainsvc.RBACSvcInterface {
	return &RBACSvcImple{repo: repo, permCache: permCache}
}

// Roles
func (s *RBACSvcImple) CreateRole(ctx context.Context, role *entity.Role) error {
	if role == nil {
		return errorx.ErrEntityNil
	}
	role.ID = uuid.New()
	now := time.Now()
	role.CreatedAt = &now
	return s.repo.CreateRole(ctx, role)
}

func (s *RBACSvcImple) GetRoleByID(ctx context.Context, roleID uuid.UUID) (*entity.Role, error) {
	return s.repo.GetRoleByID(ctx, roleID)
}

func (s *RBACSvcImple) GetRoleByName(ctx context.Context, name string) (*entity.Role, error) {
	return s.repo.GetRoleByName(ctx, strings.TrimSpace(name))
}

func (s *RBACSvcImple) ListRoles(ctx context.Context) ([]entity.Role, error) {
	return s.repo.ListRoles(ctx)
}

func (s *RBACSvcImple) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	return s.repo.DeleteRole(ctx, roleID)
}

// Permissions (read-only)
func (s *RBACSvcImple) GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*entity.Permission, error) {
	return s.repo.GetPermissionByID(ctx, permissionID)
}

func (s *RBACSvcImple) GetPermissionByName(ctx context.Context, name string) (*entity.Permission, error) {
	return s.repo.GetPermissionByName(ctx, strings.TrimSpace(name))
}

func (s *RBACSvcImple) ListPermissions(ctx context.Context) ([]entity.Permission, error) {
	return s.repo.ListPermissions(ctx)
}

// Role-Permission mapping
func (s *RBACSvcImple) AddPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if err := s.repo.AddPermissionToRole(ctx, roleID, permissionID); err != nil {
		return err
	}
	return s.invalidatePermissionCacheByRole(ctx, roleID)
}

func (s *RBACSvcImple) RemovePermissionFromRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	if err := s.repo.RemovePermissionFromRole(ctx, roleID, permissionID); err != nil {
		return err
	}
	return s.invalidatePermissionCacheByRole(ctx, roleID)
}

func (s *RBACSvcImple) ListRolePermissions(ctx context.Context, roleID uuid.UUID) ([]entity.Permission, error) {
	return s.repo.ListRolePermissions(ctx, roleID)
}

// User-Role mapping
func (s *RBACSvcImple) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.repo.AssignRoleToUser(ctx, userID, roleID); err != nil {
		return err
	}
	return s.invalidatePermissionCacheByUser(ctx, userID)
}

func (s *RBACSvcImple) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	if err := s.repo.RemoveRoleFromUser(ctx, userID, roleID); err != nil {
		return err
	}
	return s.invalidatePermissionCacheByUser(ctx, userID)
}

func (s *RBACSvcImple) ListUserRoles(ctx context.Context, userID uuid.UUID) ([]entity.Role, error) {
	return s.repo.ListUserRoles(ctx, userID)
}

func (s *RBACSvcImple) ListUserPermissions(ctx context.Context, userID uuid.UUID) ([]entity.Permission, error) {
	return s.repo.ListUserPermissions(ctx, userID)
}

func (s *RBACSvcImple) invalidatePermissionCacheByRole(ctx context.Context, roleID uuid.UUID) error {
	if s.permCache == nil {
		return nil
	}
	userIDs, err := s.repo.ListUserIDsByRole(ctx, roleID)
	if err != nil {
		return err
	}
	for _, userID := range userIDs {
		if err := s.invalidatePermissionCacheByUser(ctx, userID); err != nil {
			return err
		}
	}
	return nil
}

func (s *RBACSvcImple) invalidatePermissionCacheByUser(ctx context.Context, userID uuid.UUID) error {
	if s.permCache == nil || userID == uuid.Nil {
		return nil
	}
	return s.permCache.DeleteByUser(ctx, userID.String())
}
