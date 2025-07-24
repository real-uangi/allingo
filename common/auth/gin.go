/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/7/24 14:26
 */

// Package auth

package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/real-uangi/allingo/common/result"
	"net"
	"net/http"
	"regexp"
	"strings"
)

var privateIPBlocks = []*net.IPNet{
	{IP: net.IP{10, 0, 0, 0}, Mask: net.CIDRMask(8, 32)},
	{IP: net.IP{172, 16, 0, 0}, Mask: net.CIDRMask(12, 32)},
	{IP: net.IP{192, 168, 0, 0}, Mask: net.CIDRMask(16, 32)},
	{IP: net.IP{127, 0, 0, 0}, Mask: net.CIDRMask(8, 32)},
	{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)},
}

func isInternalIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

// 提取 Host 中的 IP（去掉端口）
func extractIPFromHost(host string) string {
	// 处理 IPv6 格式：[::1]:8080
	if strings.HasPrefix(host, "[") {
		parts := strings.SplitN(host, "]", 2)
		if len(parts) > 0 {
			return strings.TrimPrefix(parts[0], "[")
		}
	}
	// IPv4:PORT
	parts := strings.Split(host, ":")
	return parts[0]
}

// 校验 host 是否是纯 IP 且为内网 IP
func isHostValid(host string) bool {
	ipStr := extractIPFromHost(host)

	// 简单校验是否为 IP 地址（防止 host 是域名）
	ipFormat := regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}$|^([0-9a-fA-F:]+)$`)
	if !ipFormat.MatchString(ipStr) {
		return false
	}

	return isInternalIP(ipStr)
}

func InternalOnlyMiddleware(c *gin.Context) {

	if !isInternalIP(c.ClientIP()) {
		c.AbortWithStatusJSON(http.StatusForbidden, result.New(http.StatusForbidden, "Access denied: internal IPs only"))
		return
	}

	if !isHostValid(c.Request.Host) {
		c.AbortWithStatusJSON(http.StatusForbidden, result.New(http.StatusForbidden, "Access denied: invalid host"))
		return
	}

	c.Next()
}
