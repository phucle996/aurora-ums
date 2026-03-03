package middleware

import (
	"aurora/internal/cache"
	"aurora/internal/config"
	appctxKey "aurora/internal/domain/key"
	domainrepo "aurora/internal/domain/repository"
	"aurora/internal/security"
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func DeviceContext(cache *cache.DeviceSecretCache, repo domainrepo.DeviceRepoInterface, tokenCfg *config.TokenCfg, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		deviceID := ""
		if raw, err := c.Cookie("device_id"); err == nil {
			deviceID = strings.TrimSpace(raw)
			if deviceID != "" {
				ctx = context.WithValue(ctx, appctxKey.KeyDeviceID, deviceID)
				c.Set("device_id", deviceID)
			}
		}

		deviceSecret := ""
		if raw, err := c.Cookie("device_secret"); err == nil {
			deviceSecret = strings.TrimSpace(raw)
			if deviceSecret != "" {
				ctx = context.WithValue(ctx, appctxKey.KeyDeviceSecret, deviceSecret)
				c.Set("device_secret", deviceSecret)
			}
		}

		fingerprint := strings.TrimSpace(c.GetHeader("X-Device-Fingerprint"))
		if fingerprint == "" {
			fingerprint = strings.TrimSpace(c.GetHeader("X-Fingerprint"))
		}

		if jwtDeviceIDRaw, ok := c.Get("jwt_device_id"); ok {
			jwtDeviceID, _ := jwtDeviceIDRaw.(string)
			jwtDeviceID = strings.TrimSpace(jwtDeviceID)
			if jwtDeviceID == "" || deviceID == "" || jwtDeviceID != deviceID {
				c.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
				return
			}

			hash := ""
			var err error
			if cache != nil {
				hash, err = cache.Get(ctx, deviceID)
			} else {
				err = redis.Nil
			}
			if err != nil {
				if err == redis.Nil {
					hash, err = fetchDeviceSecretHash(ctx, repo, deviceID)
					if err != nil {
						c.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
						return
					}
					if cache != nil {
						_ = cache.Set(ctx, deviceID, hash, ttl)
					}
				} else {
					c.AbortWithStatusJSON(500, gin.H{"message": "internal server error"})
					return
				}
			}
			if deviceSecret == "" {
				c.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
				return
			}
			secretKey := ""
			if tokenCfg != nil {
				secretKey = tokenCfg.GetDeviceSecret()
			}
			secretHash, err := security.HashToken(deviceSecret, secretKey)
			if err != nil || secretHash != hash {
				c.AbortWithStatusJSON(401, gin.H{"message": "unauthorized"})
				return
			}
		}

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func fetchDeviceSecretHash(ctx context.Context, repo domainrepo.DeviceRepoInterface, deviceID string) (string, error) {
	if repo == nil {
		return "", errors.New("device repo is nil")
	}
	parsed, err := uuid.Parse(strings.TrimSpace(deviceID))
	if err != nil {
		return "", err
	}
	device, err := repo.GetDeviceByDeviceID(ctx, parsed)
	if err != nil {
		return "", err
	}
	if device == nil || device.DeviceSecretHash == "" || device.Revoked {
		return "", errors.New("device invalid")
	}
	return device.DeviceSecretHash, nil
}
