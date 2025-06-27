/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 16:47
 */

// Package log
package log

import (
	"fmt"
	"github.com/real-uangi/allingo/common/goid"
	"github.com/sirupsen/logrus"
	"os"
	"reflect"
	"strings"
)

type StdLogger struct {
	logger *logrus.Logger
	name   string
}

const (
	FieldLoggerName = "logger_name"
	FieldGoId       = "go_id"
)

func NewStdLogger(name string) *StdLogger {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	logger.SetFormatter(newFormatter(name))
	logger.SetOutput(os.Stdout)
	return &StdLogger{logger, name}
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

func (sl *StdLogger) Basic() *logrus.Entry {
	return sl.logger.WithFields(map[string]interface{}{
		FieldLoggerName: sl.name,
		FieldGoId:       goid.Get(),
	})
}

func (sl *StdLogger) Tracef(format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.TraceLevel
	wrapper.Entry = sl.Basic()
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Debugf(format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.DebugLevel
	wrapper.Entry = sl.Basic()
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Infof(format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.InfoLevel
	wrapper.Entry = sl.Basic()
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Warnf(format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.WarnLevel
	wrapper.Entry = sl.Basic()
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Errorf(err error, format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.ErrorLevel
	wrapper.Entry = sl.Basic().WithError(err)
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Fatalf(format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.FatalLevel
	wrapper.Entry = sl.Basic()
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Panicf(format string, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.PanicLevel
	wrapper.Entry = sl.Basic()
	wrapper.Format = format
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Trace(args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.TraceLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Debug(args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.DebugLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Info(args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.InfoLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Warn(args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.WarnLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Error(err error, args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.ErrorLevel
	wrapper.Entry = sl.Basic().WithError(err)
	if len(args) == 0 {
		wrapper.Args = []interface{}{err.Error()}
	} else {
		wrapper.Args = args
	}
	wrapper.queue()
}

func (sl *StdLogger) Fatal(args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.FatalLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Panic(args ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.PanicLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = args
	wrapper.queue()
}

func (sl *StdLogger) Print(v ...interface{}) {
	wrapper := getLogWrapper()
	wrapper.Level = logrus.InfoLevel
	wrapper.Entry = sl.Basic()
	wrapper.Args = []interface{}{strings.TrimSuffix(fmt.Sprint(v...), "\n")}
	wrapper.queue()
}
