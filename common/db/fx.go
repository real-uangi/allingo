/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/21 14:57
 */

// Package db

package db

import "go.uber.org/fx"

var Module = fx.Module(
	"db",
	fx.Provide(InitDataSource),
	CheckPoint,
)
