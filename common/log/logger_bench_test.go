/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/2/3 13:37
 */

// Package log

package log

import (
	"github.com/sirupsen/logrus"
	"io"
	"testing"
)

// 防止 stdout IO 干扰测试
func init() {
	logrus.SetOutput(io.Discard)
}

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

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		sl.Infof("hello %d %s", i, "world")
	}
}

func BenchmarkLogrusSyncInfof(b *testing.B) {
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Infof("hello %d %s", i, "world")
	}
}
