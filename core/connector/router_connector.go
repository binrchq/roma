package connector

import (
	"fmt"
	"net/http"
	"strings"

	"binrc.com/roma/core/model"
	gossh "golang.org/x/crypto/ssh"
)

// RouterConnector 路由器连接器
type RouterConnector struct {
	Config    *model.RouterConfig
	SSHClient *gossh.Client
}

// NewRouterConnector 创建路由器连接器
func NewRouterConnector(config *model.RouterConfig) *RouterConnector {
	return &RouterConnector{
		Config: config,
	}
}

// GetWebInfo 获取 Web 管理界面信息
func (r *RouterConnector) GetWebInfo() map[string]interface{} {
	host := r.Config.IPv4Pub
	if host == "" {
		host = r.Config.IPv4Priv
	}
	if host == "" {
		host = r.Config.IPv6
	}

	webPort := r.Config.WebPort
	if webPort == 0 {
		webPort = 80 // 默认 HTTP 端口
	}

	// 尝试检测是否支持 HTTPS
	protocol := "http"
	if webPort == 443 {
		protocol = "https"
	}

	webURL := fmt.Sprintf("%s://%s:%d", protocol, host, webPort)

	return map[string]interface{}{
		"type":     "router_web",
		"name":     r.Config.RouterName,
		"url":      webURL,
		"host":     host,
		"port":     webPort,
		"username": r.Config.WebUsername,
		"instructions": []string{
			fmt.Sprintf("1. 浏览器访问: %s", webURL),
			fmt.Sprintf("2. 用户名: %s", r.Config.WebUsername),
			"3. 输入密码登录",
			"4. 进入管理界面进行配置",
		},
		"common_pages": map[string]string{
			"dashboard": webURL + "/",
			"status":    webURL + "/status",
			"network":   webURL + "/network",
			"firewall":  webURL + "/firewall",
			"system":    webURL + "/system",
		},
	}
}

// CheckWebAccess 检查 Web 管理界面是否可访问
func (r *RouterConnector) CheckWebAccess() (bool, string) {
	webInfo := r.GetWebInfo()
	url := webInfo["url"].(string)

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Sprintf("无法访问 Web 界面: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 401 {
		return true, "Web 界面可访问"
	}

	return false, fmt.Sprintf("Web 界面返回状态码: %d", resp.StatusCode)
}

// ConnectSSH 连接到路由器 SSH
func (r *RouterConnector) ConnectSSH() error {
	host := r.Config.IPv4Pub
	if host == "" {
		host = r.Config.IPv4Priv
	}
	if host == "" {
		host = r.Config.IPv6
	}

	port := r.Config.Port
	if port == 0 {
		port = 22
	}

	config := &gossh.ClientConfig{
		User: r.Config.Username,
		Auth: []gossh.AuthMethod{},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	if r.Config.PrivateKey != "" {
		signer, err := gossh.ParsePrivateKey([]byte(r.Config.PrivateKey))
		if err == nil {
			config.Auth = append(config.Auth, gossh.PublicKeys(signer))
		}
	}

	if r.Config.Password != "" {
		config.Auth = append(config.Auth, gossh.Password(r.Config.Password))
	}

	client, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %v", err)
	}

	r.SSHClient = client
	return nil
}

// ExecuteCommand 执行路由器命令
func (r *RouterConnector) ExecuteCommand(command string) (string, error) {
	if r.SSHClient == nil {
		if err := r.ConnectSSH(); err != nil {
			return "", err
		}
	}

	session, err := r.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH session 失败: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

// GetSystemInfo 获取路由器系统信息
func (r *RouterConnector) GetSystemInfo() (string, error) {
	commands := []string{
		"cat /proc/cpuinfo | grep 'model name' | head -1",
		"free -h",
		"uptime",
		"uname -a",
	}

	var output strings.Builder
	for _, cmd := range commands {
		result, err := r.ExecuteCommand(cmd)
		if err != nil {
			output.WriteString(fmt.Sprintf("[%s] 失败: %v\n", cmd, err))
		} else {
			output.WriteString(fmt.Sprintf("[%s]\n%s\n\n", cmd, result))
		}
	}

	return output.String(), nil
}

// GetNetworkInfo 获取网络配置信息
func (r *RouterConnector) GetNetworkInfo() (string, error) {
	commands := []string{
		"ip addr show",
		"ip route show",
		"iptables -L -n -v",
	}

	var output strings.Builder
	for _, cmd := range commands {
		result, err := r.ExecuteCommand(cmd)
		if err != nil {
			output.WriteString(fmt.Sprintf("[%s] 失败: %v\n", cmd, err))
		} else {
			output.WriteString(fmt.Sprintf("[%s]\n%s\n\n", cmd, result))
		}
	}

	return output.String(), nil
}

// GetConnectionInfo 获取连接信息
func (r *RouterConnector) GetConnectionInfo() map[string]interface{} {
	host := r.Config.IPv4Pub
	if host == "" {
		host = r.Config.IPv4Priv
	}

	commonCommands := []string{
		"show version            # 查看版本信息",
		"show running-config     # 查看运行配置",
		"show interfaces         # 查看接口状态",
		"show ip route           # 查看路由表",
		"show arp                # 查看 ARP 表",
		"show log                # 查看日志",
	}

	return map[string]interface{}{
		"type":            "router",
		"name":            r.Config.RouterName,
		"web_info":        r.GetWebInfo(),
		"ssh_host":        host,
		"ssh_port":        r.Config.Port,
		"ssh_username":    r.Config.Username,
		"ssh_command":     fmt.Sprintf("ssh %s@%s -p %d", r.Config.Username, host, r.Config.Port),
		"common_commands": strings.Join(commonCommands, "\n"),
		"description":     r.Config.Description,
	}
}

// Close 关闭连接
func (r *RouterConnector) Close() error {
	if r.SSHClient != nil {
		return r.SSHClient.Close()
	}
	return nil
}


