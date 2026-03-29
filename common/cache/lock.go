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
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.loadValidItemLocked(key, now); ok {
		return false
	}

	var empty T
	item := newCacheItem(empty, ttl)
	item.key = parse
	c.data[key] = item
	return true
}

// Unlock releases the lock for the given key.
func (c *Cache[T]) Unlock(key, parse string) error {
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok {
		return nil
	}
	if item.key != parse {
		return business.NewError("can't unlock! current key is holding by others")
	}
	delete(c.data, key)
	return nil
}

func (c *Cache[T]) RefreshLock(key, parse string, ttl time.Duration) bool {
	if ttl <= 0 {
		return false
	}
	now := time.Now().UnixMilli()
	c.mu.Lock()
	defer c.mu.Unlock()
	item, ok := c.loadValidItemLocked(key, now)
	if !ok || item.key != parse {
		return false
	}
	item.Expiration = time.Now().Add(ttl).UnixMilli()
	return true
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
