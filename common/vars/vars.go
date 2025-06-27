/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/3/21 09:31
 */

// Package vars

package vars

func Pointer[T any](v T) *T {
	return &v
}

func Value[T any](p *T) T {
	if p != nil {
		return *p
	}
	var t T
	return t
}
