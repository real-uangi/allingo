/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 15:41
 */

// Package common

package common

import (
	"net"
	"testing"
)

func TestIsIPv6(t *testing.T) {
	ips := []string{
		"192.168.0.1",        // IPv4
		"::1",                // IPv6 loopback
		"2001:db8::1",        // IPv6
		"::ffff:192.168.0.1", // IPv4-mapped IPv6
	}
	for _, ip := range ips {
		t.Log(ip, IsIPv6(net.ParseIP(ip)))
	}
}
