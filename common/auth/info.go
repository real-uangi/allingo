/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/19 16:14
 */

// Package auth
package auth

import (
	"github.com/real-uangi/allingo/common/constants"
	"github.com/real-uangi/allingo/common/holder"
)

type UserLevel int

const (
	Anonymous UserLevel = iota

	Peasant

	User

	Vip

	Moderator UserLevel = 1111
	Admin     UserLevel = 9999
	System    UserLevel = 7777
)

type LevelInfo struct {
	Level UserLevel `json:"level"`
}

type LevelInfoInterface interface {
	IsAdmin() bool
	CurrentLevel() UserLevel
	CheckPermission(permission UserLevel) bool
}

func (info *LevelInfo) IsAdmin() bool {
	return info.Level == Admin
}

// CurrentLevel 会检查有效期并返回当前实际等级
func (info *LevelInfo) CurrentLevel() UserLevel {
	return info.Level
}

// CheckPermission 判断是否满足权限要求
func (info *LevelInfo) CheckPermission(permission UserLevel) bool {
	return permission <= info.CurrentLevel()
}

func CurrentUser() LevelInfoInterface {
	v, ok := holder.Get(constants.AuthInfoContext)
	if !ok {
		return nil
	}
	return v.(LevelInfoInterface)
}
