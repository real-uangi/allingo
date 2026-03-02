/*
 * Copyright © 2024 Uangi. All rights reserved.
 * @author uangi
 * @date 2024/11/7 16:22
 */

// Package iptools
package iptools

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
)

const EnvPreferCidr = "ALLINGO_PREFER_CIDR"

// LocalAddress represents local IP addresses.
type LocalAddress struct {
	ips []net.IP
}

// GetLocalIP retrieves local IP addresses, returning an error if something goes wrong.
func GetLocalIP() (*LocalAddress, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("failed to get local IP addresses: %v", err)
	}
	result := make([]net.IP, 0, len(addresses))
	for _, address := range addresses {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			result = append(result, ipNet.IP)
		}
	}
	return &LocalAddress{
		ips: result,
	}, nil
}

// GetV4 returns a LocalAddress containing only IPv4 addresses.
func (la *LocalAddress) GetV4() *LocalAddress {
	result := make([]net.IP, 0, len(la.ips))
	for _, ip := range la.ips {
		if ip.To4() != nil {
			result = append(result, ip)
		}
	}
	return &LocalAddress{
		ips: result,
	}
}

// GetV6 returns a LocalAddress containing only IPv6 addresses.
func (la *LocalAddress) GetV6() *LocalAddress {
	result := make([]net.IP, 0, len(la.ips))
	for _, ip := range la.ips {
		if ip.To4() == nil {
			result = append(result, ip)
		}
	}
	return &LocalAddress{
		ips: result,
	}
}

// Strings returns the string representation of IP addresses.
func (la *LocalAddress) Strings() []string {
	result := make([]string, 0, len(la.ips))
	for _, v := range la.ips {
		result = append(result, v.String())
	}
	return result
}

// GetIPByCIDR returns the IP that matches the provided CIDR.
func (la *LocalAddress) GetIPByCIDR(cidr string) (net.IP, error) {
	_, network, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, fmt.Errorf("invalid CIDR format: %v", err)
	}

	for _, ip := range la.ips {
		if network.Contains(ip) {
			return ip, nil
		}
	}
	return nil, fmt.Errorf("no matching IP found for CIDR %s", cidr)
}

// GetBestMatchIp 获取最适合的IP 优先依据环境变量 ANTARCTICA_PREFER_CIDR
func GetBestMatchIp() (string, error) {
	la, err := GetLocalIP()
	if err != nil {
		return "", err
	}
	cidr, ok := os.LookupEnv(EnvPreferCidr)
	if ok {
		ip, err := la.GetIPByCIDR(cidr)
		if err != nil {
			return "", err
		}
		return ip.String(), nil
	}
	ss := la.GetV4().Strings()
	for _, ip := range ss {
		if !strings.HasPrefix(ip, "127.0.") && !strings.HasPrefix(ip, "0.") {
			return ip, nil
		}
	}
	return "", errors.New("can not find proper ip")
}
