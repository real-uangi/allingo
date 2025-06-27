/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 16:28
 */

// Package app

package app

import (
	"github.com/real-uangi/allingo/common/log"
)

var std *App

func init() {
	std = &App{
		onFxStart: make([]func(), 0),
		onFxStop:  make([]func(), 0),
		logger:    log.For[App](),
	}
}

func Current() *App {
	return std
}
