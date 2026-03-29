/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:26
 */

// Package kv

package kv

import (
	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/cache"
	"github.com/real-uangi/allingo/common/log"
	"sync/atomic"
	"time"
)

type LocalKV struct {
	c      *cache.Cache[string]
	aic    *cache.Cache[*atomic.Int64]
	logger *log.StdLogger
}

func newLocalKV() *LocalKV {
	logger := log.For[LocalKV]()
	logger.Warn("using local kv, standalone mode only")
	return &LocalKV{
		c:      cache.New[string](time.Minute),
		aic:    cache.New[*atomic.Int64](time.Minute),
		logger: logger,
	}
}

func (kv *LocalKV) Set(key string, value string, ttl time.Duration) {
	kv.c.Set(key, value, ttl)
}

func (kv *LocalKV) Get(key string) (string, bool) {
	return kv.c.Get(key)
}

func (kv *LocalKV) SetIfAbsent(key string, value string, ttl time.Duration) (bool, error) {
	return kv.c.SetIfAbsent(key, value, ttl), nil
}

func (kv *LocalKV) CompareAndSet(key string, expected, value string, ttl time.Duration) (bool, error) {
	return kv.c.CompareAndSet(key, expected, value, ttl, func(left, right string) bool {
		return left == right
	}), nil
}

func (kv *LocalKV) CompareAndDelete(key string, expected string) (bool, error) {
	return kv.c.CompareAndDelete(key, expected, func(left, right string) bool {
		return left == right
	}), nil
}

func (kv *LocalKV) GetAndDelete(key string) (string, bool, error) {
	value, ok := kv.c.GetAndDelete(key)
	return value, ok, nil
}

func (kv *LocalKV) GetAndSet(key string, value string, ttl time.Duration) (string, bool, error) {
	old, ok := kv.c.GetAndSet(key, value, ttl)
	return old, ok, nil
}

func (kv *LocalKV) SetStruct(key string, obj interface{}, ttl time.Duration) error {
	str, err := anyToString(obj)
	if err != nil {
		return err
	}
	kv.Set(key, str, ttl)
	return nil
}

func (kv *LocalKV) GetStruct(key string, p any) error {
	str, ok := kv.Get(key)
	if !ok {
		return nil
	}
	return stringToAny(str, p)
}

func (kv *LocalKV) Del(key string) error {
	kv.c.Del(key)
	return nil
}

func (kv *LocalKV) GetType() Type {
	return Local
}

func (kv *LocalKV) Ping() error {
	return nil
}

func (kv *LocalKV) Incr(key string, i int64, ttl time.Duration, nx bool) (int64, error) {
	v := kv.aic.GetOrCreate(key, ttl, func() *atomic.Int64 {
		return new(atomic.Int64)
	})
	if ttl != 0 && !nx {
		kv.aic.Expire(key, ttl)
	}
	return v.Add(i), nil
}

type LocalLock struct {
	kv     *LocalKV
	key    string
	parse  string
	locked bool
}

func (kv *LocalKV) NewLock(key string) Lock {
	return &LocalLock{
		kv:     kv,
		key:    key,
		parse:  uuid.NewString(),
		locked: false,
	}
}

func (lock *LocalLock) TryLock(ttl time.Duration) bool {
	lock.locked = lock.kv.c.TryLock(lock.key, lock.parse, ttl)
	return lock.locked
}

func (lock *LocalLock) Unlock() error {
	if !lock.locked {
		return nil
	}
	err := lock.kv.c.Unlock(lock.key, lock.parse)
	if err == nil {
		lock.locked = false
	}
	return err
}

func (lock *LocalLock) Lock(ttl, maxWait time.Duration) error {
	err := lock.kv.c.Lock(lock.key, lock.parse, ttl, maxWait)
	if err != nil {
		return err
	}
	lock.locked = true
	return nil
}

func (lock *LocalLock) Refresh(ttl time.Duration) (bool, error) {
	if !lock.locked {
		return false, nil
	}
	ok := lock.kv.c.RefreshLock(lock.key, lock.parse, ttl)
	if !ok {
		lock.locked = false
	}
	return ok, nil
}
