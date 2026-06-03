/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 16:47
 */

// Package log
package log

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/real-uangi/allingo/common/goid"
)

type Level uint8

const (
	PanicLevel Level = iota
	FatalLevel
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
	TraceLevel
)

type Fields map[string]interface{}

type Entry struct {
	Time       time.Time
	Level      Level
	LoggerName string
	GoID       int64
	Message    string
	Err        error
	Fields     Fields
}

type Interface interface {
	WithField(key string, value interface{}) *LogWrapper
	WithFields(fields Fields) *LogWrapper

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(err error, format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Panicf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(err error, args ...interface{})
	Fatal(args ...interface{})
	Panic(args ...interface{})
	Print(v ...interface{})
}

type loggerState struct {
	mu  sync.Mutex
	out io.Writer
}

type StdLogger struct {
	state       *loggerState
	name        string
	middleInfos [][]byte
}

var (
	_ Interface = (*StdLogger)(nil)
	_ Interface = (*LogWrapper)(nil)
)

const (
	FieldLoggerName = "logger_name"
	FieldGoId       = "go_id"
)

func NewStdLogger(name string) *StdLogger {
	return &StdLogger{
		state:       &loggerState{out: os.Stdout},
		name:        name,
		middleInfos: newMiddleInfos(name),
	}
}

func For[T any]() *StdLogger {
	var t T
	typ := reflect.TypeOf(&t).Elem()
	return Of(typ)
}

func Of(typ reflect.Type) *StdLogger {
	name := typ.Name()
	if name == "" {
		name = "undefined"
	}
	paths := strings.Split(typ.PkgPath(), "/")
	var builder strings.Builder
	for i, v := range paths {
		if i == 0 && strings.Contains(v, ".") {
			continue
		}
		builder.WriteString(v)
		builder.WriteString(".")
	}
	builder.WriteString(name)
	return NewStdLogger(builder.String())
}

func (sl *StdLogger) SetOutput(w io.Writer) {
	sl.state.mu.Lock()
	defer sl.state.mu.Unlock()
	sl.state.out = w
}

func (sl *StdLogger) WithField(key string, value interface{}) *LogWrapper {
	return sl.wrapper().WithField(key, value)
}

func (sl *StdLogger) WithFields(fields Fields) *LogWrapper {
	return sl.wrapper().WithFields(fields)
}

func (sl *StdLogger) Tracef(format string, args ...interface{}) {
	sl.wrapper().logf(TraceLevel, nil, format, args...)
}

func (sl *StdLogger) Debugf(format string, args ...interface{}) {
	sl.wrapper().logf(DebugLevel, nil, format, args...)
}

func (sl *StdLogger) Infof(format string, args ...interface{}) {
	sl.wrapper().logf(InfoLevel, nil, format, args...)
}

func (sl *StdLogger) Warnf(format string, args ...interface{}) {
	sl.wrapper().logf(WarnLevel, nil, format, args...)
}

func (sl *StdLogger) Errorf(err error, format string, args ...interface{}) {
	sl.wrapper().logf(ErrorLevel, err, format, args...)
}

func (sl *StdLogger) Fatalf(format string, args ...interface{}) {
	sl.wrapper().logf(FatalLevel, nil, format, args...)
}

func (sl *StdLogger) Panicf(format string, args ...interface{}) {
	sl.wrapper().logf(PanicLevel, nil, format, args...)
}

func (sl *StdLogger) Trace(args ...interface{}) {
	sl.wrapper().log(TraceLevel, nil, args...)
}

func (sl *StdLogger) Debug(args ...interface{}) {
	sl.wrapper().log(DebugLevel, nil, args...)
}

func (sl *StdLogger) Info(args ...interface{}) {
	sl.wrapper().log(InfoLevel, nil, args...)
}

func (sl *StdLogger) Warn(args ...interface{}) {
	sl.wrapper().log(WarnLevel, nil, args...)
}

func (sl *StdLogger) Error(err error, args ...interface{}) {
	if len(args) == 0 && err != nil {
		args = []interface{}{err.Error()}
	}
	sl.wrapper().log(ErrorLevel, err, args...)
}

func (sl *StdLogger) Fatal(args ...interface{}) {
	sl.wrapper().log(FatalLevel, nil, args...)
}

func (sl *StdLogger) Panic(args ...interface{}) {
	sl.wrapper().log(PanicLevel, nil, args...)
}

func (sl *StdLogger) Print(v ...interface{}) {
	sl.wrapper().log(InfoLevel, nil, strings.TrimSuffix(fmt.Sprint(v...), "\n"))
}

func (sl *StdLogger) wrapper() *LogWrapper {
	wrapper := getLogWrapper()
	wrapper.logger = sl
	wrapper.time = time.Now()
	wrapper.goID = goid.Get()
	return wrapper
}
