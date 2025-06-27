/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/17 14:55
 */

// Package convert

package convert

import (
	"bytes"
	"github.com/bytedance/sonic"
	"io"
)

var jsonApi = sonic.ConfigDefault

func ToJsonBuffer(v interface{}) (*bytes.Buffer, error) {
	bs, err := jsonApi.Marshal(v)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(bs), nil
}

func ParseJsonFromReader(r io.Reader, dst interface{}) error {
	dec := jsonApi.NewDecoder(r)
	err := dec.Decode(dst)
	if err != nil {
		return err
	}
	return nil
}

func Json() sonic.API {
	return jsonApi
}
