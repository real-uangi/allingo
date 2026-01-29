/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 11:32
 */

// Package obfs

package obfs

import "math/bits"

// ScrambleInt
// generate your own secret using "openssl rand -hex 8" output like "9e3779b97f4a7c15"
// Then define it like "const xorSecret uint64 = 0x9e3779b97f4a7c15"
func ScrambleInt(x uint64, secret uint64) uint64 {
	x ^= secret
	return bits.RotateLeft64(x, 17)
}

func DescrambleInt(x uint64, secret uint64) uint64 {
	x = bits.RotateLeft64(x, -17)
	return x ^ secret
}
