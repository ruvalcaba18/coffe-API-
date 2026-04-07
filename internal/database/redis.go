package database

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() (*redis.Client, error) {
	var redisOptions *redis.Options

	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		parsedOptions, err := redis.ParseURL(redisURL)
		if err != nil {
			return nil, fmt.Errorf("failed to parse REDIS_URL: %v", err)
		}
		redisOptions = parsedOptions
	} else {
		redisAddress := os.Getenv("REDIS_ADDR")
		password := os.Getenv("REDIS_PASSWORD")

		if redisAddress == "" {
			redisAddress = "localhost:6379"
		}

		redisOptions = &redis.Options{
			Addr:     redisAddress,
			Password: password,
			DB:       0,
		}
	}

	redisClient := redis.NewClient(redisOptions)

	backgroundContext := context.Background()
	_, error := redisClient.Ping(backgroundContext).Result()
	if error != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", error)
	}

	return redisClient, nil
}
