/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 09:27
 */

// Package argon

package argon

import (
	"fmt"
	"testing"
)

func TestArgon2(t *testing.T) {
	password := "secret-password"
	hashed, err := GenerateFromPassword(password, DefaultParams)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Encoded:", hashed)

	ok, err := ComparePassword("secret-password", hashed)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Match:", ok)
}

var password = "secret-password"
var hashed string

func init() {
	var err error
	hashed, err = GenerateFromPassword(password, DefaultParams)
	if err != nil {
		panic(err)
	}
	fmt.Println(DefaultParams)
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateFromPassword(password, DefaultParams)
	}
}

func BenchmarkCompare(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = ComparePassword(password, hashed)
	}
}
