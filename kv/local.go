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
	"time"
)

type LocalKV struct {
	c      *cache.Cache[string]
	logger *log.StdLogger
}

func newLocalKV() *LocalKV {
	logger := log.For[LocalKV]()
	logger.Warn("using local kv, standalone mode only")
	return &LocalKV{
		c:      cache.New[string](time.Minute),
		logger: logger,
	}
}

func (kv *LocalKV) Set(key string, value string, ttl time.Duration) {
	kv.c.Set(key, value, ttl)
}

func (kv *LocalKV) Get(key string) (string, bool) {
	return kv.c.Get(key)
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
	return lock.kv.c.Unlock(lock.key, lock.parse)
}

func (lock *LocalLock) Lock(ttl, maxWait time.Duration) error {
	return lock.kv.c.Lock(lock.key, lock.parse, ttl, maxWait)
}
