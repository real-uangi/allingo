/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/14 15:47
 */

// Package log

package log

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var dropped = new(atomic.Int64)

func DroppedCount() int64 {
	return dropped.Load()
}

var logWrapperPool = sync.Pool{
	New: func() any {
		return &LogWrapper{}
	},
}

func getLogWrapper() *LogWrapper {
	return logWrapperPool.Get().(*LogWrapper)
}

func putLogWrapper(wrapper *LogWrapper) {
	// 清空内容，避免内存泄漏或旧数据影响
	wrapper.logger = nil
	wrapper.level = 0
	wrapper.time = time.Time{}
	wrapper.goID = 0
	wrapper.err = nil
	wrapper.format = ""
	wrapper.done = nil
	wrapper.args = wrapper.args[:0]
	wrapper.fields = nil
	logWrapperPool.Put(wrapper)
}

var logQueue = make(chan *LogWrapper, 1024)

type LogWrapper struct {
	logger *StdLogger
	level  Level
	time   time.Time
	goID   int64
	err    error
	format string
	args   []interface{}
	fields Fields
	done   chan struct{}
}

func (wrapper *LogWrapper) WithField(key string, value interface{}) *LogWrapper {
	if wrapper.fields == nil {
		wrapper.fields = make(Fields, 1)
	}
	wrapper.fields[key] = value
	return wrapper
}

func (wrapper *LogWrapper) WithFields(fields Fields) *LogWrapper {
	if len(fields) == 0 {
		return wrapper
	}
	if wrapper.fields == nil {
		wrapper.fields = make(Fields, len(fields))
	}
	for key, value := range fields {
		wrapper.fields[key] = value
	}
	return wrapper
}

func (wrapper *LogWrapper) Tracef(format string, args ...interface{}) {
	wrapper.logf(TraceLevel, nil, format, args...)
}

func (wrapper *LogWrapper) Debugf(format string, args ...interface{}) {
	wrapper.logf(DebugLevel, nil, format, args...)
}

func (wrapper *LogWrapper) Infof(format string, args ...interface{}) {
	wrapper.logf(InfoLevel, nil, format, args...)
}

func (wrapper *LogWrapper) Warnf(format string, args ...interface{}) {
	wrapper.logf(WarnLevel, nil, format, args...)
}

func (wrapper *LogWrapper) Errorf(err error, format string, args ...interface{}) {
	wrapper.logf(ErrorLevel, err, format, args...)
}

func (wrapper *LogWrapper) Fatalf(format string, args ...interface{}) {
	wrapper.logf(FatalLevel, nil, format, args...)
}

func (wrapper *LogWrapper) Panicf(format string, args ...interface{}) {
	wrapper.logf(PanicLevel, nil, format, args...)
}

func (wrapper *LogWrapper) Trace(args ...interface{}) {
	wrapper.log(TraceLevel, nil, args...)
}

func (wrapper *LogWrapper) Debug(args ...interface{}) {
	wrapper.log(DebugLevel, nil, args...)
}

func (wrapper *LogWrapper) Info(args ...interface{}) {
	wrapper.log(InfoLevel, nil, args...)
}

func (wrapper *LogWrapper) Warn(args ...interface{}) {
	wrapper.log(WarnLevel, nil, args...)
}

func (wrapper *LogWrapper) Error(err error, args ...interface{}) {
	if len(args) == 0 && err != nil {
		args = []interface{}{err.Error()}
	}
	wrapper.log(ErrorLevel, err, args...)
}

func (wrapper *LogWrapper) Fatal(args ...interface{}) {
	wrapper.log(FatalLevel, nil, args...)
}

func (wrapper *LogWrapper) Panic(args ...interface{}) {
	wrapper.log(PanicLevel, nil, args...)
}

func (wrapper *LogWrapper) Print(v ...interface{}) {
	wrapper.log(InfoLevel, nil, strings.TrimSuffix(fmt.Sprint(v...), "\n"))
}

func (wrapper *LogWrapper) logf(level Level, err error, format string, args ...interface{}) {
	wrapper.level = level
	wrapper.err = err
	wrapper.format = format
	wrapper.args = args
	wrapper.queue()
}

func (wrapper *LogWrapper) log(level Level, err error, args ...interface{}) {
	wrapper.level = level
	wrapper.err = err
	wrapper.args = args
	wrapper.queue()
}

func (wrapper *LogWrapper) queue() {
	if isBestEffortLevel(wrapper.level) {
		wrapper.bestEffortQueue()
		return
	}
	wrapper.mustQueue()
}

func (wrapper *LogWrapper) mustQueue() {
	logQueue <- wrapper
}

func (wrapper *LogWrapper) bestEffortQueue() {
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

func (wrapper *LogWrapper) flush() {
	defer func() {
		if ev := recover(); ev != nil {
			fmt.Fprintf(os.Stderr, "failed to flush log: %v\n", ev)
		}
		putLogWrapper(wrapper)
	}()
	if wrapper.logger == nil {
		return
	}

	entry := wrapper.entry()
	output := formatEntry(entry, wrapper.logger.middleInfos)

	wrapper.logger.state.mu.Lock()
	out := wrapper.logger.state.out
	wrapper.logger.state.mu.Unlock()
	if out == nil {
		out = os.Stdout
	}
	_, _ = out.Write(output)

	fireHooks(entry)
}

func (wrapper *LogWrapper) entry() Entry {
	message := fmt.Sprint(wrapper.args...)
	if wrapper.format != "" {
		message = fmt.Sprintf(wrapper.format, wrapper.args...)
	}
	return Entry{
		Time:       wrapper.time,
		Level:      wrapper.level,
		LoggerName: wrapper.logger.name,
		GoID:       wrapper.goID,
		Message:    message,
		Err:        wrapper.err,
		Fields:     cloneFields(wrapper.fields),
	}
}

func ExitTimeout(second int) {
	done := make(chan struct{})
	wrapper := getLogWrapper()
	wrapper.done = done

	timer := time.NewTimer(time.Duration(second) * time.Second)
	defer timer.Stop()

	select {
	case logQueue <- wrapper:
	case <-timer.C:
		putLogWrapper(wrapper)
		return
	}

	select {
	case <-done:
	case <-timer.C:
	}
}

func init() {
	go handleLog()
}

func handleLog() {
	for wrapper := range logQueue {
		if wrapper.done != nil {
			close(wrapper.done)
			putLogWrapper(wrapper)
			continue
		}
		// 保存终止前需要的信息（flush 会清空 wrapper）
		isFatal := wrapper.level == FatalLevel
		isPanic := wrapper.level == PanicLevel
		var panicMsg string
		if isPanic {
			panicMsg = fmt.Sprint(wrapper.args...)
			if wrapper.format != "" {
				panicMsg = fmt.Sprintf(wrapper.format, wrapper.args...)
			}
		}
		wrapper.flush()
		if isFatal {
			os.Exit(1)
		}
		if isPanic {
			panic(panicMsg)
		}
	}
}

func isBestEffortLevel(level Level) bool {
	return level == DebugLevel || level == InfoLevel || level == TraceLevel
}
