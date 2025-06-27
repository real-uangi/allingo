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
)

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
	wrapper.Args = nil
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
	logQueue <- wrapper
}

func (wrapper *asyncLogWrapper) flush() {
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
