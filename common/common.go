/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/17 10:14
 */

// Package common

package common

import (
	"github.com/real-uangi/allingo/common/ready"
	"github.com/real-uangi/fxtrategy"
)

var Provides = []interface{}{
	initGinEngine,
	initGinCheckpoint,
	ready.NewManager,
	fxtrategy.NewContext[ready.CheckPoint],
}

var Invokes = []interface{}{
	startHttpServer,
	enableHealthCheckApi,
}
