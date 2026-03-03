package app

import (
	"aurora/internal/transport/http/handler"
	"aurora/internal/transport/http/middleware"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires HTTP routes to handlers.
func RegisterRoutes(r *gin.Engine, m *Modules,
	health *handler.HealthHandler) {

	r.GET("/health/liveness", health.Liveness)
	r.GET("/health/readiness", health.Readiness)
	r.GET("/health/startup", health.Startup)

	auth := r.Group("/auth")
	auth.POST(
		"/register",
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "auth_register", 5, 5, time.Minute),
		m.AuthHandler.SignupAccount,
	)
	auth.POST(
		"/login",
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "auth_login", 10, 10, time.Minute),
		m.AuthHandler.Login,
	)
	auth.POST(
		"/active",
		middleware.RateLimit(m.RateLimiter, "auth_active", 5, 5, time.Minute),
		m.UserHandler.ActiveAccount,
	)
	auth.POST(
		"/forgot-password",
		middleware.RateLimit(m.RateLimiter, "forgot-password", 10, 10, time.Minute),
		m.UserHandler.ForgotPasswd,
	)
	auth.POST(
		"/reset-password/verify",
		middleware.RateLimit(m.RateLimiter, "reset-password-verify", 10, 10, time.Minute),
		m.UserHandler.VerifyResetPassword,
	)
	auth.POST(
		"/new-password",
		middleware.RateLimit(m.RateLimiter, "new-password", 10, 10, time.Minute),
		m.UserHandler.NewPassword,
	)
	auth.POST(
		"/refresh",
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "auth_refresh", 30, 30, time.Minute),
		m.AuthHandler.Refresh,
	)
	auth.POST(
		"/logout",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "auth_logout", 30, 30, time.Minute),
		m.AuthHandler.Logout,
	)
	auth.POST(
		"/mfa/totp/setup",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "mfa_totp_setup", 10, 10, time.Minute),
		m.MfaHandler.BeginTOTPSetup,
	)
	auth.POST(
		"/mfa/totp/verify",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "mfa_totp_verify", 10, 10, time.Minute),
		m.MfaHandler.VerifyTOTPSetup,
	)
	auth.POST(
		"/mfa/challenge/verify",
		middleware.RateLimit(m.RateLimiter, "mfa_challenge_verify", 15, 15, time.Minute),
		m.AuthHandler.VerifyMFAChallenge,
	)
	auth.POST(
		"/mfa/2fa/disable",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "mfa_totp_disable", 10, 10, time.Minute),
		m.MfaHandler.Disable2FA,
	)
	auth.POST(
		"/mfa/recovery-codes",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "mfa_recovery_codes", 5, 5, time.Minute),
		m.MfaHandler.GenerateRecoveryCodes,
	)
	auth.GET(
		"/mfa/methods",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "mfa_methods_list", 20, 20, time.Minute),
		m.MfaHandler.ListMethods,
	)
	auth.GET(
		"/me",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "auth_me", 30, 30, time.Minute),
		m.UserHandler.Me,
	)
	auth.POST(
		"/profile",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
		middleware.RateLimit(m.RateLimiter, "auth_profile_upsert", 10, 10, time.Minute),
		m.UserHandler.UpsertProfile,
	)

	rbac := r.Group(
		"/rbac",
		middleware.AuthJWT(m.Token, m.JwtBlacklist),
		middleware.DeviceContext(m.DeviceCache, m.DeviceRepo, m.Token, m.Token.RefreshTTL),
	)
	rbac.GET(
		"/roles",
		middleware.RequirePermission("role.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_roles_list", 30, 30, time.Minute),
		m.RbacHandler.ListRoles,
	)
	rbac.POST(
		"/roles",
		middleware.RequirePermission("role.write", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_roles_create", 10, 10, time.Minute),
		m.RbacHandler.CreateRole,
	)
	rbac.GET(
		"/roles/:id",
		middleware.RequirePermission("role.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_roles_get", 30, 30, time.Minute),
		m.RbacHandler.GetRole,
	)
	rbac.DELETE(
		"/roles/:id",
		middleware.RequirePermission("role.write", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_roles_delete", 10, 10, time.Minute),
		m.RbacHandler.DeleteRole,
	)
	rbac.GET(
		"/permissions",
		middleware.RequirePermission("permission.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_permissions_list", 30, 30, time.Minute),
		m.RbacHandler.ListPermissions,
	)
	rbac.GET(
		"/permissions/:id",
		middleware.RequirePermission("permission.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_permissions_get", 30, 30, time.Minute),
		m.RbacHandler.GetPermission,
	)
	rbac.GET(
		"/roles/:id/permissions",
		middleware.RequirePermission("permission.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_role_permissions_list", 30, 30, time.Minute),
		m.RbacHandler.ListRolePermissions,
	)
	rbac.POST(
		"/roles/:id/permissions/:permission_id",
		middleware.RequirePermission("permission.write", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_role_permissions_add", 10, 10, time.Minute),
		m.RbacHandler.AddPermissionToRole,
	)
	rbac.DELETE(
		"/roles/:id/permissions/:permission_id",
		middleware.RequirePermission("permission.write", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_role_permissions_remove", 10, 10, time.Minute),
		m.RbacHandler.RemovePermissionFromRole,
	)
	rbac.GET(
		"/users/:id/roles",
		middleware.RequirePermission("role.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_user_roles_list", 30, 30, time.Minute),
		m.RbacHandler.ListUserRoles,
	)
	rbac.POST(
		"/users/:id/roles/:role_id",
		middleware.RequirePermission("role.write", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_user_roles_add", 10, 10, time.Minute),
		m.RbacHandler.AssignRoleToUser,
	)
	rbac.DELETE(
		"/users/:id/roles/:role_id",
		middleware.RequirePermission("role.write", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_user_roles_remove", 10, 10, time.Minute),
		m.RbacHandler.RemoveRoleFromUser,
	)
	rbac.GET(
		"/users/:id/permissions",
		middleware.RequirePermission("permission.read", m.PermCache, m.RbacRepo),
		middleware.RateLimit(m.RateLimiter, "rbac_user_permissions_list", 30, 30, time.Minute),
		m.RbacHandler.ListUserPermissions,
	)

}
