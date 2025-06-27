/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/22 13:49
 */

// Package concurrent
package concurrent

import (
	"strconv"
	"sync"
	"testing"
)

//BenchmarkNative
//BenchmarkNative-10            	 9307429	       160.2 ns/op
//BenchmarkNativeRead
//BenchmarkNativeRead-10        	11864608	       122.9 ns/op
//BenchmarkConcurrent
//BenchmarkConcurrent-10        	 9741102	       200.0 ns/op
//BenchmarkConcurrentRead
//BenchmarkConcurrentRead-10    	 8789556	       145.2 ns/op

var native = make(map[string]bool)
var concurrentMap = NewStrMap[bool]()

func TestSafe(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1000)
	for i := 0; i < 1000; i++ {
		k := i
		go func() {
			defer wg.Done()
			concurrentMap.Set(strconv.Itoa(k), true)
		}()
	}
	wg.Wait()
	t.Log(concurrentMap.Len())
	t.Log(concurrentMap.Keys())
	v, _ := concurrentMap.Get("1")
	t.Log(v)
}

func BenchmarkNative(b *testing.B) {
	for i := 0; i < b.N; i++ {
		native[strconv.Itoa(i)] = true
	}
}

func BenchmarkNativeRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = native[strconv.Itoa(i)]
	}
}

func BenchmarkConcurrent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		concurrentMap.Set(strconv.Itoa(i), true)
	}
}

func BenchmarkConcurrentRead(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = concurrentMap.Get(strconv.Itoa(i))
	}
}
