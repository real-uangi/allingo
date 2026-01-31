/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/21 11:24
 */

// Package db
package db

import (
	"github.com/google/uuid"
	"github.com/real-uangi/allingo/common/convert"
	"github.com/real-uangi/allingo/common/db/helper/page"
	"testing"
)

type AExample struct {
	ID   uuid.UUID       `json:"id" gorm:"primary_key"`
	Data JSONB[[]string] `json:"data" gorm:"column:data"`
	Model
}

type SearchExample struct {
	ID string `json:"id"`
	page.Input
}

var manager *Manager

var initErr error

func init() {
	var err error
	manager, err = InitDataSource()
	if err != nil {
		initErr = err
		return
	}
	err = manager.GetDB().AutoMigrate(AExample{})
	if err != nil {
		initErr = err
	}
}

func TestDB(t *testing.T) {
	if initErr != nil {
		t.Skip()
	}

	example := &AExample{}
	example.ID = uuid.New()

	result := manager.GetDB().Create(example)

	if result.Error != nil {
		t.Fatal(result.Error)
	}

	t.Log(example.ID)

	err := manager.GetDB().Delete(&AExample{ID: example.ID}).Error
	if err != nil {
		t.Fatal(err)
	}

}

func TestMapper(t *testing.T) {
	if initErr != nil {
		t.Skip()
	}

	mapper := NewBaseMapper[AExample](manager)

	spId := uuid.New()

	t.Log(mapper.Insert(&AExample{ID: spId}))

	t.Log(mapper.UpdateByPrimaryKeySelective(&AExample{ID: spId}))

	spId2 := uuid.New()
	spId3 := uuid.New()
	list := make([]AExample, 0)
	list = append(list, AExample{ID: spId2})
	list = append(list, AExample{ID: spId3})
	t.Log(mapper.InsertBatch(list))

	t.Log(mapper.SelectOne(&AExample{ID: spId}))
	t.Log(mapper.SelectById(spId))

	t.Log(mapper.Count(&AExample{}))

	t.Log(mapper.Select(&AExample{}))

	input := SearchExample{
		//ID: spId,
		Input: page.Input{
			PageIndex: 1,
			PageSize:  10,
		},
	}
	t.Log(mapper.GetPage(&input))

	t.Log(mapper.DeleteById(spId, spId2, spId3))
	t.Log(mapper.Select(&AExample{}))
}

func TestJSONB(t *testing.T) {
	if initErr != nil {
		t.Skip()
	}

	mapper := NewBaseMapper[AExample](manager)

	data := &AExample{ID: uuid.New()}
	data.Data.Set([]string{"111", "222"})
	t.Log(mapper.Insert(data))

	find, err := mapper.SelectById(data.ID)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(convert.Json().MarshalToString(find))

	t.Log(mapper.DeleteById(data.ID))
	t.Log(find.Data.Get())
}
