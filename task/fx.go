/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/21 15:01
 */

// Package task

package task

import "go.uber.org/fx"

var Module = fx.Module(
	"task",
	fx.Provide(NewManager),
)
