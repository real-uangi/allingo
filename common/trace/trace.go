/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 17:28
 */

// Package trace
package trace

import (
	"github.com/real-uangi/allingo/common/constants"
	"github.com/real-uangi/allingo/common/holder"
	"github.com/real-uangi/allingo/common/random"
)

func GetSpecific(gid int64) string {
	v, ok := holder.GetSpecific(constants.TraceIdKey, gid)
	if ok {
		return v.(string)
	}
	return random.String(16)
}

func Get() string {
	v, ok := holder.Get(constants.TraceIdKey)
	if !ok {
		return ""
	}
	return v.(string)
}
