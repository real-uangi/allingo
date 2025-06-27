/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 16:20
 */

// Package convert

package convert

import (
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
)

func UUIDToBase64(u uuid.UUID) string {
	// 使用 base64 URL 编码并去除填充符号 =
	return base64.RawURLEncoding.EncodeToString(u[:])
}

func UUIDFromBase64(s string) (uuid.UUID, error) {
	bytes, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return uuid.Nil, err
	}
	if len(bytes) != 16 {
		return uuid.Nil, fmt.Errorf("invalid UUID byte length: %d", len(bytes))
	}
	return uuid.FromBytes(bytes)
}
