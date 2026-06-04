/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/6/4 00:00
 */

// Package log
package log

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type FieldFiller func(wrapper *LogWrapper)

var (
	fieldFillerMu     sync.RWMutex
	fieldFillers      []FieldFiller
	fieldFillerErrors atomic.Int64
)

func AddFieldFiller(filler FieldFiller) {
	if filler == nil {
		return
	}

	fieldFillerMu.Lock()
	defer fieldFillerMu.Unlock()
	fieldFillers = append(fieldFillers, filler)
}

func FieldFillerErrorCount() int64 {
	return fieldFillerErrors.Load()
}

func resetFieldFillersForTest() {
	fieldFillerMu.Lock()
	defer fieldFillerMu.Unlock()
	fieldFillers = nil
	fieldFillerErrors.Store(0)
}

func fillGlobalFields(wrapper *LogWrapper) {
	fieldFillerMu.RLock()
	fillers := append([]FieldFiller(nil), fieldFillers...)
	fieldFillerMu.RUnlock()

	for _, filler := range fillers {
		runFieldFiller(filler, wrapper)
	}
}

func runFieldFiller(filler FieldFiller, wrapper *LogWrapper) {
	defer func() {
		if ev := recover(); ev != nil {
			fieldFillerErrors.Add(1)
			fmt.Fprintf(os.Stderr, "log field filler panic: %v\n", ev)
		}
	}()
	filler(wrapper)
}
