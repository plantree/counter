package main

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

type DB struct {
	redisClient *redis.Client
	logger      *logrus.Logger
}

type RedisResult struct {
	key   string
	value interface{}
}

func NewRedisClient(redisUrl string, logger *logrus.Logger) *DB {
	db := &DB{}
	opt, err := redis.ParseURL(redisUrl)
	if err != nil {
		logger.Info("redis url parse error: ", err)
		return nil
	}
	db.redisClient = redis.NewClient(opt)
	db.logger = logger
	return db
}

func (db *DB) RefreshExpire(key string) error {
	return db.redisClient.Expire(ctx, key, REDIS_KEY_TTL).Err()
}

func (db *DB) Set(key string, value interface{}, use_ttl bool) error {
	// refresh expire automatically
	var err error
	if use_ttl {
		_, err = db.redisClient.Set(ctx, key, value, REDIS_KEY_TTL).Result()
	} else {
		// no ttl
		_, err = db.redisClient.Set(ctx, key, value, 0).Result()
	}
	if err != nil {
		errMsg := fmt.Errorf("Set key[%s] with value[%s] failed. err[%v]", key, value, err)
		return errMsg
	}
	return nil
}

func (db *DB) Get(key string) (*RedisResult, error) {
	val, err := db.redisClient.Get(ctx, key).Result()
	switch {
	case err == redis.Nil:
		errMsg := fmt.Errorf("key[%s] does not exist", key)
		return nil, errMsg
	case err != nil:
		errMsg := fmt.Errorf("get key[%s] failed. err[%v]", key, err)
		return nil, errMsg
	case val == "":
		errMsg := fmt.Errorf("key[%s] is empty", key)
		return nil, errMsg
	default:
		// refresh expire
		if err = db.RefreshExpire(key); err != nil {
			errMsg := fmt.Errorf("expire key[%s] failed. err[%v]", key, err)
			return nil, errMsg
		}
		return &RedisResult{key: key, value: val}, nil
	}
}

func (db *DB) BatchGet(keys ...string) ([]RedisResult, error) {
	// using pipeline
	results := make([]RedisResult, 0)
	values := make([]*redis.StringCmd, 0)
	pipe := db.redisClient.Pipeline()
	for _, key := range keys {
		val := pipe.Get(ctx, key)
		pipe.Expire(ctx, key, REDIS_KEY_TTL)
		values = append(values, val)
	}
	_, err := pipe.Exec(ctx)
	if err != nil {
		errMsg := fmt.Errorf("exec pipeline failed. err[%v]", err)
		return nil, errMsg
	}
	for index, val := range values {
		results = append(results, RedisResult{key: keys[index], value: val.Val()})
	}
	return results, nil
}

func (db *DB) Incr(key string) (*RedisResult, error) {
	pipe := db.redisClient.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, REDIS_KEY_TTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		errMsg := fmt.Errorf("incr key[%s] failed. err[%v]", key, err)
		return nil, errMsg
	}
	return &RedisResult{key: key, value: incr.Val()}, nil
}

func (db *DB) Delete(keys ...string) (int64, error) {
	return db.redisClient.Del(ctx, keys...).Result()
}

func (db *DB) GetPrefixMatchKeys(pattern string) ([]string, error) {
	// using `scan` instead of `keys`
	allKeys := make([]string, 0)
	var cursor uint64

	for {
		var keys []string
		var err error
		keys, cursor, err = db.redisClient.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			errMsg := fmt.Errorf("scan pattern[%s] failed. err:[%v]", pattern, err)
			return nil, errMsg
		}
		allKeys = append(allKeys, keys...)
		if cursor == 0 {
			break
		}
	}
	return allKeys, nil
}
