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

// IsPrivateIP 检查IP是否为内网IP
// 输入: ip - IP地址字符串
// 输出: bool - 是否为内网IP
// 必要性: 安全策略中需要排除内网IP，避免误封禁内网地址
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	// 使用Go 1.17+的IsPrivate方法（如果可用）
	// 为了兼容性，也手动实现检查
	if parsedIP.IsLoopback() || parsedIP.IsLinkLocalUnicast() || parsedIP.IsLinkLocalMulticast() {
		return true
	}

	// 检查IPv4私有地址范围
	if ipv4 := parsedIP.To4(); ipv4 != nil {
		// 10.0.0.0/8
		if ipv4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ipv4[0] == 192 && ipv4[1] == 168 {
			return true
		}
		// 127.0.0.0/8 (localhost)
		if ipv4[0] == 127 {
			return true
		}
		// 169.254.0.0/16 (link-local)
		if ipv4[0] == 169 && ipv4[1] == 254 {
			return true
		}
		return false
	}

	// 检查IPv6私有地址范围
	if ipv6 := parsedIP.To16(); ipv6 != nil {
		// fc00::/7 (unique local address)
		if ipv6[0] == 0xfc || ipv6[0] == 0xfd {
			return true
		}
		// fe80::/10 (link-local)
		if ipv6[0] == 0xfe && (ipv6[1]&0xc0) == 0x80 {
			return true
		}
	}

	return false
}
