package middleware

import (
	"aurora/internal/cache"
	appctxKey "aurora/internal/domain/key"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/transport/http/response"
	"aurora/pkg/logger"
	"context"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func RequirePermission(required string, permCache *cache.UserPermissionCache, repo domainrepo.RBACRepoInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		claimsAny, ok := c.Get(claimsContextKey)
		if !ok {
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, "auth.perm", "missing jwt claims")
			c.Abort()
			return
		}

		claims, ok := claimsAny.(map[string]any)
		if !ok {
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, "auth.perm", "invalid jwt claims")
			c.Abort()
			return
		}

		roleNames := extractRolesFromClaims(claims)
		if len(roleNames) > 0 {
			c.Set("roles", roleNames)
		}

		userID := c.GetString("user_id")
		if userID == "" {
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, "auth.perm", "missing user id")
			c.Abort()
			return
		}

		perms, err := fetchPermissions(ctx, permCache, repo, userID)
		if err != nil {
			response.RespondServiceUnavailable(c, "permission cache unavailable")
			logger.HandlerWarn(ctx, "auth.perm", "permission fetch failed: %s", err.Error())
			c.Abort()
			return
		}

		if !containsPermission(perms, required) {
			response.RespondForbidden(c, "forbidden")
			logger.HandlerInfo(ctx, "auth.perm", "permission denied")
			c.Abort()
			return
		}

		c.Set("permissions", perms)
		ctx = context.WithValue(ctx, appctxKey.KeyPermissions, perms)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func extractRolesFromClaims(claims map[string]any) []string {
	raw, ok := claims["roles"]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				out = append(out, strings.TrimSpace(s))
			}
		}
		return out
	default:
		return nil
	}
}

func containsPermission(perms []string, required string) bool {
	required = strings.TrimSpace(required)
	if required == "" {
		return false
	}
	for _, p := range perms {
		if strings.EqualFold(strings.TrimSpace(p), required) {
			return true
		}
	}
	return false
}

func fetchPermissions(
	ctx context.Context,
	permCache *cache.UserPermissionCache,
	repo domainrepo.RBACRepoInterface,
	userID string,
) ([]string, error) {

	perms, err := permCache.Get(ctx, userID)
	if err == nil {
		return perms, nil
	}
	if err != redis.Nil {
		return nil, err
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	roles, err := repo.ListUserRoles(ctx, parsedUserID)
	if err != nil {
		return nil, err
	}

	permSet := make(map[string]struct{})
	for _, role := range roles {
		rolePerms, err := repo.ListRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, err
		}
		for _, perm := range rolePerms {
			name := strings.TrimSpace(perm.Name)
			if name != "" {
				permSet[name] = struct{}{}
			}
		}
	}

	permNames := make([]string, 0, len(permSet))
	for name := range permSet {
		permNames = append(permNames, name)
	}
	slices.Sort(permNames)

	if permCache != nil {
		_ = permCache.Set(ctx, parsedUserID, permNames, 60*time.Minute)
	}

	return permNames, nil
}
