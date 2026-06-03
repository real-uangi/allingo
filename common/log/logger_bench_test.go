/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/2/3 13:37
 */

// Package log

package log

import (
	"io"
	"testing"
)

func BenchmarkArgsCopy(b *testing.B) {
	src := []interface{}{"a", 1, 2, 3, "b"}
	dst := make([]interface{}, 0, 8)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		dst = append(dst[:0], src...)
	}
}

func BenchmarkAsyncLoggerInfof(b *testing.B) {
	sl := NewStdLogger("test")
	sl.SetOutput(io.Discard)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sl.Infof("hello %d %s", i, "world")
	}
}
