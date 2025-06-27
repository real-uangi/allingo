/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/18 15:45
 */

// Package db
package db

import (
	"errors"
	"github.com/real-uangi/allingo/common/env"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"time"
)

// Manager 数据库连接管理 database/sql 级别的DB,Stmt都是并发安全的
type Manager struct {
	db *gorm.DB
}

var logger = newLogger()

func InitDataSource() (*Manager, error) {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		return nil, errors.New("DB_DSN environment variable not set")
	}

	gormConfig := &gorm.Config{
		Logger:                 logger,
		AllowGlobalUpdate:      false,
		PrepareStmt:            true,
		SkipDefaultTransaction: false,
	}
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, err
	}

	connMaxIdleTime := env.GetIntOrDefault("DB_CONN_MAX_IDLE_TIME", 180)
	maxIdleConn := env.GetIntOrDefault("DB_MAX_IDLE_CONN", 10)
	maxOpenConn := env.GetIntOrDefault("DB_MAX_OPEN_CONN", 20)
	//set conn pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)
	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetMaxOpenConns(maxOpenConn)

	return &Manager{
		db: db,
	}, nil
}

func (manager *Manager) GetDB() *gorm.DB {
	return manager.db
}
