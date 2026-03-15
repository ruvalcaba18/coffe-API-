package cache

import (
	"context"
	"time"
)

type Service interface {
	Get(requestContext context.Context, key string) (string, error)
	Set(requestContext context.Context, key string, value interface{}, expiration time.Duration) error
	Del(requestContext context.Context, keys ...string) error
	SetNX(requestContext context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	Incr(requestContext context.Context, key string) (int64, error)
	Expire(requestContext context.Context, key string, expiration time.Duration) error
	FlushDB(requestContext context.Context) error
	Close() error
}
