package cache

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type DeviceSecretCache struct {
	redis  *redis.Client
	prefix string
}

func NewDeviceSecretCache(client *redis.Client) *DeviceSecretCache {
	return &DeviceSecretCache{
		redis:  client,
		prefix: "device:secret:",
	}
}

func (c *DeviceSecretCache) Set(ctx context.Context, deviceID, deviceSecretHash string, ttl time.Duration) error {
	if c == nil || c.redis == nil {
		return errors.New("redis client is nil")
	}
	deviceID = strings.TrimSpace(deviceID)
	deviceSecretHash = strings.TrimSpace(deviceSecretHash)
	if deviceID == "" || deviceSecretHash == "" {
		return errors.New("device id or secret hash is empty")
	}
	if ttl <= 0 {
		return errors.New("device secret ttl must be greater than zero")
	}
	return c.redis.Set(ctx, c.prefix+deviceID, deviceSecretHash, ttl).Err()
}

func (c *DeviceSecretCache) Get(ctx context.Context, deviceID string) (string, error) {
	if c == nil || c.redis == nil {
		return "", errors.New("redis client is nil")
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return "", errors.New("device id is empty")
	}
	return c.redis.Get(ctx, c.prefix+deviceID).Result()
}
