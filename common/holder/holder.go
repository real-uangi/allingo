/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/9 10:23
 */

// Package holder
package holder

import (
	"github.com/real-uangi/allingo/common/concurrent"
	"github.com/real-uangi/allingo/common/goid"
	"sync"
)

type holder struct {
	data map[string]any
	mu   sync.RWMutex
}

func (h *holder) get(key string) (v any, ok bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	v, ok = h.data[key]
	return
}

func (h *holder) set(key string, value any) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.data[key] = value
}

var data = concurrent.NewInt64Map[*holder]()

func Put(key string, value any) {
	validate(goid.Get()).set(key, value)
}

func Get(key string) (v any, ok bool) {
	return validate(goid.Get()).get(key)
}

func GetSpecific(key string, gid int64) (v any, ok bool) {
	return validate(gid).get(key)
}

func validate(gid int64) *holder {
	return data.GetOrCreate(gid, newHolder)
}

func newHolder() *holder {
	return &holder{
		data: make(map[string]any),
		mu:   sync.RWMutex{},
	}
}

func Clear() {
	data.Remove(goid.Get())
}

func Size() int {
	return data.Len()
}
