package goid

import (
	"sync"
	"testing"
)

func TestGet(t *testing.T) {
	m := make(map[int64]bool)
	var mu sync.Mutex
	const total = 100000
	var wg = sync.WaitGroup{}
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			j := Get()
			mu.Lock()
			if m[j] == true {
				t.Fail()
			}
			m[j] = true
			wg.Done()
			mu.Unlock()
			wg.Wait()
		}()
	}
	wg.Wait()
}

// defer safe
func TestDefer(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	defer t.Log(Get())
	t.Log(Get())
	t.Log(Get())
	go func() {
		defer wg.Done()
		defer func() {
			t.Log("async ", Get())
		}()
		t.Log("async ", Get())
		t.Log("async ", Get())
	}()
	wg.Wait()
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Get()
	}
}
