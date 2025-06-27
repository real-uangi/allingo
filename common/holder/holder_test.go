/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/9 10:38
 */

// Package holder
package holder

import (
	"sync"
	"testing"
)

func TestHolder(t *testing.T) {
	Put("a", 0)
	Put("b", 1)
	v, ok := Get("a")
	t.Log(v.(int))
	t.Log(ok)
	v, ok = Get("b")
	t.Log(v.(int))
	t.Log(ok)
	var wg sync.WaitGroup
	const loops = 10000
	wg.Add(loops)
	for i := 0; i < loops; i++ {
		j := i
		go func() {
			Put("test", j)
			v, ok := Get("test")
			if !ok {
				t.Error("holder miss!")
			}
			if v.(int) != j {
				t.Errorf("wrong value! expected %d, got %d", j, v.(int))
			}
			wg.Done()
			wg.Wait()
		}()
	}
	wg.Wait()
}

func BenchmarkHolder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Put("test", "test")
	}
}

func BenchmarkGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Put("test", "test")
		Get("test")
	}
}
