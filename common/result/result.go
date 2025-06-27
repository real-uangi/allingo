/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/19 15:18
 */

// Package result
package result

import (
	"github.com/real-uangi/allingo/common/convert"
	"net/http"
	"time"
)

type RawField []byte

type Result[T any] struct {
	Code    int       `json:"code"`
	Data    T         `json:"data"`
	Message string    `json:"message"`
	Time    time.Time `json:"time"`
}

func New[T any](code int, data T) *Result[T] {
	return &Result[T]{
		Code:    code,
		Data:    data,
		Message: http.StatusText(code),
		Time:    time.Now(),
	}
}

func Custom[T any](code int, message string, data T) *Result[T] {
	return &Result[T]{
		Code:    code,
		Data:    data,
		Message: message,
		Time:    time.Now(),
	}
}

func Ok[T any](data T) *Result[T] {
	return New(http.StatusOK, data)
}

func NotFound() *Result[RawField] {
	return New[RawField](http.StatusNotFound, nil)
}

func BadRequest(err error) *Result[RawField] {
	if err == nil {
		return New[RawField](http.StatusBadRequest, nil)
	}
	return Custom[RawField](http.StatusBadRequest, err.Error(), nil)
}

func (result Result[T]) Render(w http.ResponseWriter) error {
	result.WriteContentType(w)
	jsonBytes, err := convert.Json().Marshal(result)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func (result Result[T]) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header.Get("Content-Type"); len(val) == 0 {
		header.Set("Content-Type", "application/json; charset=utf-8")
	}
}
