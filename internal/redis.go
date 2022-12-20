package internal

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

func NewRedisClient(redisUrl string, logger *logrus.Logger) *redis.Client {
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		logger.Info("redis url parse error: ", err)
		return nil
	}
	rdb := redis.NewClient(opt)
	return rdb
}
