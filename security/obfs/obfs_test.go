/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 11:36
 */

// Package obfs

package obfs_test

import (
	"github.com/real-uangi/allingo/security/obfs"
	"math/rand"
	"testing"
)

const testSecret int64 = 0x028bce36e931eeb9
const testSecret32 int32 = 0x3d7931ee

func TestInt64(t *testing.T) {
	originInt := int64(100001)
	t.Log("originInt", originInt)

	encoded := obfs.ScrambleInt(originInt, testSecret)
	t.Log("encoded", encoded)

	decoded := obfs.DescrambleInt(encoded, testSecret)
	t.Log("decoded", decoded)

}

func TestInt32(t *testing.T) {
	origin := int32(100001)
	t.Log("origin", origin)
	encoded := obfs.ScrambleInt32(origin, testSecret32)
	t.Log("encoded", encoded)
	decoded := obfs.DescrambleInt32(encoded, testSecret32)
	t.Log("decoded", decoded)
}

func TestInt64Security(t *testing.T) {
	for i := 0; i < 100000000; i++ {
		originInt := rand.Int63()
		encoded := obfs.ScrambleInt(originInt, testSecret)
		decoded := obfs.DescrambleInt(encoded, testSecret)
		if decoded != originInt {
			t.Fail()
		}
	}
	for i := 0; i < 100000000; i++ {
		originInt := rand.Int31()
		encoded := obfs.ScrambleInt32(originInt, testSecret32)
		decoded := obfs.DescrambleInt32(encoded, testSecret32)
		if decoded != originInt {
			t.Fail()
		}
	}
}

func BenchmarkScrambleInt(b *testing.B) {
	originInt := rand.Int63()
	for i := 0; i < b.N; i++ {
		_ = obfs.ScrambleInt(originInt, testSecret)
	}
}

func BenchmarkDescrambleInt(b *testing.B) {
	originInt := rand.Int63()
	for i := 0; i < b.N; i++ {
		_ = obfs.DescrambleInt(originInt, testSecret)
	}
}

func BenchmarkScrambleInt32(b *testing.B) {
	originInt32 := rand.Int31()
	for i := 0; i < b.N; i++ {
		_ = obfs.ScrambleInt32(originInt32, testSecret32)
	}
}

func BenchmarkDescrambleInt32(b *testing.B) {
	originInt32 := rand.Int31()
	for i := 0; i < b.N; i++ {
		_ = obfs.DescrambleInt32(originInt32, testSecret32)
	}
}
