package internal

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func NewRedisClient(options *redis.Options) *redis.Client {
	rdb := redis.NewClient(options)
	return rdb
}
