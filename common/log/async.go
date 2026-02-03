/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/14 15:47
 */

// Package log

package log

import (
	"github.com/sirupsen/logrus"
	"sync"
	"sync/atomic"
)

var dropped = new(atomic.Int64)

func DroppedCount() int64 {
	return dropped.Load()
}

var logWrapperPool = sync.Pool{
	New: func() any {
		return &asyncLogWrapper{}
	},
}

func getLogWrapper() *asyncLogWrapper {
	return logWrapperPool.Get().(*asyncLogWrapper)
}

func putLogWrapper(wrapper *asyncLogWrapper) {
	// 清空内容，避免内存泄漏或旧数据影响
	wrapper.Level = 0
	wrapper.Entry = nil
	wrapper.Format = ""
	if wrapper.Args != nil {
		wrapper.Args = wrapper.Args[:0]
	}
	logWrapperPool.Put(wrapper)
}

var logQueue = make(chan *asyncLogWrapper, 1024)

type asyncLogWrapper struct {
	Level  logrus.Level
	Entry  *logrus.Entry
	Format string
	Args   []interface{}
}

func (wrapper *asyncLogWrapper) queue() {
	switch wrapper.Level {
	case logrus.DebugLevel, logrus.InfoLevel, logrus.TraceLevel:
		wrapper.bestEffortQueue()
	default:
		wrapper.mustQueue()
	}
}

func (wrapper *asyncLogWrapper) mustQueue() {
	logQueue <- wrapper
}

func (wrapper *asyncLogWrapper) bestEffortQueue() {
	select {
	case logQueue <- wrapper:
		return
	default:
		// 队列满，丢弃低级别日志
		putLogWrapper(wrapper)
		dropped.Add(1)
		return
	}
}

func (wrapper *asyncLogWrapper) flush() {
	defer func() {
		_ = recover()
	}()
	defer putLogWrapper(wrapper)
	if wrapper.Entry == nil {
		return
	}

	if wrapper.Format == "" {
		wrapper.Entry.Log(wrapper.Level, wrapper.Args...)
	} else {
		wrapper.Entry.Logf(wrapper.Level, wrapper.Format, wrapper.Args...)
	}
}

func init() {
	go handleLog()
}

func handleLog() {
	for wrapper := range logQueue {
		wrapper.flush()
	}
}
