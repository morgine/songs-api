package cache

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type Engine interface {
	Set(key string, value []byte, expiration time.Duration) error
	Get(key string) (value []byte, err error)
	Del(key string) error
}

type RedisEngine struct {
	client *redis.Client
}

func (r *RedisEngine) Del(key string) error {
	return r.client.Del(noCtx, key).Err()
}

func NewRedisEngine(client *redis.Client) Engine {
	return &RedisEngine{client: client}
}

var noCtx = context.Background()

func (r *RedisEngine) Set(key string, value []byte, expiration time.Duration) error {
	return r.client.Set(noCtx, key, string(value), expiration).Err()
}

func (r *RedisEngine) Get(key string) (value []byte, err error) {
	value, err = r.client.Get(noCtx, key).Bytes()
	if err != nil && err != redis.Nil {
		return nil, err
	} else {
		return value, nil
	}
}
