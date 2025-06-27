/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 17:26
 */

// Package kv

package kv

import (
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

func (kv *LocalKV) Del(key string) error {
	kv.c.Del(key)
	return nil
}
