/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 10:59
 */

// Package base62

package base62_test

import (
	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/convert/base62"
	"math/rand"
	"testing"
)

var (
	originInt64 = rand.Int63()
	originUUID  = uuid.New()
)

func TestBase62(t *testing.T) {

	t.Log("originInt64:", originInt64)
	t.Log("originUUID:", originUUID.String())
	t.Log("=====================================")

	base62Int64 := base62.EncodeInt64(originInt64)
	t.Log("base62Int64:", base62Int64)
	decodeInt64, err := base62.DecodeInt64(base62Int64)
	if err != nil {
		t.Error(err)
	}
	t.Log("decodeInt64:", decodeInt64)

	t.Log("=====================================")

	base62UUID := base62.EncodeUUID(originUUID)
	t.Log("base62UUID:", base62UUID)
	decodeUUID, err := base62.DecodeUUID(base62UUID)
	if err != nil {
		t.Error(err)
	}
	t.Log("decodeUUID:", decodeUUID.String())

}

func BenchmarkEncodeUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = base62.EncodeUUID(originUUID)
	}
}

func BenchmarkDecodeUUID(b *testing.B) {
	encoded := base62.EncodeUUID(originUUID)
	for i := 0; i < b.N; i++ {
		_, _ = base62.DecodeUUID(encoded)
	}
}

func BenchmarkEncodeInt64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = base62.EncodeInt64(originInt64)
	}
}

func BenchmarkDecodeInt64(b *testing.B) {
	encoded := base62.EncodeInt64(originInt64)
	for i := 0; i < b.N; i++ {
		_, _ = base62.DecodeInt64(encoded)
	}
}
