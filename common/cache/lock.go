/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/7/9 17:17
 */

// Package cache

package cache

import (
	"github.com/real-uangi/allingo/common/business"
	"time"
)

// TryLock attempts to acquire a lock for the given key with the specified TTL.
// Returns true if the lock was acquired, false otherwise.
func (c *Cache[T]) TryLock(key, parse string, ttl time.Duration) bool {
	now := time.Now().UnixMilli()
	expiration := now + ttl.Milliseconds()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.data[key]
	if ok && item.Expiration > now {
		return false
	}

	var empty T
	c.data[key] = &cacheItem[T]{
		Data:       empty,
		Expiration: expiration,
		key:        parse,
	}
	return true
}

// Unlock releases the lock for the given key.
func (c *Cache[T]) Unlock(key, parse string) error {
	c.mu.RLock()
	item, ok := c.data[key]
	c.mu.RUnlock()
	if item.Expiration < time.Now().UnixMilli() {
		return nil
	}
	if ok && item.key != parse {
		return business.NewError("can't unlock! current key is holding by others")
	}
	c.Del(key)
	return nil
}

// Lock acquires a lock for the given key with the specified TTL, blocking until successful.
func (c *Cache[T]) Lock(key, parse string, ttl, maxWait time.Duration) error {
	deadline := time.Now().Add(maxWait)
	for {
		if c.TryLock(key, parse, ttl) {
			return nil
		}
		if time.Now().After(deadline) {
			return business.NewError("lock timeout")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
