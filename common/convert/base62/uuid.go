/*
 * Copyright 2026 Uangi. All rights reserved.
 * @author uangi
 * @date 2026/1/29 10:56
 */

// Package base62

package base62

import (
	"errors"
	"github.com/google/uuid"
	"math/big"
)

func EncodeUUID(u uuid.UUID) string {
	n := new(big.Int).SetBytes(u[:])
	return EncodeBigInt(n)
}

func DecodeUUID(s string) (uuid.UUID, error) {
	n, err := DecodeBigInt(s)
	if err != nil {
		return uuid.Nil, err
	}

	b := n.Bytes()

	// UUID 必须 16 bytes
	if len(b) > 16 {
		return uuid.Nil, errors.New("overflow uuid")
	}

	var buf [16]byte
	copy(buf[16-len(b):], b)

	return uuid.FromBytes(buf[:])
}
