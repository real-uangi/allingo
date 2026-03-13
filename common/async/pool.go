/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/3/13 08:35
 */

// Package async

package async

import (
	"github.com/panjf2000/ants/v2"
	"github.com/real-uangi/allingo/common/env"
	"runtime"
)

var pool *ants.Pool

func init() {
	var err error
	poolSize := env.GetIntOrDefault("ALLINGO_ASYNC_POOL_SIZE", 4*runtime.NumCPU())
	pool, err = ants.NewPool(poolSize)
	if err != nil {
		panic(err)
	}
}
