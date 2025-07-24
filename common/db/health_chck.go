/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/10 12:24
 */

// Package db

package db

import (
	"github.com/real-uangi/allingo/common/ready"
	"github.com/real-uangi/fxtrategy"
)

type checkpoint struct {
	manager *Manager
}

func (c *checkpoint) Ready() error {
	sqlDB, err := c.manager.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func newCheckpoint(manager *Manager) *checkpoint {
	return &checkpoint{
		manager: manager,
	}
}

// CheckPoint 数据库健康检测
func CheckPoint(manager *Manager) fxtrategy.Strategy[ready.CheckPoint] {
	return fxtrategy.Strategy[ready.CheckPoint]{
		NS: fxtrategy.NamedStrategy[ready.CheckPoint]{
			Name: "db",
			Item: newCheckpoint(manager),
		},
	}
}
