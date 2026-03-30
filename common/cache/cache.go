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
	key        string
}

func newCacheItem[T any](data T, ttl time.Duration) *cacheItem[T] {
	expiration := int64(0)
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixMilli()
	}
	return &cacheItem[T]{
		Data:       data,
		Expiration: expiration,
	}
}

func (c *Cache[T]) loadValidItemLocked(key string, now int64) (*cacheItem[T], bool) {
	item, ok := c.data[key]
	if !ok {
		return nil, false
	}
	if item.Expiration > 0 && item.Expiration < now {
		delete(c.data, key)
		return nil, false
	}
	return item, true
}

type Cache[T any] struct {
	mu       sync.RWMutex
	data     map[string]*cacheItem[T]
	interval time.Duration
}

func New[T any](interval time.Duration) *Cache[T] {
	if interval < time.Second {
		interval = time.Second
	}
	c := &Cache[T]{
		data:     make(map[string]*cacheItem[T]),
		interval: interval,
	}
	_ = async.SubmitKeepRunning(c.cleanup)
	return c
}

func (c *Cache[T]) Set(key string, value T, ttl time.Duration) {
	c.mu.Lock()
	c.data[key] = newCacheItem(value, ttl)
	c.mu.Unlock()
}

func (c *Cache[T]) Get(key string) (value T, ok bool) {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok {
		return value, false
	}
	return item.Data, true
}

func (c *Cache[T]) GetOrCreate(key string, ttl time.Duration, fallbackFunc func() T) T {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if ok {
		return item.Data
	}
	v := fallbackFunc()
	c.data[key] = newCacheItem(v, ttl)
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

func (c *Cache[T]) SetIfAbsent(key string, value T, ttl time.Duration) bool {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.loadValidItemLocked(key, now); ok {
		return false
	}
	c.data[key] = newCacheItem(value, ttl)
	return true
}

func (c *Cache[T]) CompareAndSet(key string, expected, value T, ttl time.Duration, equal func(left, right T) bool) bool {
	if equal == nil {
		return false
	}
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok || !equal(item.Data, expected) {
		return false
	}
	c.data[key] = newCacheItem(value, ttl)
	return true
}

func (c *Cache[T]) CompareAndDelete(key string, expected T, equal func(left, right T) bool) bool {
	if equal == nil {
		return false
	}
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok || !equal(item.Data, expected) {
		return false
	}
	delete(c.data, key)
	return true
}

func (c *Cache[T]) GetAndDelete(key string) (value T, ok bool) {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok {
		return value, false
	}
	delete(c.data, key)
	return item.Data, true
}

func (c *Cache[T]) GetAndSet(key string, value T, ttl time.Duration) (old T, ok bool) {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if ok {
		old = item.Data
	}
	c.data[key] = newCacheItem(value, ttl)
	return old, ok
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

func (c *Cache[T]) cleanup() error {
	toDelete := make([]string, 0)
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		now := time.Now().UnixMilli()
		select {
		case <-ticker.C:
			c.mu.RLock()
			for k, v := range c.data {
				if v.Expiration > 0 && v.Expiration < now {
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

func (c *Cache[T]) Expire(key string, ttl time.Duration) {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok {
		return
	}
	if ttl > 0 {
		item.Expiration = time.Now().Add(ttl).UnixMilli()
		return
	}
	item.Expiration = 0
}
