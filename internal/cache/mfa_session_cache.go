package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type MFASessionCache struct {
	redis  *redis.Client
	prefix string
}

func NewMFASessionCache(client *redis.Client) *MFASessionCache {
	return &MFASessionCache{
		redis:  client,
		prefix: "mfasession:",
	}
}

func (c *MFASessionCache) Set(ctx context.Context, userID string, token string, ttl time.Duration) error {
	if c == nil || c.redis == nil {
		return errors.New("redis client is nil")
	}
	if userID == "" || token == "" {
		return errors.New("user id or token is empty")
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return c.redis.Set(ctx, c.prefix+userID, token, ttl).Err()
}

func (c *MFASessionCache) Get(ctx context.Context, userID string) (string, error) {
	if c == nil || c.redis == nil {
		return "", errors.New("redis client is nil")
	}
	if userID == "" {
		return "", errors.New("user id is empty")
	}
	return c.redis.Get(ctx, c.prefix+userID).Result()
}

func (c *MFASessionCache) Delete(ctx context.Context, userID string) error {
	if c == nil || c.redis == nil {
		return errors.New("redis client is nil")
	}
	if userID == "" {
		return errors.New("user id is empty")
	}
	return c.redis.Del(ctx, c.prefix+userID).Err()
}
