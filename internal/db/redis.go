package db

import (
	"context"
	"foxy/internal/env"
	"github.com/redis/go-redis/v9"
	"time"
)

func GetRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     env.FoxyEnvironment.RedisUrl,
		Password: "",
		DB:       0, // use default DB
	})
}

func RedisGet(key string) (*string, error) {
	client := GetRedisClient()
	//noinspection GoUnhandledErrorResult
	defer client.Close()

	val, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &val, nil
}

func RedisSet(key string, val interface{}, expiration time.Duration) error {
	client := GetRedisClient()
	//noinspection GoUnhandledErrorResult
	defer client.Close()

	return client.Set(context.Background(), key, val, expiration).Err()
}

func RedisDelete(key string) error {
	client := GetRedisClient()
	//noinspection GoUnhandledErrorResult
	defer client.Close()

	return client.Del(context.Background(), key).Err()
}
