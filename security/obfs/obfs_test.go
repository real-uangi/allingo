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

const testSecret uint64 = 0xb28bce36e931eeb9

func TestInt64(t *testing.T) {
	originInt := rand.Int63()
	t.Log("originInt", originInt)

	encoded := obfs.ScrambleInt(uint64(originInt), testSecret)
	t.Log("encoded", encoded)

	decoded := obfs.DescrambleInt(encoded, testSecret)
	t.Log("decoded", decoded)

}

func TestInt64Security(t *testing.T) {
	for i := 0; i < 10000000; i++ {
		originInt := uint64(rand.Int63())
		encoded := obfs.ScrambleInt(originInt, testSecret)
		decoded := obfs.DescrambleInt(encoded, testSecret)
		if decoded != originInt {
			t.Fail()
		}
	}
}

func BenchmarkScrambleInt(b *testing.B) {
	originInt := uint64(rand.Int63())
	for i := 0; i < b.N; i++ {
		_ = obfs.ScrambleInt(originInt, testSecret)
	}
}

func BenchmarkDescrambleInt(b *testing.B) {
	originInt := uint64(rand.Int63())
	for i := 0; i < b.N; i++ {
		_ = obfs.DescrambleInt(originInt, testSecret)
	}
}
