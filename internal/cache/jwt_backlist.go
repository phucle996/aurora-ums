package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type JWTBlacklist struct {
	redis  *redis.Client
	prefix string
}

func NewJWTBlacklist(client *redis.Client) *JWTBlacklist {
	return &JWTBlacklist{
		redis:  client,
		prefix: "jwt:blacklist:",
	}
}

func (b *JWTBlacklist) Block(ctx context.Context, jti string, ttl time.Duration) error {
	if b == nil || b.redis == nil {
		return errors.New("redis client is nil")
	}
	if jti == "" {
		return errors.New("jti is empty")
	}
	if ttl <= 0 {
		ttl = time.Second
	}
	return b.redis.Set(ctx, b.prefix+jti, "1", ttl).Err()
}

func (b *JWTBlacklist) IsBlocked(ctx context.Context, jti string) (bool, error) {
	if b == nil || b.redis == nil {
		return false, errors.New("redis client is nil")
	}
	if jti == "" {
		return false, nil
	}
	_, err := b.redis.Get(ctx, b.prefix+jti).Result()
	if err == nil {
		return true, nil
	}
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	return false, err
}
