/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/25 17:17
 */

// Package api
package api

import (
	"github.com/gin-gonic/gin"
	"github.com/real-uangi/allingo/common/constants"
	"github.com/real-uangi/allingo/common/holder"
	"os"
	"strconv"
)

var CookieDomain = os.Getenv("COOKIE_DOMAIN")
var SecureCookie = os.Getenv("PROFILE_ACTIVE") != "dev"

func GinContext() *gin.Context {
	p, ok := holder.Get(constants.GinContextInjectionKey)
	if !ok {
		return nil
	}
	return p.(*gin.Context)
}

func GetRemoteIP() string {
	return GinContext().GetHeader(constants.CFConnectingIpHeader)
}

func GetRemoteRegion() string {
	return GinContext().GetHeader(constants.CFIPCountryHeader)
}

func GetUserAgent() string {
	return GinContext().GetHeader(constants.UserAgentHeader)
}

func NoCache() {
	GinContext().Header("Cache-Control", "no-cache")
}

func DefaultCache() {
	GinContext().Header("Cache-Control", "public, max-age=86400")
}

func Cache(seconds int64) {
	GinContext().Header("Cache-Control", "public, max-age="+strconv.FormatInt(seconds, 10))
}

func SetCookie(name string, value string) {
	GinContext().SetCookie(name, value, 86400, "/", CookieDomain, SecureCookie, true)
}
