package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) Service {
	return &redisCache{client: client}
}

func (cacheInstance *redisCache) Get(requestContext context.Context, key string) (string, error) {
	return cacheInstance.client.Get(requestContext, key).Result()
}

func (cacheInstance *redisCache) Set(requestContext context.Context, key string, value interface{}, expiration time.Duration) error {
	return cacheInstance.client.Set(requestContext, key, value, expiration).Err()
}

func (cacheInstance *redisCache) Del(requestContext context.Context, keys ...string) error {
	return cacheInstance.client.Del(requestContext, keys...).Err()
}

func (cacheInstance *redisCache) SetNX(requestContext context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return cacheInstance.client.SetNX(requestContext, key, value, expiration).Result()
}

func (cacheInstance *redisCache) Incr(requestContext context.Context, key string) (int64, error) {
	return cacheInstance.client.Incr(requestContext, key).Result()
}

func (cacheInstance *redisCache) Expire(requestContext context.Context, key string, expiration time.Duration) error {
	return cacheInstance.client.Expire(requestContext, key, expiration).Err()
}

func (cacheInstance *redisCache) FlushDB(requestContext context.Context) error {
	return cacheInstance.client.FlushDB(requestContext).Err()
}

func (cacheInstance *redisCache) Close() error {
	return cacheInstance.client.Close()
}
