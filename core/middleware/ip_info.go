package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IPInfo IP信息结构（对应 ipseek.cc API 返回格式）
type IPInfo struct {
	IP         string  `json:"ip"`
	ASN        string  `json:"asn"`
	ISP        string  `json:"isp"`
	Continent  string  `json:"continent"`
	Country    string  `json:"country"`
	Province   string  `json:"province"`
	City       string  `json:"city"`
	Area       string  `json:"area"`
	Additional string  `json:"additional"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	PostalCode string  `json:"postal_code"`
	CIDR       string  `json:"cidr"`
}

// GetIPInfo 从 ipseek.cc API 获取IP信息
// 输入: ip - IP地址
// 输出: string - IP信息（JSON格式）；error - 错误信息
// 必要性: 获取IP的地理位置和ISP信息，用于黑名单管理
func GetIPInfo(ip string) (string, error) {
	if ip == "" {
		return "", fmt.Errorf("IP address is empty")
	}

	// 调用 ipseek.cc API（根据用户示例，直接访问根路径会返回当前IP，需要传递ip参数）
	// 但根据curl示例，直接访问 https://ipseek.cc 会返回当前IP信息
	// 如果要查询指定IP，使用 ?ip=xxx 参数
	url := fmt.Sprintf("https://ipseek.cc/?ip=%s", ip)
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch IP info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 解析JSON以验证格式
	var ipInfo IPInfo
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return "", fmt.Errorf("failed to parse IP info: %w", err)
	}

	// 返回JSON字符串
	return string(body), nil
}

// GetIPInfoParsed 获取IP信息并解析为结构体
// 输入: ip - IP地址
// 输出: *IPInfo - IP信息结构体；error - 错误信息
// 必要性: 获取结构化的IP信息，便于前端显示
func GetIPInfoParsed(ip string) (*IPInfo, error) {
	jsonStr, err := GetIPInfo(ip)
	if err != nil {
		return nil, err
	}

	var ipInfo IPInfo
	if err := json.Unmarshal([]byte(jsonStr), &ipInfo); err != nil {
		return nil, err
	}

	return &ipInfo, nil
}
