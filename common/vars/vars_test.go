/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/3/21 09:33
 */

// Package vars

package vars

import "testing"

func BenchmarkGeneric(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Pointer("xxx")
		_ = Pointer(123)
		_ = Pointer(1.23)
	}
}

func TestPointer(t *testing.T) {
	t.Log(Pointer(1.23))
	t.Log(Pointer(1.23))
	t.Log(*Pointer(1.23))
	t.Log(Pointer("xxx"))
	t.Log(Pointer("xxx"))
	t.Log(*Pointer("xxx"))
}

func TestValue2(t *testing.T) {
	var a *int64
	var b *float64
	var c *string
	var d *bool
	t.Log(Value(a), Value(b), Value(c), Value(d))
	a = Pointer(int64(1))
	b = Pointer(2.0)
	c = Pointer("hello")
	d = Pointer(true)
	t.Log(Value(a), Value(b), Value(c), Value(d))
}
