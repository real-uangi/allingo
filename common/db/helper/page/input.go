/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/5/21 19:08
 */

// Package page
package page

type InputInterface interface {
	GetPageIndex() int
	GetPageSize() int
	GetOffset() int
	ResetPageInfo()
}

type Input struct {
	PageIndex int `json:"pageIndex" form:"pageIndex"`
	PageSize  int `json:"pageSize" form:"pageSize"`
}

func (i *Input) GetPageIndex() int {
	if i.PageIndex == 0 {
		return 1
	}
	return i.PageIndex
}

func (i *Input) GetPageSize() int {
	if i.PageSize == 0 {
		return 10
	}
	return i.PageSize
}

func (i *Input) ResetPageInfo() {
	i.PageIndex = 0
	i.PageSize = 0
}

func (i *Input) GetOffset() int {
	return (i.GetPageIndex() - 1) * i.GetPageSize()
}
