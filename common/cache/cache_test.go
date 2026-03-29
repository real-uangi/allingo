/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/7/25 09:59
 */

// Package cache
package cache

import (
	"sync"
	"sync/atomic"
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

func TestLock(t *testing.T) {

	c := New[string](500 * time.Millisecond)

	start := time.Now()
	err := c.Lock("A", "KEY-A", time.Second, time.Second)
	if err != nil {
		t.Error(err)
	}

	ok := c.TryLock("A", "KEY-A", time.Second)
	if !ok {
		t.Log("lock is current locked, this is right")
	} else {
		t.Error("locking failed on previous lock")
	}

	err = c.Lock("A", "KEY-A", 10*time.Second, time.Second)
	if err != nil {
		t.Error(err)
	}
	t.Log(time.Now().Sub(start).String())

	err = c.Unlock("A", "KEY-A")
	if err != nil {
		t.Error(err)
	}
	t.Log("unlocked at", time.Now().Sub(start).String())

	err = c.Lock("A", "KEY-A", 10*time.Second, time.Second)
	if err != nil {
		t.Error(err)
	}
	t.Log("locked at", time.Now().Sub(start).String())

	err = c.Unlock("A", "KEY-B")
	if err != nil {
		t.Log("cannot unlock with wrong key, this is correct ", err)
	} else {
		t.Fatal("unlocked with wrong key, this is NOT correct ", time.Now().Sub(start).String())
	}

	err = c.Lock("A", "KEY-A", 10*time.Second, 20*time.Second)
	if err != nil {
		t.Error(err)
	}
	t.Log("locked at", time.Now().Sub(start).String())

}

func TestGetOrCreateIsAtomic(t *testing.T) {
	c := New[int](500 * time.Millisecond)

	var created atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			value := c.GetOrCreate("shared", time.Second, func() int {
				time.Sleep(time.Millisecond)
				created.Add(1)
				return 42
			})
			if value != 42 {
				t.Errorf("unexpected value: %d", value)
			}
		}()
	}

	close(start)
	wg.Wait()

	if got := created.Load(); got != 1 {
		t.Fatalf("expected fallback to run once, got %d", got)
	}
}
