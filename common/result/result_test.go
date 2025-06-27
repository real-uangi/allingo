/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/27 11:20
 */

// Package result
package result

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/real-uangi/allingo/common/business"
	"testing"
)

func TestFromError(t *testing.T) {
	t.Log(marshalFromErr(FromError(business.ErrNotFound)))
	t.Log(marshalFromErr(FromError(business.ErrTooManyRequests)))
	t.Log(marshalFromErr(FromError(business.ErrForbidden)))

	t.Log(marshalFromErr(FromError(errors.New("this should be 500"))))

	t.Log(marshalStr(BadRequest(errors.New("this should be 500"))))
	t.Log(marshalStr(BadRequest(nil)))

}

func marshalStr(v any) (string, error) {
	bs, err := json.Marshal(v)
	return string(bs), err
}

func marshalFromErr(c int, v any) (string, error) {
	bs, err := json.Marshal(v)
	return fmt.Sprintf("%d - %s", c, string(bs)), err
}
