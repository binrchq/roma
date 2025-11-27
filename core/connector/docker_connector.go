package connector

import (
	"fmt"
	"strings"

	"binrc.com/roma/core/model"
	"binrc.com/roma/core/utils"
	gossh "golang.org/x/crypto/ssh"
)

// DockerConnector Docker 连接器（通过 SSH 到宿主机管理容器）
// 注意：跳板机场景下，直接 SSH 到容器内部（通过 loop.go）
//       MCP/API 场景下，SSH 到宿主机执行 docker 命令
type DockerConnector struct {
	Config    *model.DockerConfig
	SSHClient *gossh.Client
}

// NewDockerConnector 创建 Docker 连接器
func NewDockerConnector(config *model.DockerConfig) *DockerConnector {
	return &DockerConnector{
		Config: config,
	}
}

// Connect 连接到 Docker 宿主机（MCP/API 场景）
// 注意：这里连接的是宿主机，而不是容器内部
func (d *DockerConnector) Connect() error {
	// 宿主机信息（假设配置的 IP 就是宿主机 IP）
	host := d.Config.IPv4Priv
	if host == "" {
		host = d.Config.IPv6
	}

	// 宿主机 SSH 端口（默认 22，不是容器映射的端口）
	port := 22
	
	// 注意：这里需要宿主机的 SSH 认证信息
	// 当前模型中使用的是容器的认证信息
	// 实际生产环境可能需要分别存储宿主机和容器的认证信息
	config := &gossh.ClientConfig{
		User: d.Config.Username, // 宿主机用户
		Auth: []gossh.AuthMethod{},
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
	}

	// 优先使用私钥
	if d.Config.PrivateKey != "" {
		signer, err := gossh.ParsePrivateKey([]byte(d.Config.PrivateKey))
		if err == nil {
			config.Auth = append(config.Auth, gossh.PublicKeys(signer))
		}
	}

	if d.Config.Password != "" {
		// 解密密码
		decryptedPassword, err := utils.DecryptPassword(d.Config.Password)
		if err != nil {
			return fmt.Errorf("密码解密失败: %v", err)
		}
		config.Auth = append(config.Auth, gossh.Password(decryptedPassword))
	}

	client, err := gossh.Dial("tcp", fmt.Sprintf("%s:%d", host, port), config)
	if err != nil {
		return fmt.Errorf("连接宿主机失败: %v", err)
	}

	d.SSHClient = client
	return nil
}

// ExecuteDockerCommand 执行 Docker 命令
func (d *DockerConnector) ExecuteDockerCommand(command string) (string, error) {
	if d.SSHClient == nil {
		if err := d.Connect(); err != nil {
			return "", err
		}
	}

	session, err := d.SSHClient.NewSession()
	if err != nil {
		return "", fmt.Errorf("创建 SSH session 失败: %v", err)
	}
	defer session.Close()

	// 执行 docker 命令
	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("命令执行失败: %v\n输出: %s", err, string(output))
	}

	return string(output), nil
}

// ListContainers 列出所有容器
func (d *DockerConnector) ListContainers() (string, error) {
	return d.ExecuteDockerCommand("docker ps -a --format 'table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}'")
}

// InspectContainer 查看容器详情
func (d *DockerConnector) InspectContainer(containerID string) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker inspect %s", containerID))
}

// ContainerLogs 查看容器日志
func (d *DockerConnector) ContainerLogs(containerID string, lines int) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker logs --tail %d %s", lines, containerID))
}

// StartContainer 启动容器
func (d *DockerConnector) StartContainer(containerID string) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker start %s", containerID))
}

// StopContainer 停止容器
func (d *DockerConnector) StopContainer(containerID string) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker stop %s", containerID))
}

// RestartContainer 重启容器
func (d *DockerConnector) RestartContainer(containerID string) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker restart %s", containerID))
}

// ExecInContainer 在容器中执行命令
func (d *DockerConnector) ExecInContainer(containerID, command string) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker exec %s %s", containerID, command))
}

// GetContainerStats 获取容器资源使用情况
func (d *DockerConnector) GetContainerStats(containerID string) (string, error) {
	return d.ExecuteDockerCommand(fmt.Sprintf("docker stats %s --no-stream --no-trunc", containerID))
}

// ListImages 列出镜像
func (d *DockerConnector) ListImages() (string, error) {
	return d.ExecuteDockerCommand("docker images --format 'table {{.Repository}}\t{{.Tag}}\t{{.ID}}\t{{.Size}}'")
}

// ListNetworks 列出网络
func (d *DockerConnector) ListNetworks() (string, error) {
	return d.ExecuteDockerCommand("docker network ls")
}

// ListVolumes 列出卷
func (d *DockerConnector) ListVolumes() (string, error) {
	return d.ExecuteDockerCommand("docker volume ls")
}

// GetConnectionInfo 获取连接信息
func (d *DockerConnector) GetConnectionInfo() map[string]interface{} {
	host := d.Config.IPv4Priv
	if host == "" {
		host = d.Config.IPv6
	}

	commands := []string{
		"docker ps -a              # 查看所有容器",
		"docker images             # 查看镜像",
		"docker logs <container>   # 查看日志",
		"docker exec -it <container> /bin/bash  # 进入容器",
		"docker stats             # 查看资源使用",
	}

	return map[string]interface{}{
		"type":           "docker",
		"container_name": d.Config.ContainerName,
		"host":           host,
		"port":           d.Config.Port,
		"username":       d.Config.Username,
		"ssh_command":    fmt.Sprintf("ssh %s@%s -p %d", d.Config.Username, host, d.Config.Port),
		"common_commands": strings.Join(commands, "\n"),
	}
}

// Close 关闭连接
func (d *DockerConnector) Close() error {
	if d.SSHClient != nil {
		return d.SSHClient.Close()
	}
	return nil
}

