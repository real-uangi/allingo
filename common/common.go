/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/17 10:14
 */

// Package common

package common

var Provides = []interface{}{
	initGinEngine,
}

var Invokes = []interface{}{
	startHttpServer,
}
