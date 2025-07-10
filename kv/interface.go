/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:12
 */

// Package kv

package kv

import (
	"github.com/real-uangi/allingo/common/convert"
	"github.com/real-uangi/allingo/common/env"
	"time"
)

type Type uint

const (
	Local Type = iota
	Redis
)

type KV interface {
	Set(key string, value string, ttl time.Duration)
	Get(key string) (string, bool)
	Del(key string) error
	SetStruct(key string, obj interface{}, ttl time.Duration) error
	GetStruct(key string, p any) error
	GetType() Type
	NewLock(key string) Lock
}

type Lock interface {
	TryLock(ttl time.Duration) bool
	Unlock() error
	Lock(ttl time.Duration, maxWait time.Duration) error
}

func InitKV() KV {
	redisAddr := env.Get("REDIS_ADDR")
	redisPassword := env.Get("REDIS_PASSWORD")
	if redisAddr != "" || redisPassword != "" {
		return newRedisKV(redisAddr, redisPassword)
	}
	return newLocalKV()
}

func anyToString(obj any) (string, error) {
	return convert.Json().MarshalToString(obj)
}

func stringToAny(str string, p any) error {
	return convert.Json().UnmarshalFromString(str, p)
}
