/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/9 10:42
 */

// Package ready

package ready

import (
	"github.com/real-uangi/allingo/common/log"
	"github.com/real-uangi/fxtrategy"
)

type Manager struct {
	ctx    *fxtrategy.Context[CheckPoint]
	logger *log.StdLogger
}

func NewManager(ctx *fxtrategy.Context[CheckPoint]) *Manager {
	return &Manager{
		ctx:    ctx,
		logger: log.For[Manager](),
	}
}
