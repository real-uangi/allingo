/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/4/9 10:40
 */

// Package ready

package ready

import "github.com/real-uangi/fxtrategy"

type CheckPoint interface {
	Ready() error
	fxtrategy.Nameable
}

const CPGroupName = "ready_checkpoint"
