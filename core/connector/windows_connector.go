package connector

import (
	"fmt"

	"binrc.com/roma/core/model"
)

// WindowsConnector Windows 连接器
type WindowsConnector struct {
	Config *model.WindowsConfig
}

// NewWindowsConnector 创建 Windows 连接器
func NewWindowsConnector(config *model.WindowsConfig) *WindowsConnector {
	return &WindowsConnector{
		Config: config,
	}
}

// GetRDPInfo 获取 RDP 连接信息
func (w *WindowsConnector) GetRDPInfo() map[string]interface{} {
	host := w.Config.IPv4Pub
	if host == "" {
		host = w.Config.IPv4Priv
	}
	if host == "" {
		host = w.Config.IPv6
	}

	port := w.Config.Port
	if port == 0 {
		port = 3389 // 默认 RDP 端口
	}

	// 生成 RDP 文件内容
	rdpContent := fmt.Sprintf(`full address:s:%s:%d
username:s:%s
prompt for credentials:i:0
administrative session:i:1
screen mode id:i:2
use multimon:i:0
desktopwidth:i:1920
desktopheight:i:1080
session bpp:i:32
compression:i:1
keyboardhook:i:2
audiocapturemode:i:0
videoplaybackmode:i:1
connection type:i:7
networkautodetect:i:1
bandwidthautodetect:i:1
displayconnectionbar:i:1
enableworkspacereconnect:i:0
disable wallpaper:i:0
allow font smoothing:i:0
allow desktop composition:i:0
disable full window drag:i:1
disable menu anims:i:1
disable themes:i:0
disable cursor setting:i:0
bitmapcachepersistenable:i:1
audiomode:i:0
redirectprinters:i:1
redirectcomports:i:0
redirectsmartcards:i:1
redirectclipboard:i:1
redirectposdevices:i:0
autoreconnection enabled:i:1
authentication level:i:0
negotiate security layer:i:1`,
		host, port, w.Config.Username,
	)

	return map[string]interface{}{
		"type":        "windows_rdp",
		"hostname":    w.Config.Hostname,
		"host":        host,
		"port":        port,
		"username":    w.Config.Username,
		"rdp_file":    rdpContent,
		"description": w.Config.Description,
		"instructions": map[string]string{
			"windows": "将 RDP 文件保存为 .rdp 文件，然后双击打开",
			"macos":   "使用 Microsoft Remote Desktop 应用打开",
			"linux":   "使用 Remmina 或 rdesktop 连接",
		},
		"commands": map[string]string{
			"linux": fmt.Sprintf("rdesktop %s:%d -u %s", host, port, w.Config.Username),
			"macos": fmt.Sprintf("打开 Microsoft Remote Desktop，添加 %s:%d", host, port),
		},
	}
}

// GetSSHProxyInfo 获取 SSH 代理访问 RDP 的信息
func (w *WindowsConnector) GetSSHProxyInfo() map[string]interface{} {
	// 如果配置了 SSH，可以通过 SSH 隧道访问 RDP
	host := w.Config.IPv4Pub
	if host == "" {
		host = w.Config.IPv4Priv
	}

	sshPort := w.Config.Port
	if sshPort == 0 {
		sshPort = 22
	}

	rdpPort := w.Config.Port
	if rdpPort == 0 {
		rdpPort = 3389
	}

	// SSH 隧道命令
	sshTunnelCmd := fmt.Sprintf("ssh -L 3389:%s:%d %s@%s -p %d",
		host, rdpPort, w.Config.Username, host, sshPort)

	return map[string]interface{}{
		"type":            "ssh_tunnel",
		"ssh_tunnel_cmd":  sshTunnelCmd,
		"local_rdp_port":  3389,
		"instructions": []string{
			"1. 在本地终端执行 SSH 隧道命令",
			fmt.Sprintf("   %s", sshTunnelCmd),
			"2. 保持终端窗口打开",
			"3. 使用 RDP 客户端连接到 localhost:3389",
		},
		"rdp_connect": "localhost:3389",
	}
}

// GetPowerShellInfo 获取 PowerShell 远程信息
func (w *WindowsConnector) GetPowerShellInfo() map[string]interface{} {
	host := w.Config.IPv4Pub
	if host == "" {
		host = w.Config.IPv4Priv
	}

	return map[string]interface{}{
		"type":     "powershell_remoting",
		"hostname": w.Config.Hostname,
		"host":     host,
		"username": w.Config.Username,
		"commands": []string{
			"# PowerShell 远程连接",
			fmt.Sprintf("$cred = Get-Credential -UserName '%s' -Message 'Enter Password'", w.Config.Username),
			fmt.Sprintf("Enter-PSSession -ComputerName %s -Credential $cred", host),
			"",
			"# 或使用 WinRM",
			fmt.Sprintf("winrs -r:%s -u:%s cmd", host, w.Config.Username),
		},
		"requirements": []string{
			"Windows 服务器需要启用 WinRM 服务",
			"防火墙需要开放 5985 (HTTP) 或 5986 (HTTPS) 端口",
			"PowerShell Remoting 需要启用",
		},
	}
}

// GetConnectionInfo 获取所有连接信息
func (w *WindowsConnector) GetConnectionInfo() map[string]interface{} {
	return map[string]interface{}{
		"rdp_info":        w.GetRDPInfo(),
		"ssh_proxy_info":  w.GetSSHProxyInfo(),
		"powershell_info": w.GetPowerShellInfo(),
		"summary": fmt.Sprintf("Windows 服务器: %s (%s)",
			w.Config.Hostname,
			w.Config.IPv4Pub,
		),
	}
}

