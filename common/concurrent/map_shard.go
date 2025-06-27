/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/22 13:18
 */

// Package concurrent
package concurrent

import "sync"

// shared A "thread" safe string to anything map.
type shared[K comparable, V any] struct {
	items        map[K]V
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

func (s *shared[K, V]) Set(k K, v V) {
	s.Lock()
	defer s.Unlock()
	s.items[k] = v
}

func (s *shared[K, V]) Get(k K) (v V, ok bool) {
	s.RLock()
	defer s.RUnlock()
	v, ok = s.items[k]
	return
}

func (s *shared[K, V]) Delete(k K) {
	s.Lock()
	defer s.Unlock()
	delete(s.items, k)
}

func (s *shared[K, V]) Len() int {
	s.RLock()
	defer s.RUnlock()
	return len(s.items)
}

func (s *shared[K, V]) Keys() []K {
	s.RLock()
	defer s.RUnlock()
	keys := make([]K, 0, len(s.items))
	for k := range s.items {
		keys = append(keys, k)
	}
	return keys
}
