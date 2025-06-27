/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/18 16:20
 */

// Package db
package db

import (
	"context"
	"errors"
	"github.com/real-uangi/allingo/common/env"
	"github.com/real-uangi/allingo/common/log"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"strings"
	"time"
)

var isDev = false

func init() {
	isDev = env.Get(env.RunningMode) != env.ReleaseMode
}

type dbLogger struct {
	log.StdLogger
	spaceReplacer *strings.Replacer
}

func (l *dbLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	return l
}

func (l *dbLogger) Info(ctx context.Context, s string, i ...interface{}) {
	l.StdLogger.Infof(s, i...)
}

func (l *dbLogger) Warn(ctx context.Context, s string, i ...interface{}) {
	l.StdLogger.Warnf(s, i...)
}

func (l *dbLogger) Error(ctx context.Context, s string, i ...interface{}) {
	l.StdLogger.Errorf(nil, s, i...)
}

func (l *dbLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, _ := fc()
		l.Errorf(err, "error occurs when executing sql :%s", sql)
	} else if isDev {
		sql, rows := fc()
		if len(sql) < 6 {
			return
		}
		l.Infof("[%d rows in %.3f s] %s", rows, time.Since(begin).Seconds(), l.spaceReplacer.Replace(sql))
	}
}

func newLogger() *dbLogger {
	return &dbLogger{
		StdLogger:     *log.For[dbLogger](),
		spaceReplacer: strings.NewReplacer("\n", " ", "\r", " "),
	}
}
