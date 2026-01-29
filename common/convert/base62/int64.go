/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 10:59
 */

// Package base62

package base62

import (
	"errors"
	"fmt"
	"math"
)

func EncodeInt64(n int64) string {
	if n == 0 {
		return "0"
	}

	var buf [11]byte // int64 最大 11 位 base62
	i := len(buf)

	for n > 0 {
		i--
		buf[i] = base62Alphabet[n%62]
		n /= 62
	}

	return string(buf[i:])
}

func DecodeInt64(s string) (int64, error) {
	var n int64

	for i := 0; i < len(s); i++ {
		c := s[i]

		var v int64
		switch {
		case c >= '0' && c <= '9':
			v = int64(c - '0')
		case c >= 'a' && c <= 'z':
			v = int64(c-'a') + 10
		case c >= 'A' && c <= 'Z':
			v = int64(c-'A') + 36
		default:
			return 0, fmt.Errorf("invalid base62 char: %c", c)
		}

		if n > (math.MaxInt64-v)/62 {
			return 0, errors.New("int64 overflow")
		}

		n = n*62 + v
	}

	return n, nil
}
