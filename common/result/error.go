/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/27 11:14
 */

// Package result
package result

import (
	"github.com/real-uangi/allingo/common/business"
	"net/http"
)

func FromError(err error) (int, *Result[RawField]) {
	be, ok := err.(*business.RuntimeError)
	if ok {
		return http.StatusOK, Custom[RawField](be.GetStatusCode(), be.Error(), nil)
	}
	return http.StatusInternalServerError, Custom[RawField](http.StatusInternalServerError, err.Error(), nil)
}
