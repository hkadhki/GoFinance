package cache

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

var Default Cache

func InitCache(ctx context.Context) {
	client := redis.NewClient(&redis.Options{
		Addr:     getenv("REDIS_ADDR", "localhost:6379"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       getenvInt("REDIS_DB", 0),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		log.Println("redis disabled:", err)
		Default = NewNoCache()
		return
	}

	log.Println("Redis connected")
	Default = NewCache(client)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}
