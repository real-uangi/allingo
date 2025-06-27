/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/7/25 09:29
 */

// Package cache
package cache

import (
	"github.com/real-uangi/allingo/common/async"
	"sync"
	"time"
)

type cacheItem[T any] struct {
	Data       T
	Expiration int64
}

func newCacheItem[T any](data T, ttl time.Duration) *cacheItem[T] {
	return &cacheItem[T]{
		Data:       data,
		Expiration: time.Now().Add(ttl).UnixMilli(),
	}
}

type Cache[T any] struct {
	mu       sync.RWMutex
	data     map[string]*cacheItem[T]
	interval time.Duration
}

func New[T any](interval time.Duration) *Cache[T] {
	c := &Cache[T]{
		data:     make(map[string]*cacheItem[T]),
		interval: interval,
	}
	async.Go(c.cleanup, true)
	return c
}

func (c *Cache[T]) Set(key string, value T, ttl time.Duration) {
	c.mu.Lock()
	c.data[key] = newCacheItem(value, ttl)
	c.mu.Unlock()
}

func (c *Cache[T]) Get(key string) (value T, ok bool) {
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()
	if ok {
		value = item.Data
	}
	return
}

func (c *Cache[T]) GetOrCreate(key string, ttl time.Duration, fallbackFunc func() T) T {
	v, ok := c.Get(key)
	if ok {
		return v
	}
	v = fallbackFunc()
	c.Set(key, v, ttl)
	return v
}

func (c *Cache[T]) GetOrDefault(key string, fallback T) T {
	v, ok := c.Get(key)
	if ok {
		return v
	}
	return fallback
}

func (c *Cache[T]) Del(key string) {
	c.mu.Lock()
	delete(c.data, key)
	c.mu.Unlock()
}

func (c *Cache[T]) Keys() []string {
	c.mu.RLock()
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	c.mu.RUnlock()
	return keys
}

func (c *Cache[T]) Len() int {
	c.mu.RLock()
	length := len(c.data)
	c.mu.RUnlock()
	return length
}

func (c *Cache[T]) cleanup() {
	toDelete := make([]string, 0)
	ticker := time.NewTicker(c.interval)
	for {
		now := time.Now().UnixMilli()
		select {
		case <-ticker.C:
			c.mu.RLock()
			for k, v := range c.data {
				if v.Expiration < now {
					toDelete = append(toDelete, k)
				}
			}
			c.mu.RUnlock()
			if len(toDelete) > 0 {
				c.mu.Lock()
				for _, k := range toDelete {
					delete(c.data, k)
				}
				c.mu.Unlock()
				toDelete = toDelete[:0]
			}
		}
	}
}

func (c *Cache[T]) Purge() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[string]*cacheItem[T])
}
