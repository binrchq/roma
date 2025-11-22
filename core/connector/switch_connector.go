package connector

import (
	"fmt"
	"strings"

	"binrc.com/roma/core/model"
	gossh "golang.org/x/crypto/ssh"
)

// SwitchConnector 交换机连接器
type SwitchConnector struct {
	Config    *model.SwitchConfig
	SSHClient *gossh.Client
}

// NewSwitchConnector 创建交换机连接器
func NewSwitchConnector(config *model.SwitchConfig) *SwitchConnector {
	return &SwitchConnector{
		Config: config,
	}
}

// ConnectSSH 连接到交换机 SSH
func (s *SwitchConnector) ConnectSSH() error {
	host := s.Config.IPv4Pub
	if host == "" {
		host = s.Config.IPv4Priv
	}
	if host == "" {
		host = s.Config.IPv6
	}

	port := s.Config.Port
	if port == 0 {
		port = 22
	}

	config := &gossh.ClientConfig{
		User: s.Config.Username,
		Auth: []gossh.AuthMethod{},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	// 交换机通常只使用密码认证
	if s.Config.Password != "" {
		config.Auth = append(config.Auth, gossh.Password(s.Config.Password))
	}

	client, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return fmt.Errorf("SSH 连接失败: %v", err)
	}

	s.SSHClient = client
	return nil
}

// ExecuteCommand 执行交换机命令
func (s *SwitchConnector) ExecuteCommand(command string) (string, error) {
	if s.SSHClient == nil {
		if err := s.ConnectSSH(); err != nil {
			return "", err
		}
	}

	session, err := s.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH session 失败: %v", err)
	}
	defer session.Close()

	// 为交换机配置伪终端（某些交换机需要）
	modes := gossh.TerminalModes{
		gossh.ECHO:          0,     // 禁用回显
		gossh.TTY_OP_ISPEED: 14400, // 输入速度
		gossh.TTY_OP_OSPEED: 14400, // 输出速度
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		// 如果请求 PTY 失败，继续尝试执行命令
	}

	output, err := session.CombinedOutput(command)
	if err != nil {
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

// ShowVersion 显示交换机版本信息
func (s *SwitchConnector) ShowVersion() (string, error) {
	return s.ExecuteCommand("show version")
}

// ShowInterfaces 显示接口状态
func (s *SwitchConnector) ShowInterfaces() (string, error) {
	return s.ExecuteCommand("show interfaces status")
}

// ShowVLAN 显示 VLAN 配置
func (s *SwitchConnector) ShowVLAN() (string, error) {
	return s.ExecuteCommand("show vlan brief")
}

// ShowMAC 显示 MAC 地址表
func (s *SwitchConnector) ShowMAC() (string, error) {
	return s.ExecuteCommand("show mac address-table")
}

// ShowRunningConfig 显示运行配置
func (s *SwitchConnector) ShowRunningConfig() (string, error) {
	return s.ExecuteCommand("show running-config")
}

// ShowLog 显示日志
func (s *SwitchConnector) ShowLog() (string, error) {
	return s.ExecuteCommand("show log")
}

// GetSystemInfo 获取交换机系统信息
func (s *SwitchConnector) GetSystemInfo() (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// 版本信息
	version, err := s.ShowVersion()
	if err == nil {
		info["version"] = version
	}

	// 接口状态
	interfaces, err := s.ShowInterfaces()
	if err == nil {
		info["interfaces"] = interfaces
	}

	// VLAN 信息
	vlan, err := s.ShowVLAN()
	if err == nil {
		info["vlan"] = vlan
	}

	return info, nil
}

// ConfigureInterface 配置接口
func (s *SwitchConnector) ConfigureInterface(interfaceName, config string) (string, error) {
	commands := fmt.Sprintf(`configure terminal
interface %s
%s
end
write memory`, interfaceName, config)

	return s.ExecuteCommand(commands)
}

// ConfigureVLAN 配置 VLAN
func (s *SwitchConnector) ConfigureVLAN(vlanID int, name string) (string, error) {
	commands := fmt.Sprintf(`configure terminal
vlan %d
name %s
end
write memory`, vlanID, name)

	return s.ExecuteCommand(commands)
}

// SaveConfig 保存配置
func (s *SwitchConnector) SaveConfig() (string, error) {
	return s.ExecuteCommand("write memory")
}

// RebootSwitch 重启交换机
func (s *SwitchConnector) RebootSwitch() (string, error) {
	return s.ExecuteCommand("reload")
}

// GetConnectionInfo 获取连接信息
func (s *SwitchConnector) GetConnectionInfo() map[string]interface{} {
	host := s.Config.IPv4Pub
	if host == "" {
		host = s.Config.IPv4Priv
	}
	if host == "" {
		host = s.Config.IPv6
	}

	port := s.Config.Port
	if port == 0 {
		port = 22
	}

	// 常用命令（Cisco IOS 风格）
	ciscoCommands := []string{
		"# 基础查看命令",
		"show version                    # 版本信息",
		"show running-config             # 运行配置",
		"show startup-config             # 启动配置",
		"show interfaces status          # 接口状态",
		"show vlan brief                 # VLAN 信息",
		"show mac address-table          # MAC 地址表",
		"show ip interface brief         # IP 接口简要信息",
		"show arp                        # ARP 表",
		"show spanning-tree              # STP 状态",
		"show log                        # 日志",
		"",
		"# 配置命令",
		"configure terminal              # 进入配置模式",
		"interface GigabitEthernet0/1    # 进入接口配置",
		"vlan 10                         # 创建 VLAN",
		"write memory                    # 保存配置",
		"",
		"# 常用操作",
		"ping <ip>                       # Ping 测试",
		"traceroute <ip>                 # 路由追踪",
		"reload                          # 重启交换机",
	}

	// H3C 命令风格
	h3cCommands := []string{
		"display version                 # 版本信息",
		"display current-configuration   # 当前配置",
		"display interface brief         # 接口简要信息",
		"display vlan all                # VLAN 信息",
		"display mac-address             # MAC 地址表",
		"save                            # 保存配置",
		"reboot                          # 重启",
	}

	// 华为命令风格
	huaweiCommands := []string{
		"display version                 # 版本信息",
		"display current-configuration   # 当前配置",
		"display interface brief         # 接口信息",
		"display vlan                    # VLAN 信息",
		"display mac-address             # MAC 地址表",
		"save                            # 保存配置",
		"reboot                          # 重启",
	}

	return map[string]interface{}{
		"type":        "switch",
		"name":        s.Config.SwitchName,
		"host":        host,
		"port":        port,
		"username":    s.Config.Username,
		"ssh_command": fmt.Sprintf("ssh %s@%s -p %d", s.Config.Username, host, port),
		"commands": map[string]interface{}{
			"cisco":  strings.Join(ciscoCommands, "\n"),
			"h3c":    strings.Join(h3cCommands, "\n"),
			"huawei": strings.Join(huaweiCommands, "\n"),
		},
		"tips": []string{
			"1. 不同品牌交换机的命令格式可能不同",
			"2. Cisco 使用 'show' 命令",
			"3. H3C/华为使用 'display' 命令",
			"4. 配置更改后记得保存 (write memory 或 save)",
			"5. 重要操作前建议先备份配置",
		},
		"description": s.Config.Description,
	}
}

// Close 关闭连接
func (s *SwitchConnector) Close() error {
	if s.SSHClient != nil {
		return s.SSHClient.Close()
	}
	return nil
}

