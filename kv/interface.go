/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:12
 */

// Package kv

package kv

import (
	"errors"
	"time"

	"github.com/real-uangi/allingo/common/convert"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/ready"
	"github.com/real-uangi/fxtrategy"
)

type Type uint

const (
	Local Type = iota
	Redis
)

var ErrWrongType = errors.New("wrong type")

var errEmptyWriteValues = errors.New("at least one member or value is required")

type KeyKV interface {
	Exists(key string) (bool, error)
	Del(key string) (bool, error)
	Expire(key string, ttl time.Duration) (bool, error)
	Persist(key string) (bool, error)
	TTL(key string) (time.Duration, bool, error)
}

type StringKV interface {
	Set(key string, value string, ttl time.Duration) error
	Get(key string) (string, bool, error)
	SetStruct(key string, obj any, ttl time.Duration) error
	GetStruct(key string, p any) (bool, error)
}

type AtomicStringKV interface {
	SetIfAbsent(key string, value string, ttl time.Duration) (bool, error)
	SetIfPresent(key string, value string, ttl time.Duration) (bool, error)
	CompareAndSet(key string, expected, value string, ttl time.Duration) (bool, error)
	CompareAndDelete(key string, expected string) (bool, error)
	GetAndDelete(key string) (string, bool, error)
	GetAndSet(key string, value string, ttl time.Duration) (string, bool, error)
}

type BatchKV interface {
	MGet(keys ...string) (map[string]string, error)
	MSet(values map[string]string, ttl time.Duration) error
	MDel(keys ...string) (int64, error)
}

type CounterKV interface {
	Incr(key string, delta int64, ttl time.Duration, createTTLOnly bool) (int64, error)
}

type HashKV interface {
	HSet(key string, field string, value string, ttl time.Duration) (bool, error)
	HSetIfAbsent(key string, field string, value string, ttl time.Duration) (bool, error)
	HGet(key string, field string) (string, bool, error)
	HDel(key string, fields ...string) (int64, error)
	HExists(key string, field string) (bool, error)
	HGetAll(key string) (map[string]string, error)
	HLen(key string) (int64, error)
	HIncr(key string, field string, delta int64, ttl time.Duration) (int64, error)
}

type SetKV interface {
	SAdd(key string, ttl time.Duration, members ...string) (int64, error)
	SRem(key string, members ...string) (int64, error)
	SContains(key string, member string) (bool, error)
	SMembers(key string) ([]string, error)
	SCard(key string) (int64, error)
	SPop(key string, count int64) ([]string, error)
}

type ListKV interface {
	LPush(key string, ttl time.Duration, values ...string) (int64, error)
	RPush(key string, ttl time.Duration, values ...string) (int64, error)
	LPop(key string) (string, bool, error)
	RPop(key string) (string, bool, error)
	LLen(key string) (int64, error)
	LRange(key string, start, stop int64) ([]string, error)
	LTrim(key string, start, stop int64) error
}

type ScoredMember struct {
	Member string
	Score  float64
}

type SortedSetKV interface {
	ZAdd(key string, ttl time.Duration, members ...ScoredMember) (int64, error)
	ZRem(key string, members ...string) (int64, error)
	ZScore(key string, member string) (float64, bool, error)
	ZCard(key string) (int64, error)
	ZRange(key string, start, stop int64) ([]ScoredMember, error)
	ZRevRange(key string, start, stop int64) ([]ScoredMember, error)
	ZRangeByScore(key string, min, max float64, limit int64) ([]ScoredMember, error)
	ZPopMin(key string, count int64) ([]ScoredMember, error)
	ZPopMax(key string, count int64) ([]ScoredMember, error)
}

type LockProvider interface {
	NewLock(key string) Lock
}

type KV interface {
	KeyKV
	StringKV
	AtomicStringKV
	BatchKV
	CounterKV
	HashKV
	SetKV
	ListKV
	SortedSetKV
	LockProvider
	GetType() Type
	Ping() error
}

type Lock interface {
	TryLock(ttl time.Duration) bool
	Unlock() error
	Lock(ttl time.Duration, maxWait time.Duration) error
	Refresh(ttl time.Duration) (bool, error)
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

type checkpoint struct {
	kv KV
}

func (c *checkpoint) Ready() error {
	return c.kv.Ping()
}

func newCheckpoint(kv KV) *checkpoint {
	return &checkpoint{
		kv: kv,
	}
}

func (c *checkpoint) ItemName() string {
	return "KV-storage"
}

// CheckPoint KV缓存健康检测
var CheckPoint = fxtrategy.ProvideItem[ready.CheckPoint](newCheckpoint, ready.CPGroupName)
