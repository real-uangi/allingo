/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:12
 */

// Package kv

package kv

import (
	"github.com/real-uangi/allingo/common/env"
	"time"
)

type KV interface {
	Set(key string, value string, ttl time.Duration)
	Get(key string) (string, bool)
	Del(key string) error
}

func InitKV() KV {
	redisAddr := env.Get("REDIS_ADDR")
	redisPassword := env.Get("REDIS_PASSWORD")
	if redisAddr != "" || redisPassword != "" {
		return newRedisKV(redisAddr, redisPassword)
	}
	return newLocalKV()
}
