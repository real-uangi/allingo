/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:16
 */

// Package kv

package kv

import (
	"context"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"github.com/redis/go-redis/v9"
	"runtime"
	"time"
)

type RedisKV struct {
	client *redis.Client
	option *redis.Options
	logger *log.StdLogger
}

func newRedisKV(addr, password string) *RedisKV {
	options := &redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           env.GetIntOrDefault("REDIS_DB", 0),
		PoolSize:     env.GetIntOrDefault("REDIS_POOL_SIZE", 10*runtime.GOMAXPROCS(0)),
		MinIdleConns: env.GetIntOrDefault("REDIS_MAX_IDLE", 0),
	}
	return &RedisKV{
		option: options,
		client: redis.NewClient(options),
	}
}

func (kv *RedisKV) Set(key string, value string, ttl time.Duration) {
	err := filterErr(kv.client.Set(context.Background(), key, value, ttl).Err())
	if err != nil {

	}
}

func (kv *RedisKV) Get(key string) (string, bool) {
	v, err := kv.client.Get(context.Background(), key).Result()
	if err := filterErr(err); err != nil {
		return "", false
	}
	return v, true
}

func (kv *RedisKV) Del(key string) error {
	return filterErr(kv.client.Del(context.Background(), key).Err())
}

func filterErr(err error) error {
	if err == redis.Nil {
		return nil
	}
	return err
}
