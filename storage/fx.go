/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/21 15:00
 */

// Package storage

package storage

import "go.uber.org/fx"

var Module = fx.Module(
	"storage",
	fx.Provide(InitStorage),
)
