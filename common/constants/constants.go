/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 17:21
 */

// Package constants
package constants

const (
	TraceIdKey    = "TRACE_ID"
	TraceIdHeader = "TRACE-ID"
)

const (
	DefaultProfile   = "dev"
	ProfileActiveKey = "PROFILE_ACTIVE"
)

const (
	DefaultGrayScale = "v1"
	GrayScaleKey     = "GRAY_SCALE"
	GrayScaleHeader  = "GRAY-SCALE"
)

const (
	AuthInfoContext = "system_auth_info"

	AuthHeader     = "Authorization"
	AUthInfoHeader = "Authorization-Body"

	BearerPrefix       = "Bearer "
	BearerPrefixLength = len(BearerPrefix)
	AuthInfoRedisKey   = "auth:"

	LastCheckAuthHeader = "Last-Check-Auth"

	ClientIpContext = "client_ip"
)

const (
	GinContextInjectionKey = "gin_context_injection"
	RequestStartTimeKey    = "request_start_time"
)

const (
	CFConnectingIpHeader = "CF-Connecting-IP"
	CFIPCountryHeader    = "CF-IPCountry"

	UserAgentHeader = "User-Agent"
)
