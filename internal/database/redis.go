package database

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) {
	redisAddress := os.Getenv("REDIS_ADDR")
	password := os.Getenv("REDIS_PASSWORD")

	if redisAddress == "" {
		redisAddress = "localhost:6379"
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: password, 
		DB:       0,        
	})

	backgroundContext := context.Background()
	_, error := redisClient.Ping(backgroundContext).Result()
	if error != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", error)
	}

	return redisClient, nil
}
