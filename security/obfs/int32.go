/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/3/8 10:56
 */

// Package obfs

package obfs

const mask31 int32 = 0x7fffffff

func rotl31(x int32, k uint) int32 {
	x &= mask31
	k %= 31
	return ((x << k) | (x >> (31 - k))) & mask31
}

func rotr31(x int32, k uint) int32 {
	x &= mask31
	k %= 31
	return ((x >> k) | (x << (31 - k))) & mask31
}

func ScrambleInt32(x int32, secret int32) int32 {
	x &= mask31
	secret &= mask31

	x ^= secret
	x = rotl31(x, 11)

	return x
}

func DescrambleInt32(x int32, secret int32) int32 {
	x &= mask31
	secret &= mask31

	x = rotr31(x, 11)
	x ^= secret

	return x
}
