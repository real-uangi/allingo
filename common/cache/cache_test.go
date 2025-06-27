/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/7/25 09:59
 */

// Package cache
package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {

	c := New[string](500 * time.Millisecond)

	c.Set("a", "aaa", 2*time.Second)
	c.Set("b", "bbb", 3*time.Second)

	t.Log(c.GetOrDefault("a", "fallback a"))
	t.Log(c.GetOrDefault("c", "fallback c"))

	t.Log(c.GetOrCreate("a", time.Second, func() string {
		return "create a"
	}))
	t.Log(c.GetOrCreate("c", time.Second, func() string {
		return "create c"
	}))

	t.Log(c.Get("a"))
	t.Log(c.Get("b"))
	t.Log(c.Get("c"))

	for {
		ks := c.Keys()
		if len(ks) == 0 {
			break
		}
		t.Log(ks)
		time.Sleep(time.Second)
	}

}
