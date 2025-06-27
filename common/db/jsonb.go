/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/12/24 14:09
 */

// Package db
package db

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/real-uangi/allingo/common/convert"
)

type JSONB[T any] struct {
	t T
}

func NewJSONB[T any](t T) *JSONB[T] {
	return &JSONB[T]{t: t}
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *JSONB[T]) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	return convert.Json().Unmarshal(bytes, &j.t)
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (j JSONB[T]) Value() (driver.Value, error) {
	return convert.Json().Marshal(j.t)
}

func (j *JSONB[T]) Set(value T) {
	j.t = value
}

func (j *JSONB[T]) Get() T {
	return j.t
}

func (JSONB[T]) GormDataType() string {
	return "jsonb"
}

// MarshalJSON returns m as the JSON encoding of m.
func (j JSONB[T]) MarshalJSON() ([]byte, error) {
	return convert.Json().Marshal(j.t)
}

// UnmarshalJSON sets *m to a copy of data.
func (j *JSONB[T]) UnmarshalJSON(data []byte) error {
	return convert.Json().Unmarshal(data, &j.t)
}
