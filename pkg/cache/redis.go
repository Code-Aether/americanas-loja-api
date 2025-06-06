package cache

import (
	"context"
	"log"

	"github.com/go-redis/redis/v8"
)

func NewRedisClient(redisURL, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("Redis connection failed: %v", err)
	}
	return rdb
}
