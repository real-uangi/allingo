/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/22 13:46
 */

// Package concurrent
package concurrent

type shardingFunc[K comparable] func(key K) uint32

func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func hashInt(key int64) uint32 {
	return uint32(key)
}
