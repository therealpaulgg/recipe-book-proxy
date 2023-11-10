package cache

import (
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/samber/do"
)

func NewRedisClient(i *do.Injector) (*redis.Client, error) {
	return redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_CLIENT"),
		Password: "",
		DB: 0,
	}), nil
}