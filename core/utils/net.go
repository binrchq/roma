package utils

import (
	"errors"
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

// ResolveHostName 用途: 将域名解析为IP地址，确保连接时可用
// 输入: host - 原始主机名（IP或域名）
// 输出: string - 解析后的IP地址（若原值已是IP则原样返回）；error - 解析失败的原因
// 必要性: 资源允许填写域名时需要显式解析，确保连接行为可控且可观测
func ResolveHostName(host string) (string, error) {
	cleanHost := strings.TrimSpace(host)
	if cleanHost == "" {
		return "", errors.New("host is empty")
	}

	if net.ParseIP(cleanHost) != nil {
		return cleanHost, nil
	}

	ips, err := net.LookupHost(cleanHost)
	if err != nil {
		return cleanHost, err
	}
	if len(ips) == 0 {
		return cleanHost, errors.New("no ip records found")
	}
	return ips[0], nil
}
