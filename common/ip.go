/*
 * Copyright 2025 Uangi. All rights reserved.
 * @author uangi
 * @date 2025/6/18 15:38
 */

// Package common

package common

import "net"

func IsIPv6(ip net.IP) bool {
	return ip != nil && ip.To4() == nil && len(ip) == net.IPv6len
}
