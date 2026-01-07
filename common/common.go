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
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(initGinEngine),
	fx.Provide(ready.NewManager),
	fxtrategy.ProvideContext[ready.CheckPoint](ready.CPGroupName),
	fx.Invoke(startHttpServer),
	fx.Invoke(enableHealthCheckApi),
)
