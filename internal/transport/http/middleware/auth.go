package middleware

import (
	"aurora/internal/cache"
	"aurora/internal/config"
	appctxKey "aurora/internal/domain/key"
	"aurora/internal/security"
	"aurora/internal/transport/http/response"
	"aurora/pkg/logger"
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

const claimsContextKey = "jwt_claims"

func AuthJWT(tokenCfg *config.TokenCfg, blacklist *cache.JWTBlacklist) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		token := extractAccessToken(c)
		if token == "" {
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, "auth.jwt", "missing access token")
			c.Abort()
			return
		}

		claims, err := security.DecodeJWT(token, tokenCfg.GetAccessSecret())
		if err != nil {
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, "auth.jwt", "invalid access token")
			c.Abort()
			return
		}

		if blacklist != nil {
			if jti, ok := claims["jti"].(string); ok && strings.TrimSpace(jti) != "" {
				blocked, err := blacklist.IsBlocked(ctx, strings.TrimSpace(jti))
				if err != nil {
					response.RespondUnauthorized(c, "unauthorized")
					logger.HandlerWarn(ctx, "auth.jwt", "blacklist check failed: %s", err.Error())
					c.Abort()
					return
				}
				if blocked {
					response.RespondUnauthorized(c, "unauthorized")
					logger.HandlerInfo(ctx, "auth.jwt", "token blacklisted")
					c.Abort()
					return
				}
			}
		}

		if sub, ok := claims["sub"].(string); ok && strings.TrimSpace(sub) != "" {
			ctx = context.WithValue(ctx, appctxKey.KeyUserID, sub)
			c.Set("user_id", sub)
		}

		if deviceID, ok := claims["device_id"].(string); ok && strings.TrimSpace(deviceID) != "" {
			deviceID = strings.TrimSpace(deviceID)
			ctx = context.WithValue(ctx, appctxKey.KeyJWTDeviceID, deviceID)
			c.Set("jwt_device_id", deviceID)
		}

		if workspaceID, ok := claims["workspace_id"].(string); ok && strings.TrimSpace(workspaceID) != "" {
			workspaceID = strings.TrimSpace(workspaceID)
			ctx = context.WithValue(ctx, appctxKey.KeyWorkspaceID, workspaceID)
			c.Set("workspace_id", workspaceID)
		}

		if tenantID, ok := claims["tenant_id"].(string); ok && strings.TrimSpace(tenantID) != "" {
			tenantID = strings.TrimSpace(tenantID)
			ctx = context.WithValue(ctx, appctxKey.KeyTenantID, tenantID)
			c.Set("tenant_id", tenantID)
		}

		if jti, ok := claims["jti"].(string); ok && strings.TrimSpace(jti) != "" {
			jti = strings.TrimSpace(jti)
			ctx = context.WithValue(ctx, appctxKey.KeyJWTID, jti)
			c.Set("jti", jti)
		}

		if exp, ok := getNumericClaim(claims, "exp"); ok && exp > 0 {
			ctx = context.WithValue(ctx, appctxKey.KeyJWTExp, exp)
			c.Set("jwt_exp", exp)
		}

		userLevel, ok := getNumericClaim(claims, "user_level")
		if !ok {
			response.RespondUnauthorized(c, "unauthorized")
			logger.HandlerInfo(ctx, "auth.jwt", "missing user level in access token")
			c.Abort()
			return
		}
		ctx = context.WithValue(ctx, appctxKey.KeyUserLevel, int32(userLevel))
		c.Set("user_level", int32(userLevel))

		c.Set(claimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func extractAccessToken(c *gin.Context) string {
	if token, err := c.Cookie("access_token"); err == nil && strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token)
	}
	authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
	if authHeader == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		return strings.TrimSpace(authHeader[7:])
	}
	return ""
}

func getNumericClaim(claims map[string]any, key string) (int64, bool) {
	val, ok := claims[key]
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return int64(v), true
	case int64:
		return v, true
	case int:
		return int64(v), true
	default:
		return 0, false
	}
}
