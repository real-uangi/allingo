/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/8/29 13:45
 */

// Package business

package business_test

import (
	"github.com/real-uangi/allingo/common/business"
	"net/http"
	"testing"
)

func genError() error {
	return business.NewError("111")
}

func TestNewError(t *testing.T) {

	t.Log(genError().Error())

	t.Log(business.NewError("111"))

	t.Log(business.ErrUnauthorized)

	t.Log(business.NewErrorWithStatus(http.StatusUnauthorized))

}
