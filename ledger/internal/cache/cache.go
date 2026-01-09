package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration)
	Delete(ctx context.Context, keys ...string)
	DeleteByPattern(ctx context.Context, pattern string)
}
