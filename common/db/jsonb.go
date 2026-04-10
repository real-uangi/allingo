/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/12/24 14:09
 */

// Package db
package db

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/real-uangi/allingo/common/convert"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type JSONB[T any] struct {
	t T
}

func NewJSONB[T any](t T) *JSONB[T] {
	return &JSONB[T]{t: t}
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *JSONB[T]) Scan(value interface{}) error {
	switch v := value.(type) {
	case nil:
		var zero T
		j.t = zero
		return nil
	case []byte:
		return convert.Json().Unmarshal(v, &j.t)
	case string:
		return convert.Json().Unmarshal([]byte(v), &j.t)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
}

// Value 实现 driver.Valuer 接口，返回 JSON 文本而不是 []byte。
// pgx simple protocol 会按 Go 类型做参数插值，[]byte 会被当成 bytea，
// 导致 jsonb 列写入失败或行为异常。
func (j JSONB[T]) Value() (driver.Value, error) {
	return convert.Json().MarshalToString(j.t)
}

// GormValue 为 GORM 生成显式 jsonb cast，避免在缺少列类型上下文时被错误推断。
func (j JSONB[T]) GormValue(_ context.Context, _ *gorm.DB) clause.Expr {
	value, err := j.Value()
	if err != nil {
		return clause.Expr{SQL: "NULL"}
	}
	return clause.Expr{
		SQL:  "?::jsonb",
		Vars: []interface{}{value},
	}
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
