/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/6/3 00:00
 */

// Package log
package log

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type HookFunc func(Entry) error

type hookWrapper struct {
	levels map[Level]struct{}
	hook   HookFunc
}

var (
	hookMu     sync.RWMutex
	hooks      []hookWrapper
	hookErrors atomic.Int64
)

func AddHook(levels []Level, hook HookFunc) {
	if len(levels) == 0 || hook == nil {
		return
	}
	levelSet := make(map[Level]struct{}, len(levels))
	for _, level := range levels {
		levelSet[level] = struct{}{}
	}

	hookMu.Lock()
	defer hookMu.Unlock()
	hooks = append(hooks, hookWrapper{
		levels: levelSet,
		hook:   hook,
	})
}

func HookErrorCount() int64 {
	return hookErrors.Load()
}

func resetHooksForTest() {
	hookMu.Lock()
	defer hookMu.Unlock()
	hooks = nil
	hookErrors.Store(0)
}

func fireHooks(entry Entry) {
	hookMu.RLock()
	matched := make([]HookFunc, 0, len(hooks))
	for _, registered := range hooks {
		if _, ok := registered.levels[entry.Level]; ok {
			matched = append(matched, registered.hook)
		}
	}
	hookMu.RUnlock()

	for _, hook := range matched {
		fireHook(hook, entry)
	}
}

func fireHook(hook HookFunc, entry Entry) {
	defer func() {
		if ev := recover(); ev != nil {
			hookErrors.Add(1)
			fmt.Fprintf(os.Stderr, "log hook panic: %v\n", ev)
		}
	}()

	entry.Fields = cloneFields(entry.Fields)
	if err := hook(entry); err != nil {
		hookErrors.Add(1)
		fmt.Fprintf(os.Stderr, "log hook error: %v\n", err)
	}
}
