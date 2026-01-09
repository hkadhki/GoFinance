package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	client *redis.Client
}

func NewCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}
	return data, true
}

func (r *RedisCache) Set(
	ctx context.Context,
	key string,
	value []byte,
	ttl time.Duration,
) {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		log.Println("cache set error:", err)
	}
}

func (r *RedisCache) Delete(
	ctx context.Context,
	keys ...string,
) {
	if len(keys) == 0 {
		return
	}
	_ = r.client.Unlink(ctx, keys...).Err()
}

func (r *RedisCache) DeleteByPattern(
	ctx context.Context,
	pattern string,
) {
	var cursor uint64

	for {
		keys, nextCursor, err := r.client.Scan(
			ctx,
			cursor,
			pattern,
			100,
		).Result()
		if err != nil {
			log.Println("cache scan error:", err)
			return
		}

		if len(keys) > 0 {
			_ = r.client.Unlink(ctx, keys...).Err()
		}

		if nextCursor == 0 {
			break
		}

		cursor = nextCursor
	}
}
