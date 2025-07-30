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

func TestSP(t *testing.T) {
	t.Log(convert.UUIDMustToBase64("90626524-9722-42f1-9752-4aab29155683"))
	t.Log(convert.UUIDMustToBase64("b87e3f1f-7aac-463d-a749-4379e78ada67"))
	t.Log(convert.UUIDMustToBase64("2e85dd5f-3fe3-4aa9-8675-e06a92ad254d"))
}
