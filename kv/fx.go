/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/21 14:59
 */

// Package kv

package kv

import "go.uber.org/fx"

var Module = fx.Module(
	"kv",
	fx.Provide(InitKV),
	CheckPoint,
)
