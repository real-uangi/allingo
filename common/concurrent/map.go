/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/22 12:52
 */

// Package concurrent
package concurrent

const ShardCount = 32

// Map A "thread" safe map of type string:Anything.
// To avoid lock bottlenecks this map is dived to several (ShardCount) map shards.
type Map[K comparable, V any] struct {
	shards   []*shared[K, V]
	sharding func(key K) uint32
}

func create[K comparable, V any](sharding shardingFunc[K]) Map[K, V] {
	m := Map[K, V]{
		sharding: sharding,
		shards:   make([]*shared[K, V], ShardCount),
	}
	for i := 0; i < ShardCount; i++ {
		m.shards[i] = &shared[K, V]{items: make(map[K]V)}
	}
	return m
}

// NewStrMap Creates a new concurrent map.
func NewStrMap[V any]() Map[string, V] {
	return create[string, V](fnv32)
}

func NewInt64Map[V any]() Map[int64, V] {
	return create[int64, V](hashInt)
}

// getShard returns shard under given key
func (m Map[K, V]) getShard(key K) *shared[K, V] {
	return m.shards[uint(m.sharding(key))%uint(ShardCount)]
}

func (m Map[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		m.getShard(key).Set(key, value)
	}
}

// Set Sets the given value under the specified key.
func (m Map[K, V]) Set(key K, value V) {
	m.getShard(key).Set(key, value)
}

// SetIfAbsent Sets the given value under the specified key if no value was associated with it.
func (m Map[K, V]) SetIfAbsent(key K, value V) bool {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

func (m Map[K, V]) GetOrCreate(key K, createFunc func() V) V {
	// Get map shard.
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()
	val, ok := shard.items[key]
	if ok {
		return val
	}
	val = createFunc()
	shard.items[key] = val
	return val
}

// Get retrieves an element from map under given key.
func (m Map[K, V]) Get(key K) (V, bool) {
	return m.getShard(key).Get(key)
}

// Len returns the number of elements within the map.
func (m Map[K, V]) Len() int {
	count := 0
	for i := 0; i < ShardCount; i++ {
		count += m.shards[i].Len()
	}
	return count
}

// HasKey Looks up an item under specified key
func (m Map[K, V]) HasKey(key K) bool {
	_, ok := m.getShard(key).Get(key)
	return ok
}

// Remove removes an element from the map.
func (m Map[K, V]) Remove(key K) {
	m.getShard(key).Delete(key)
}

// Pop removes an element from the map and returns it
func (m Map[K, V]) Pop(key K) (v V, exists bool) {
	// Try to get shard.
	shard := m.getShard(key)
	v, exists = shard.Get(key)
	if exists {
		shard.Delete(key)
	}
	return v, exists
}

// IsEmpty checks if map is empty.
func (m Map[K, V]) IsEmpty() bool {
	return m.Len() == 0
}

// Keys returns all keys as []string
func (m Map[K, V]) Keys() []K {
	count := m.Len()
	keys := make([]K, 0, count)
	for _, shard := range m.shards {
		keys = append(keys, shard.Keys()...)
	}
	return keys
}
