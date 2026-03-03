package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserPermissionCache struct {
	redis  *redis.Client
	prefix string
}

func NewUserPermissionCache(client *redis.Client) *UserPermissionCache {
	return &UserPermissionCache{
		redis:  client,
		prefix: "user:perms:",
	}
}

func (c *UserPermissionCache) Set(ctx context.Context, userID uuid.UUID, perms []string, ttl time.Duration) error {

	if userID == uuid.Nil {
		return errors.New("user id is empty")
	}
	if ttl <= 0 {
		ttl = 60 * time.Minute
	}
	payload, err := json.Marshal(perms)
	if err != nil {
		return err
	}
	return c.redis.Set(ctx, c.prefix+userID.String(), payload, ttl).Err()
}

func (c *UserPermissionCache) Get(ctx context.Context, userID string) ([]string, error) {
	if c == nil || c.redis == nil {
		return nil, errors.New("redis client is nil")
	}
	if userID == "" {
		return nil, errors.New("user id is empty")
	}
	raw, err := c.redis.Get(ctx, c.prefix+userID).Result()
	if err != nil {
		return nil, err
	}
	var perms []string
	if err := json.Unmarshal([]byte(raw), &perms); err != nil {
		return nil, err
	}
	return perms, nil
}

func (c *UserPermissionCache) DeleteByUser(ctx context.Context, userID string) error {
	if c == nil || c.redis == nil {
		return errors.New("redis client is nil")
	}
	if userID == "" {
		return errors.New("user id is empty")
	}

	pattern := c.prefix + userID + "*"
	var cursor uint64
	for {
		keys, next, err := c.redis.Scan(ctx, cursor, pattern, 200).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.redis.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			break
		}
	}
	return nil
}
