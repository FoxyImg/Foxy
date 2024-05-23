package db

import (
	"context"
	"github.com/redis/go-redis/v9"
	"os"
	"time"
)

func GetRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0, // use default DB
	})
}

func RedisGet(key string) (*string, error) {
	client := GetRedisClient()
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
	defer client.Close()

	return client.Set(context.Background(), key, val, 0).Err()
}
