package redisinfra

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type EventPublisher interface {
	Publish(ctx context.Context, stream string, values map[string]any) error
}

type RedisStreamPublisher struct {
	rdb *redis.Client
}

func NewRedisStreamPublisher(rdb *redis.Client) *RedisStreamPublisher {
	return &RedisStreamPublisher{rdb: rdb}
}

func (p *RedisStreamPublisher) Publish(
	ctx context.Context,
	stream string,
	values map[string]any,
) error {
	return p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Err()
}
