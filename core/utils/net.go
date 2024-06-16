package utils

import (
	"net"
	"strings"
)

// 检查是否是合法的IP地址
func IsIP(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil
}

// 检查是否是合法的IP:port地址
func IsIPPort(str string) bool {
	parts := strings.Split(str, ":")
	if len(parts) != 2 {
		return false
	}
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return false
	}
	_, err := net.LookupPort("tcp", parts[1])

	return err == nil
}

// 检查是否是合法的域名
func IsDomain(hostname string) bool {
	// 使用net.ParseIP尝试解析，如果解析成功说明不是域名
	ip := net.ParseIP(hostname)
	if ip != nil {
		return false
	}

	// 使用net.LookupHost尝试解析，如果解析失败说明不是域名
	_, err := net.LookupHost(hostname)
	return err == nil
}

func IsDomainPort(hostname string) bool {
	// 使用net.ParseIP尝试解析，如果解析成功说明不是域名
	ip := net.ParseIP(hostname)
	if ip != nil {
		return false
	}

	// 使用net.LookupHost尝试解析，如果解析失败说明不是域名
	_, err := net.LookupHost(hostname)
	return err == nil
}
