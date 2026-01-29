/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 10:55
 */

// Package base62

package base62

import (
	"errors"
	"math/big"
	"strings"
)

var base = big.NewInt(62)

func EncodeBigInt(n *big.Int) string {
	if n.Sign() == 0 {
		return "0"
	}

	num := new(big.Int).Set(n)
	mod := new(big.Int)

	var buf []byte

	for num.Sign() > 0 {
		num.DivMod(num, base, mod)
		buf = append(buf, base62Alphabet[mod.Int64()])
	}

	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf)
}

func DecodeBigInt(s string) (*big.Int, error) {
	n := big.NewInt(0)

	for _, c := range s {
		idx := int64(strings.IndexRune(base62Alphabet, c))
		if idx < 0 {
			return nil, errors.New("invalid base62 char")
		}

		n.Mul(n, base)
		n.Add(n, big.NewInt(idx))
	}

	return n, nil
}
