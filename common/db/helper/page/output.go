/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/21 19:09
 */

// Package page
package page

type Output[Data any] struct {
	PageIndex int    `json:"pageIndex"`
	PageSize  int    `json:"pageSize"`
	List      []Data `json:"list"`
	Total     int64  `json:"total"`
}

func NewOutput[T any](arr []T, input InputInterface, total int64) *Output[T] {
	return &Output[T]{
		PageIndex: input.GetPageIndex(),
		PageSize:  input.GetPageSize(),
		List:      arr,
		Total:     total,
	}
}
