package cache

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client

func Init(ctx context.Context) {
	addr := getenv("REDIS_ADDR", "localhost:6379")
	pass := os.Getenv("REDIS_PASSWORD")
	dbNum := getenvInt("REDIS_DB", 0)

	Client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: pass,
		DB:       dbNum,
	})

	ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := Client.Ping(ctxPing).Err(); err != nil {
		log.Printf("redis disabled: %v", err)
		Client = nil
		return
	}

	log.Println("Redis connected")
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
