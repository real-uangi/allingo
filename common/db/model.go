/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/21 11:23
 */

// Package db
package db

import (
	"gorm.io/gorm"
	"time"
)

// Model is recommended basic fields for every table
type Model struct {
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt"`
}

type ModelConstraint interface {
	HasChanged() bool
	IsDel() bool
}

func (m Model) HasChanged() bool {
	return m.CreatedAt.Equal(m.UpdatedAt)
}

func (m Model) IsDel() bool {
	return m.DeletedAt.Valid
}

// HardDeleteModel 适合需要硬删除的场景
type HardDeleteModel struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m HardDeleteModel) HasChanged() bool {
	return m.CreatedAt.Equal(m.UpdatedAt)
}

func (m HardDeleteModel) IsDel() bool {
	return false
}
