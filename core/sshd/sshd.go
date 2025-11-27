package sshd

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"binrc.com/roma/core/utils/logger"
	"github.com/fatih/color"
	"github.com/loganchef/ssh"
	gossh "golang.org/x/crypto/ssh"
)

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewTerminal 创建交互式 SSH 终端
// 输入: sess - SSH 会话；ip - 目标 IP；port - 目标端口；sshUser - SSH 用户名；key - 私钥；resType - 资源类型；password - 密码（可选）
// 输出: error - 错误信息
// 必要性: 这是建立交互式 SSH 终端的核心函数，支持公钥和密码两种认证方式
func NewTerminal(sess *ssh.Session, ip string, port int, sshUser string, key string, resType string, password ...string) error {
	var pwd string
	if len(password) > 0 {
		pwd = password[0]
	}
	upstreamClient, err := NewSSHClient(ip, port, sshUser, key, resType, pwd)
	if err != nil {
		return err
	}

	upstreamSess, err := upstreamClient.NewSession()
	if err != nil {
		return err
	}
	defer upstreamSess.Close()

	upstreamSess.Stdout = *sess
	upstreamSess.Stdin = *sess
	upstreamSess.Stderr = *sess

	pty, winCh, _ := (*sess).Pty()

	// 设置终端模式，确保输入输出正常显示
	termModes := gossh.TerminalModes{
		gossh.ECHO:          1,     // 启用回显
		gossh.TTY_OP_ISPEED: 14400, // 输入速度
		gossh.TTY_OP_OSPEED: 14400, // 输出速度
		gossh.IGNCR:         0,     // 不忽略 CR
		gossh.ICRNL:         1,     // 将 CR 转换为 NL
		gossh.ONLCR:         1,     // 将 NL 映射为 CR-NL
	}

	// 确保 TERM 环境变量被设置
	// 在 Docker 容器中，如果客户端没有请求 PTY，pty.Term 可能为空，使用默认值
	term := pty.Term
	if term == "" {
		term = "xterm-256color"
	}

	// 获取窗口大小，如果未设置则使用默认值
	height := pty.Window.Height
	width := pty.Window.Width
	if height <= 0 {
		height = 24
	}
	if width <= 0 {
		width = 80
	}

	// 设置 TERM 环境变量
	if err := upstreamSess.Setenv("TERM", term); err != nil {
		// 如果设置环境变量失败，记录但不中断（某些 SSH 服务器可能不支持 Setenv）
		logger.Logger.Warning(fmt.Sprintf("Failed to set TERM environment variable: %v", err))
	}

	// 请求 PTY，即使在 Docker 容器中也要请求，以确保终端正常工作
	if err := upstreamSess.RequestPty(term, height, width, termModes); err != nil {
		// 如果请求 PTY 失败，记录警告但继续（某些情况下可能不需要 PTY）
		logger.Logger.Warning(fmt.Sprintf("Failed to request PTY: %v, continuing without PTY", err))
	}

	if err := upstreamSess.Shell(); err != nil {
		return err
	}

	// 只有在有窗口变化通道时才处理窗口大小变化
	if winCh != nil {
		go func() {
			for win := range winCh {
				if err := upstreamSess.WindowChange(win.Height, win.Width); err != nil {
					logger.Logger.Warning(fmt.Sprintf("Failed to change window size: %v", err))
					break
				}
			}
		}()
	}

	if err := upstreamSess.Wait(); err != nil {
		return err
	}

	return nil
}

// NewSSHClient 创建 SSH 客户端连接
// 输入: ip - 目标 IP 地址；port - 目标端口；sshUser - SSH 用户名；key - 私钥内容（可为空）；resType - 资源类型；password - 密码（可为空）
// 输出: *gossh.Client - SSH 客户端；error - 错误信息
// 必要性: 这是建立 SSH 连接的核心函数，支持公钥和密码两种认证方式
func NewSSHClient(ip string, port int, sshUser string, key string, resType string, password ...string) (*gossh.Client, error) {
	var pwd string
	if len(password) > 0 {
		pwd = password[0]
	}

	// 认证优先级：
	// 1. 优先使用资源自身的密钥字段（通过 key 参数传入，来自资源的 PrivateKey 字段）
	// 2. 如果资源没有配置密钥，则从 passports 表查找该资源类型的默认密钥
	// 3. 如果都没有，且提供了密码，则使用密码认证

	// 记录密钥来源，便于调试
	keySource := "resource"
	if key == "" {
		keySource = "passport"
		op := operation.NewPassportOperation()
		keys, err := op.GetPassportByType(resType)
		if err != nil {
			logger.Logger.Error(fmt.Sprintf("Failed to get passport for resource type %s: %v", resType, err))
			// 如果 passports 表中也没有密钥，且没有密码，则返回错误
			if pwd == "" {
				return nil, fmt.Errorf("no key found (resource PrivateKey is empty, and no passport found for type %s) and no password provided", resType)
			}
		} else if len(keys) > 0 {
			key = keys[0].Passport
			if sshUser == "" {
				sshUser = keys[0].ServiceUser
			}
			logger.Logger.Info(fmt.Sprintf("Using passport key for resource type %s", resType))
		} else {
			keySource = "none"
			logger.Logger.Warning(fmt.Sprintf("No passport found for resource type %s", resType))
		}
	} else {
		logger.Logger.Debug(fmt.Sprintf("Using resource PrivateKey (length: %d)", len(key)))
	}

	// 构建认证方法列表
	authMethods := []gossh.AuthMethod{}

	// 优先使用私钥认证
	if key != "" {
		// 清理密钥字符串（去除前后空白）
		key = strings.TrimSpace(key)
		if key == "" {
			logger.Logger.Warning("Private key is empty after trimming whitespace")
		} else {
			// 处理转义的换行符：将字符串 "\n" 转换为实际的换行符
			// 数据库中可能存储的是转义的换行符（\n），需要转换为实际的换行符
			key = strings.ReplaceAll(key, "\\n", "\n")

			// 尝试解析私钥
			signer, err := gossh.ParsePrivateKey([]byte(key))
			if err != nil {
				logger.Logger.Error(fmt.Sprintf("Failed to parse private key from %s (key length: %d, first 50 chars: %s): %v", keySource, len(key), key[:min(len(key), 50)], err))
				// 如果密钥解析失败，记录详细错误，但继续尝试密码认证
			} else {
				authMethods = append(authMethods, gossh.PublicKeys(signer))
				logger.Logger.Debug(fmt.Sprintf("Successfully parsed private key from %s", keySource))
			}
		}
	}

	// 如果提供了密码，添加密码认证（作为备选）
	if pwd != "" {
		authMethods = append(authMethods, gossh.Password(pwd))
		logger.Logger.Debug("Added password authentication method")
	}

	// 如果没有任何认证方法，返回详细的错误信息
	if len(authMethods) == 0 {
		var keyInfo string
		if key != "" {
			keyInfo = fmt.Sprintf("key provided but failed to parse (length: %d)", len(key))
		} else {
			keyInfo = "no key provided"
		}
		var pwdInfo string
		if pwd != "" {
			pwdInfo = "password provided"
		} else {
			pwdInfo = "no password provided"
		}
		return nil, fmt.Errorf("no authentication method available (%s, %s)", keyInfo, pwdInfo)
	}

	// 设置用户名（如果未提供，使用默认值）
	if sshUser == "" {
		sshUser = "root"
	}

	configs := &gossh.ClientConfig{
		User:            sshUser,
		Auth:            authMethods,
		HostKeyCallback: gossh.HostKeyCallback(func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }),
	}

	dialHost := strings.TrimSpace(ip)
	if resolved, err := utils.ResolveHostName(dialHost); err != nil {
		logger.Logger.Warning(fmt.Sprintf("Resolve host failed for %s: %v", dialHost, err))
	} else if resolved != "" && resolved != dialHost {
		logger.Logger.Debug(fmt.Sprintf("Resolve host %s -> %s", dialHost, resolved))
		dialHost = resolved
	}

	addr := fmt.Sprintf("%s:%d", dialHost, port)

	// 使用带超时的 TCP 连接（10秒超时）
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	// 在 TCP 连接上建立 SSH 连接（30秒握手超时）
	sshConn, chans, reqs, err := gossh.NewClientConn(conn, addr, configs)
	if err != nil {
		conn.Close()
		logger.Logger.Error(err)
		return nil, err
	}

	client := gossh.NewClient(sshConn, chans, reqs)
	return client, nil
}

// ParseRawCommand ParseRawCommand
func ParseRawCommand(command string) (string, []string, error) {
	parts := strings.Split(command, " ")

	if len(parts) < 1 {
		return "", nil, errors.New("No command in payload: " + command)
	}

	if len(parts) < 2 {
		return parts[0], []string{}, nil
	}

	return parts[0], parts[1:], nil
}

// ParseRemainingCommand removes the used portion from rawCmd
func ParseRemainingCommand(rawCmd string) (string, []string, error) {
	// List of known SSH parameters to be removed
	sshParams := map[string]bool{
		"-p": true,
		"-i": true,
		"-l": true,
		"-o": true,
		"-P": true,
	}

	parts := strings.Fields(rawCmd)
	var remainingParts []string
	skipNext := false

	for i, part := range parts {
		if skipNext {
			skipNext = false
			continue
		}

		// Check for SSH parameters
		if _, isSSHParam := sshParams[part]; isSSHParam {
			skipNext = true // Skip the next part which is the value of the SSH parameter
			continue
		}

		// If not an SSH parameter, add to remaining parts
		remainingParts = parts[i:]
		break
	}

	if len(remainingParts) == 0 {
		return "", nil, nil
	}

	remainingCmd := remainingParts[0]
	remainingArgs := remainingParts[1:]
	return remainingCmd, remainingArgs, nil
}

// ErrorInfo ErrorInfo
func ErrorInfo(err error, sess *ssh.Session) {
	read := color.New(color.FgRed)
	read.Fprint(*sess, fmt.Sprintf("%s\n", err))
}
