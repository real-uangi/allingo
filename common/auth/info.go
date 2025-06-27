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
	PowerUser
	EliteUser

	Moderator UserLevel = 1111
	Admin     UserLevel = 9999
	System    UserLevel = 7777
)

type Info struct {
	Token     string    `json:"token"`
	UserName  string    `json:"userName"`
	UserLevel UserLevel `json:"userLevel"`
	UserId    string    `json:"userId"`
	Account   string    `json:"account"`
	Email     string    `json:"email"`
	Sex       int8      `json:"sex"`
}

func (info *Info) IsAdmin() bool {
	return info.UserLevel == Admin
}

// CurrentLevel 会检查有效期并返回当前实际等级
func (info *Info) CurrentLevel() UserLevel {
	return info.UserLevel
}

// CheckPermission 判断是否满足权限要求
func (info *Info) CheckPermission(permission UserLevel) bool {
	return permission <= info.CurrentLevel()
}

func CurrentUser() *Info {
	v, ok := holder.Get(constants.AuthInfoContext)
	if !ok {
		return nil
	}
	return v.(*Info)
}
