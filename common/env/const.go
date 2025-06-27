/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 10:10
 */

// Package env

package env

import "github.com/gin-gonic/gin"

const RunningMode = "RUN_MODE"

// easier for maintainers
const (
	DebugMode   = gin.DebugMode
	ReleaseMode = gin.ReleaseMode
)
