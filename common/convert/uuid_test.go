/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/16 16:20
 */

// Package convert

package convert_test

import (
	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/convert"
	"testing"
)

func TestUUIDToBase64(t *testing.T) {
	id := uuid.NewString()
	t.Log(id)
	encoded := convert.UUIDToBase64(uuid.MustParse(id))
	t.Log(encoded)
	decoded, err := convert.UUIDFromBase64(encoded)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(decoded)
	if decoded.String() != id {
		t.Fatal("not equal")
	}
}
