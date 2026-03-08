/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 11:32
 */

// Package obfs

package obfs

const mask63 int64 = 0x7fffffffffffffff

func rotl63(x int64, k uint) int64 {
	x &= mask63
	k %= 63
	return ((x << k) | (x >> (63 - k))) & mask63
}

func rotr63(x int64, k uint) int64 {
	x &= mask63
	k %= 63
	return ((x >> k) | (x << (63 - k))) & mask63
}

// ScrambleInt
// generate your own secret using "openssl rand -hex 8" output like "9e3779b97f4a7c15"
// Then define it like "const xorSecret uint64 = 0x9e3779b97f4a7c15"
func ScrambleInt(x int64, secret int64) int64 {
	x &= mask63
	secret &= mask63

	x ^= secret
	x = rotl63(x, 17)

	return x
}

func DescrambleInt(x int64, secret int64) int64 {
	x &= mask63
	secret &= mask63

	x = rotr63(x, 17)
	x ^= secret

	return x
}
