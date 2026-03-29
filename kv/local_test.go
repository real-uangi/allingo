package kv

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLocalKVSetIfAbsentConcurrent(t *testing.T) {
	kv := newLocalKV()

	var success atomic.Int32
	var wg sync.WaitGroup
	start := make(chan struct{})

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			ok, err := kv.SetIfAbsent("race", "winner", 80*time.Millisecond)
			if err != nil {
				t.Errorf("SetIfAbsent returned error: %v", err)
				return
			}
			if ok {
				success.Add(1)
			}
		}()
	}

	close(start)
	wg.Wait()

	if got := success.Load(); got != 1 {
		t.Fatalf("expected exactly one winner, got %d", got)
	}

	time.Sleep(100 * time.Millisecond)

	ok, err := kv.SetIfAbsent("race", "winner-2", time.Second)
	if err != nil {
		t.Fatalf("SetIfAbsent after expiration returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected SetIfAbsent to succeed after key expiration")
	}
}

func TestLocalKVAtomicStringOps(t *testing.T) {
	kv := newLocalKV()

	ok, err := kv.SetIfAbsent("name", "value-1", time.Second)
	if err != nil {
		t.Fatalf("SetIfAbsent returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected first SetIfAbsent to succeed")
	}

	ok, err = kv.SetIfAbsent("name", "value-2", time.Second)
	if err != nil {
		t.Fatalf("second SetIfAbsent returned error: %v", err)
	}
	if ok {
		t.Fatal("expected second SetIfAbsent to fail")
	}

	ok, err = kv.CompareAndSet("name", "wrong", "value-2", time.Second)
	if err != nil {
		t.Fatalf("CompareAndSet returned error: %v", err)
	}
	if ok {
		t.Fatal("expected CompareAndSet to fail on mismatched value")
	}

	ok, err = kv.CompareAndSet("name", "value-1", "value-2", time.Second)
	if err != nil {
		t.Fatalf("CompareAndSet with matching value returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected CompareAndSet to succeed")
	}

	old, existed, err := kv.GetAndSet("name", "value-3", 80*time.Millisecond)
	if err != nil {
		t.Fatalf("GetAndSet returned error: %v", err)
	}
	if !existed || old != "value-2" {
		t.Fatalf("expected GetAndSet to return previous value-2, got existed=%v old=%q", existed, old)
	}

	time.Sleep(100 * time.Millisecond)
	if _, ok := kv.Get("name"); ok {
		t.Fatal("expected GetAndSet TTL to expire the new value")
	}

	ok, err = kv.SetIfAbsent("name", "value-4", time.Second)
	if err != nil {
		t.Fatalf("SetIfAbsent before GetAndDelete returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected SetIfAbsent to recreate key")
	}

	old, existed, err = kv.GetAndDelete("name")
	if err != nil {
		t.Fatalf("GetAndDelete returned error: %v", err)
	}
	if !existed || old != "value-4" {
		t.Fatalf("expected GetAndDelete to return previous value-4, got existed=%v old=%q", existed, old)
	}
	if _, ok := kv.Get("name"); ok {
		t.Fatal("expected key to be deleted after GetAndDelete")
	}

	ok, err = kv.SetIfAbsent("name", "token", time.Second)
	if err != nil {
		t.Fatalf("SetIfAbsent before CompareAndDelete returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected SetIfAbsent to recreate key for CompareAndDelete")
	}

	ok, err = kv.CompareAndDelete("name", "wrong")
	if err != nil {
		t.Fatalf("CompareAndDelete returned error: %v", err)
	}
	if ok {
		t.Fatal("expected CompareAndDelete to fail on mismatched value")
	}

	value, ok := kv.Get("name")
	if !ok || value != "token" {
		t.Fatalf("expected key to remain after failed CompareAndDelete, got ok=%v value=%q", ok, value)
	}

	ok, err = kv.CompareAndDelete("name", "token")
	if err != nil {
		t.Fatalf("CompareAndDelete with matching value returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected CompareAndDelete to succeed")
	}
	if _, ok := kv.Get("name"); ok {
		t.Fatal("expected key to be deleted after CompareAndDelete")
	}
}

func TestLocalKVLockRefresh(t *testing.T) {
	kv := newLocalKV()

	owner := kv.NewLock("refresh-lock")
	if !owner.TryLock(80 * time.Millisecond) {
		t.Fatal("expected owner to acquire lock")
	}

	other := kv.NewLock("refresh-lock")
	ok, err := other.Refresh(120 * time.Millisecond)
	if err != nil {
		t.Fatalf("non-owner Refresh returned error: %v", err)
	}
	if ok {
		t.Fatal("expected non-owner Refresh to fail")
	}

	time.Sleep(50 * time.Millisecond)

	ok, err = owner.Refresh(120 * time.Millisecond)
	if err != nil {
		t.Fatalf("owner Refresh returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected owner Refresh to succeed")
	}

	time.Sleep(60 * time.Millisecond)

	challenger := kv.NewLock("refresh-lock")
	if challenger.TryLock(50 * time.Millisecond) {
		t.Fatal("expected refreshed lock to still be held")
	}

	time.Sleep(80 * time.Millisecond)

	if !challenger.TryLock(50 * time.Millisecond) {
		t.Fatal("expected refreshed lock to expire eventually")
	}
	if err := challenger.Unlock(); err != nil {
		t.Fatalf("challenger Unlock returned error: %v", err)
	}
}

func TestLocalKVIncrRefreshesTTL(t *testing.T) {
	kv := newLocalKV()
	ttl := 200 * time.Millisecond

	value, err := kv.Incr("counter", 1, ttl, false)
	if err != nil {
		t.Fatalf("first Incr returned error: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected first Incr to return 1, got %d", value)
	}

	time.Sleep(120 * time.Millisecond)

	value, err = kv.Incr("counter", 1, ttl, false)
	if err != nil {
		t.Fatalf("second Incr returned error: %v", err)
	}
	if value != 2 {
		t.Fatalf("expected second Incr to return 2, got %d", value)
	}

	time.Sleep(120 * time.Millisecond)

	value, err = kv.Incr("counter", 1, ttl, false)
	if err != nil {
		t.Fatalf("third Incr returned error: %v", err)
	}
	if value != 3 {
		t.Fatalf("expected TTL refresh to keep counter alive, got %d", value)
	}

	time.Sleep(250 * time.Millisecond)

	value, err = kv.Incr("counter", 1, ttl, false)
	if err != nil {
		t.Fatalf("fourth Incr returned error: %v", err)
	}
	if value != 1 {
		t.Fatalf("expected expired counter to restart at 1, got %d", value)
	}
}
