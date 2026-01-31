/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/27 13:46
 */

// Package db
package db

import (
	"errors"
	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/db/helper/page"
	"github.com/real-uangi/allingo/common/log"
	"gorm.io/gorm"
)

type BaseMapper[T ModelConstraint] interface {
	SelectOne(t *T) (*T, error)
	Select(t *T) ([]T, error)
	SelectById(id uuid.UUID) (*T, error)

	GetPage(input page.InputInterface) (*page.Output[T], error)

	UpdateByPrimaryKeySelective(t *T) (int64, error)

	DeleteById(ids ...uuid.UUID) (int64, error)

	Insert(e *T) (int64, error)
	InsertBatch(list []T) (int64, error)

	Count(t *T) (int64, error)

	Transaction(f func(tx *gorm.DB) error) error

	GetConn() *gorm.DB
}

type BaseMapperImpl[T ModelConstraint] struct {
	emptyEntity  T
	emptyPointer *T
	conn         *gorm.DB
	logger       *log.StdLogger
}

func NewBaseMapper[T ModelConstraint](manager *Manager) BaseMapper[T] {
	m := new(BaseMapperImpl[T])
	var t T
	m.emptyEntity = t
	m.emptyPointer = &t
	m.conn = manager.GetDB()
	m.logger = log.For[BaseMapper[T]]()
	return m
}

func (m *BaseMapperImpl[T]) SelectOne(t *T) (*T, error) {
	result := m.conn.Where(t).First(t)
	if err := filterError(result.Error); err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return t, nil
}

func (m *BaseMapperImpl[T]) Select(t *T) ([]T, error) {
	list := make([]T, 0)
	result := m.conn.Where(t).Order("created_at").Find(&list)
	if err := filterError(result.Error); err != nil {
		return nil, err
	}
	return list, nil
}

func (m *BaseMapperImpl[T]) SelectById(id uuid.UUID) (*T, error) {
	t := new(T)
	result := m.conn.Where("id = ?", id).First(t)
	if err := filterError(result.Error); err != nil {
		return nil, err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return t, nil
}

func (m *BaseMapperImpl[T]) GetPage(input page.InputInterface) (*page.Output[T], error) {
	pageIndex := input.GetPageIndex()
	pageSize := input.GetPageSize()
	orderBy := input.GetOrderBy()
	if orderBy == "" {
		orderBy = "id"
	}
	input.ResetPageInfo()
	var total int64
	countResult := m.conn.Model(m.emptyEntity).Where(input).Order("created_at").Count(&total)
	if err := filterError(countResult.Error); err != nil {
		return nil, err
	}
	offset := pageSize * (pageIndex - 1)
	output := &page.Output[T]{
		PageIndex: pageIndex,
		PageSize:  pageSize,
		List:      nil,
		Total:     total,
	}
	more := int(total) - offset
	if more <= 0 {
		return output, nil
	}
	output.List = make([]T, 0)
	result := m.conn.Model(m.emptyEntity).Where(input).Limit(pageSize).Offset(offset).Order(orderBy).Find(&output.List)
	if err := filterError(result.Error); err != nil {
		return nil, err
	}
	return output, nil
}

func (m *BaseMapperImpl[T]) UpdateByPrimaryKeySelective(t *T) (int64, error) {
	result := m.conn.Model(t).Updates(t)
	if err := filterError(result.Error); err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func (m *BaseMapperImpl[T]) DeleteById(ids ...uuid.UUID) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := m.conn.Delete(m.emptyPointer, &ids)
	if err := filterError(result.Error); err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func (m *BaseMapperImpl[T]) Insert(e *T) (int64, error) {
	result := m.conn.Create(e)
	if err := filterError(result.Error); err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func (m *BaseMapperImpl[T]) InsertBatch(list []T) (int64, error) {
	l := len(list)
	if l == 0 {
		return 0, nil
	}
	if l == 1 {
		return m.Insert(&list[0])
	}
	result := m.conn.CreateInBatches(&list, 100)
	if err := filterError(result.Error); err != nil {
		return 0, err
	}
	return result.RowsAffected, nil
}

func (m *BaseMapperImpl[T]) Count(t *T) (int64, error) {
	var count int64
	result := m.conn.Model(m.emptyEntity).Where(t).Count(&count)
	if err := filterError(result.Error); err != nil {
		return 0, err
	}
	return count, nil
}

func (m *BaseMapperImpl[T]) Transaction(f func(tx *gorm.DB) error) error {
	panicked := true
	var err error
	tx := m.conn.Begin()
	if err := tx.Error; err != nil {
		return err
	}
	defer func() {
		if panicked || err != nil {
			m.logger.Warnf("rolling back transaction, panicked: %v, error: %v", panicked, err)
			tx.Rollback()
		}
	}()
	err = f(tx)
	if err != nil {
		return err
	}
	panicked = false
	if err := tx.Commit().Error; err != nil {
		return err
	}
	return nil
}

func (m *BaseMapperImpl[T]) GetConn() *gorm.DB {
	return m.conn
}

func filterError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}
