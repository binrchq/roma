package connect

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strings"
	"text/tabwriter"
	"time"

	clickhousecli "binrc.com/dbcli/clickhouse-cli"
	elasticsearchcli "binrc.com/dbcli/elasticsearch-cli"
	mongodbcli "binrc.com/dbcli/mongodb-cli"
	mssqlcli "binrc.com/dbcli/mssql-cli"
	mysqlcli "binrc.com/dbcli/mysql-cli"
	postgrescli "binrc.com/dbcli/postgres-cli"
	rediscli "binrc.com/dbcli/redis-cli"
	"binrc.com/roma/core/api"
	"binrc.com/roma/core/connector"
	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/sshd"
	"binrc.com/roma/core/types"
	"github.com/loganchef/ssh"
)

// NewConnectionWithCommand 非交互式执行命令
func NewConnectionWithCommand(sess *ssh.Session, resModel model.Resource, resType string, command string) (interface{}, error) {
	ConnectionLoop := resModel.GetConnect()
	if ConnectionLoop == nil {
		return nil, errors.New("缺少连接方式")
	}

	// 根据资源类型处理不同的连接逻辑
	switch strings.ToLower(resType) {
	case "database":
		return handleDatabaseCommand(sess, ConnectionLoop, resModel, command)
	case "linux", "docker", "router", "switch":
		// 所有 SSH 类型的资源都通过 SSH 执行命令
		return handleSSHCommand(sess, ConnectionLoop, resModel, resType, command)
	default:
		return nil, fmt.Errorf("资源类型 %s 不支持非交互式命令执行", resType)
	}
}

func NewConnectionLoop(sess *ssh.Session, resModel model.Resource, resType string) error {
	// 将 r 转换为相应的资源类型并创建资源
	ConnectionLoop := resModel.GetConnect()
	if ConnectionLoop == nil {
		return errors.New("缺少连接方式")
	}

	// 根据资源类型处理不同的连接逻辑
	switch strings.ToLower(resType) {
	case "linux":
		return handleLinuxConnection(sess, ConnectionLoop)
	case "docker":
		return handleDockerConnection(sess, ConnectionLoop, resModel)
	case "database":
		return handleDatabaseConnection(sess, ConnectionLoop, resModel)
	case "windows":
		return handleWindowsConnection(sess, ConnectionLoop, resModel)
	case "router":
		return handleRouterConnection(sess, ConnectionLoop, resModel)
	case "switch":
		return handleSwitchConnection(sess, ConnectionLoop, resModel)
	default:
		// 默认行为：尝试 SSH 连接
		return handleLinuxConnection(sess, ConnectionLoop)
	}
}

// handleLinuxConnection 处理 Linux 服务器连接（标准 SSH）
func handleLinuxConnection(sess *ssh.Session, connections []*types.Connection) error {
	// 收集所有 SSH 连接配置
	sshConnections := []*types.Connection{}
	for _, connection := range connections {
		if connection.Type == constants.ConnectSSH && connection.Host != "" && connection.Port != 0 {
			sshConnections = append(sshConnections, connection)
		}
	}

	if len(sshConnections) == 0 {
		return errors.New("没有可用的 SSH 连接配置")
	}

	// 显示连接提示
	if len(sshConnections) == 1 {
		fmt.Fprintf(*sess, "[*] Connecting to %s:%d ...\n", sshConnections[0].Host, sshConnections[0].Port)
	} else {
		fmt.Fprintf(*sess, "[*] Trying %d addresses ...\n", len(sshConnections))
	}

	// 使用 channel 来接收第一个成功的连接
	type result struct {
		conn *types.Connection
		err  error
	}
	resultCh := make(chan result, len(sshConnections))

	// 并发测试所有连接（只测试连通性，不建立 Terminal）
	for _, conn := range sshConnections {
		go func(c *types.Connection) {
			// 测试 SSH 连接是否能建立
			client, err := sshd.NewSSHClient(c.Host, c.Port, c.Username, c.PrivateKey, "linux")
			if client != nil {
				client.Close() // 立即关闭测试连接
			}
			resultCh <- result{conn: c, err: err}
		}(conn)
	}

	// 等待第一个成功的连接
	var successConn *types.Connection
	var lastErr error
	for i := 0; i < len(sshConnections); i++ {
		res := <-resultCh
		if res.err == nil && successConn == nil {
			// 找到第一个成功的连接
			successConn = res.conn
			break
		}
		lastErr = res.err
	}

	if successConn == nil {
		// 所有连接都失败
		return fmt.Errorf("[-] Connection failed: %v", lastErr)
	}

	// 使用成功的连接建立 Terminal
	fmt.Fprintf(*sess, "[+] Connected to %s:%d\n", successConn.Host, successConn.Port)
	return sshd.NewTerminal(sess, successConn.Host, successConn.Port, successConn.Username, successConn.PrivateKey, "linux")
}

// handleDockerConnection 处理 Docker 容器连接（直接 SSH 到容器）
func handleDockerConnection(sess *ssh.Session, connections []*types.Connection, resModel model.Resource) error {
	// Docker 容器直接通过 SSH 连接（容器内运行了 sshd）
	// 连接方式和 Linux 一样，只是打印容器相关提示
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	fmt.Fprintf(tw, "Docker: %s\n", resModel.GetName())
	fmt.Fprintf(tw, "------------------------------------------------------------\n")

	tw.Flush()
	fmt.Fprint(*sess, buffer.String())

	// 直接 SSH 连接到容器
	return handleLinuxConnection(sess, connections)
}

// handleDatabaseCommand 非交互式执行数据库命令
func handleDatabaseCommand(sess *ssh.Session, connections []*types.Connection, resModel model.Resource, command string) (interface{}, error) {
	dbConfig, ok := resModel.(*model.DatabaseConfig)
	if !ok {
		return nil, errors.New("资源类型不是数据库配置")
	}

	// 使用 DatabaseConnector 执行查询
	connector := connector.NewDatabaseConnector(dbConfig)

	// 支持多个 SQL 语句（用分号分隔）
	statements := splitSQLStatements(command)

	// 清理和去重：移除重复的语句
	seen := make(map[string]bool)
	uniqueStatements := []string{}
	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		// 移除末尾的分号（splitSQLStatements 已经处理了分号，但可能还有残留）
		stmt = strings.TrimSuffix(stmt, ";")
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		// 使用小写比较去重（SQL 语句大小写不敏感）
		stmtLower := strings.ToLower(stmt)
		if !seen[stmtLower] {
			seen[stmtLower] = true
			uniqueStatements = append(uniqueStatements, stmt)
		}
	}

	// 如果没有找到任何语句，直接返回（避免重复执行）
	if len(uniqueStatements) == 0 {
		return "", nil
	}

	var allOutput strings.Builder

	for i, stmt := range uniqueStatements {
		// 如果是多个语句，添加分隔符
		if len(uniqueStatements) > 1 {
			if i > 0 {
				allOutput.WriteString("\n")
			}
			allOutput.WriteString("------------------------------------------------------------\n")
			allOutput.WriteString(fmt.Sprintf("执行: %s\n", stmt))
			allOutput.WriteString("------------------------------------------------------------\n")
		}

		result, err := connector.ExecuteQuery(stmt)
		if err != nil {
			return nil, fmt.Errorf("执行失败 [%s]: %v", stmt, err)
		}

		// 格式化输出结果
		var output string
		switch v := result.(type) {
		case string:
			output = v
		case []map[string]interface{}:
			// 格式化输出结果
			if len(v) > 0 {
				var buffer bytes.Buffer

				// 获取列名
				keys := make([]string, 0, len(v[0]))
				for k := range v[0] {
					keys = append(keys, k)
				}

				// 如果是单列结果（如 SHOW databases），直接输出值，不显示表头
				if len(keys) == 1 {
					key := keys[0]
					for _, row := range v {
						val := row[key]
						if val == nil {
							buffer.WriteString("NULL\n")
						} else {
							// 处理 []byte 类型（MySQL 驱动返回的字符串可能是 []byte）
							if b, ok := val.([]byte); ok {
								buffer.WriteString(string(b))
							} else {
								buffer.WriteString(fmt.Sprintf("%v", val))
							}
							buffer.WriteString("\n")
						}
					}
				} else {
					// 多列结果，使用表格格式
					tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

					// 打印表头
					for _, key := range keys {
						fmt.Fprintf(tw, "%s\t", key)
					}
					fmt.Fprintln(tw)

					// 打印分隔线
					for range keys {
						fmt.Fprintf(tw, "---\t")
					}
					fmt.Fprintln(tw)

					// 打印数据
					for _, row := range v {
						for _, key := range keys {
							val := row[key]
							if val == nil {
								fmt.Fprintf(tw, "NULL\t")
							} else {
								// 处理 []byte 类型（MySQL 驱动返回的字符串可能是 []byte）
								if b, ok := val.([]byte); ok {
									fmt.Fprintf(tw, "%s\t", string(b))
								} else {
									fmt.Fprintf(tw, "%v\t", val)
								}
							}
						}
						fmt.Fprintln(tw)
					}
					tw.Flush()
				}
				output = buffer.String()
			} else {
				output = "查询结果为空\n"
			}
		default:
			output = fmt.Sprintf("%v\n", v)
		}

		allOutput.WriteString(output)
	}

	// 输出到 SSH 会话
	fmt.Fprint(*sess, allOutput.String())
	// 返回空字符串，避免在 TUI 中重复输出（已经在上面输出到 sess 了）
	return "", nil
}

// splitSQLStatements 按分号分割 SQL 语句，但保留字符串中的分号
func splitSQLStatements(sql string) []string {
	var statements []string
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(sql); i++ {
		char := sql[i]

		switch char {
		case '\'':
			if !inDoubleQuote {
				inSingleQuote = !inSingleQuote
			}
			current.WriteByte(char)
		case '"':
			if !inSingleQuote {
				inDoubleQuote = !inDoubleQuote
			}
			current.WriteByte(char)
		case ';':
			if !inSingleQuote && !inDoubleQuote {
				// 不在引号内，这是真正的语句分隔符
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					statements = append(statements, stmt)
				}
				current.Reset()
			} else {
				current.WriteByte(char)
			}
		default:
			current.WriteByte(char)
		}
	}

	// 处理最后一个语句（可能没有分号结尾）
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		statements = append(statements, stmt)
	}

	return statements
}

// handleSSHCommand 非交互式执行 SSH 命令（适用于 Linux、Docker、Router、Switch）
func handleSSHCommand(sess *ssh.Session, connections []*types.Connection, resModel model.Resource, resType string, command string) (interface{}, error) {
	// 收集所有 SSH 连接配置
	sshConnections := []*types.Connection{}
	for _, connection := range connections {
		if connection.Type == constants.ConnectSSH && connection.Host != "" && connection.Port != 0 {
			sshConnections = append(sshConnections, connection)
		}
	}

	if len(sshConnections) == 0 {
		return nil, errors.New("没有可用的 SSH 连接配置")
	}

	// 获取用户名和IP地址用于审计日志
	username := (*sess).User()
	remoteAddr := (*sess).RemoteAddr()
	ipAddress := ""
	if remoteAddr != nil {
		ipAddress = remoteAddr.String()
		// 提取IP地址（去掉端口）
		if idx := strings.LastIndex(ipAddress, ":"); idx != -1 {
			ipAddress = ipAddress[:idx]
		}
	}

	// 获取资源信息
	resourceID := uint(0)
	resourceName := resModel.GetName()
	if resModel.GetID() > 0 {
		resourceID = uint(resModel.GetID())
	}

	// 检测高危命令并记录审计日志
	highRisk := isHighRiskCommand(command)
	if highRisk {
		recordTUICommandAuditLog(username, command, resType, resourceID, resourceName, ipAddress, "pending", "")
	}

	// 尝试第一个连接
	successConn := sshConnections[0]
	client, err := sshd.NewSSHClient(successConn.Host, successConn.Port, successConn.Username, successConn.PrivateKey, "linux")
	if err != nil {
		if highRisk {
			recordTUICommandAuditLog(username, command, resType, resourceID, resourceName, ipAddress, "failed", fmt.Sprintf("连接失败: %v", err))
		}
		return nil, fmt.Errorf("连接失败: %v", err)
	}
	defer client.Close()

	// 创建会话并执行命令
	session, err := client.NewSession()
	if err != nil {
		if highRisk {
			recordTUICommandAuditLog(username, command, resType, resourceID, resourceName, ipAddress, "failed", fmt.Sprintf("创建会话失败: %v", err))
		}
		return nil, fmt.Errorf("创建会话失败: %v", err)
	}
	defer session.Close()

	// 执行命令并获取输出
	output, err := session.CombinedOutput(command)
	if err != nil {
		errMsg := fmt.Sprintf("执行失败: %v", err)
		if highRisk {
			recordTUICommandAuditLog(username, command, resType, resourceID, resourceName, ipAddress, "failed", errMsg)
		}
		return nil, fmt.Errorf("%s, 输出: %s", errMsg, string(output))
	}

	// 如果之前记录了审计日志，更新状态为成功
	if highRisk {
		recordTUICommandAuditLog(username, command, resType, resourceID, resourceName, ipAddress, "success", "")
	}

	return string(output), nil
}

// isHighRiskCommand 检测是否为高危命令（本地函数，避免循环依赖）
func isHighRiskCommand(command string) bool {
	command = strings.ToLower(strings.TrimSpace(command))

	// 高危命令关键词列表
	highRiskKeywords := []string{
		"rm -rf",
		"rm -r",
		"rm -f",
		"dd if=",
		"mkfs",
		"fdisk",
		"chmod 777",
		"chmod +x",
		"chown",
		"systemctl stop",
		"systemctl disable",
		"kill -9",
		"> /dev/null",
		"| sh",
		"| bash",
		"curl |",
		"wget |",
		"format",
		"del /f /s /q",
		"format c:",
		"shutdown",
		"reboot",
		"halt",
		"poweroff",
		"init 0",
		"init 6",
		"iptables -f",
		"iptables -x",
		"drop database",
		"truncate",
		"drop table",
		"delete from",
		"update.*set.*=",
		"alter table",
		"grant all",
		"revoke",
	}

	for _, keyword := range highRiskKeywords {
		if strings.Contains(command, keyword) {
			return true
		}
	}

	return false
}

// recordTUICommandAuditLog 记录TUI中命令执行的审计日志
func recordTUICommandAuditLog(username, command, resourceType string, resourceID uint, resourceName, ipAddress, status, errorMessage string) {
	// 调用 API 中的函数
	api.RecordTUICommandAuditLog(username, command, resourceType, resourceID, resourceName, ipAddress, status, errorMessage)
}

// handleDatabaseConnection 处理数据库连接（打印连接信息和示例 SQL）
func handleDatabaseConnection(sess *ssh.Session, connections []*types.Connection, resModel model.Resource) error {
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	// 获取数据库配置信息
	dbConfig, ok := resModel.(*model.DatabaseConfig)
	if !ok {
		return errors.New("资源类型不是数据库配置")
	}

	dbType := strings.ToLower(dbConfig.DatabaseType)

	fmt.Fprintf(tw, "Database: %s (%s)\n", dbConfig.DatabaseNick, dbConfig.DatabaseType)
	fmt.Fprintf(tw, "------------------------------------------------------------\n")

	// 打印连接信息
	for _, connection := range connections {
		if connection.Type == constants.ConnectDatabase {
			fmt.Fprintf(tw, "Host: %s:%d\n", connection.Host, connection.Port)
			fmt.Fprintf(tw, "User: %s / %s\n", connection.Username, connection.Password)
			fmt.Fprintf(tw, "DB: %s\n", dbConfig.DatabaseName)

			// 根据数据库类型打印连接命令
			switch dbType {
			case "mysql":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  mysql -h %s -P %d -u %s -p'%s' %s\n", connection.Host, connection.Port, connection.Username, connection.Password, dbConfig.DatabaseName)
			case "postgresql", "postgres":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  PGPASSWORD='%s' psql -h %s -p %d -U %s -d %s\n", connection.Password, connection.Host, connection.Port, connection.Username, dbConfig.DatabaseName)
			case "mongodb", "mongo":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  mongo mongodb://%s:%s@%s:%d/%s\n", connection.Username, connection.Password, connection.Host, connection.Port, dbConfig.DatabaseName)
			case "redis":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  redis-cli -h %s -p %d -a '%s'\n", connection.Host, connection.Port, connection.Password)
			case "mssql", "sqlserver":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  sqlcmd -S %s,%d -U %s -P '%s' -d %s\n", connection.Host, connection.Port, connection.Username, connection.Password, dbConfig.DatabaseName)
			case "clickhouse":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  clickhouse-client --host %s --port %d --user %s --password '%s' --database %s\n", connection.Host, connection.Port, connection.Username, connection.Password, dbConfig.DatabaseName)
			case "elasticsearch", "es":
				fmt.Fprintf(tw, "\nConnect:\n")
				fmt.Fprintf(tw, "  curl -u %s:%s http://%s:%d\n", connection.Username, connection.Password, connection.Host, connection.Port)
			}
			fmt.Fprintf(tw, "------------------------------------------------------------\n")
			break
		}
	}

	tw.Flush()
	fmt.Fprint(*sess, buffer.String())

	// 连接数据库 CLI
	for _, connection := range connections {
		if connection.Type == constants.ConnectDatabase {
			fmt.Fprintf(*sess, "[*] Connecting ...\n")

			switch dbType {
			case "mysql":
				// 使用 Config 配置
				config := &mysqlcli.Config{
					Host:            connection.Host,
					Port:            connection.Port,
					Username:        connection.Username,
					Password:        connection.Password,
					Database:        dbConfig.DatabaseName,
					Charset:         "utf8mb4",
					Collation:       "utf8mb4_unicode_ci",
					Timeout:         10 * time.Second,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					ParseTime:       true,
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 30 * time.Minute,
				}

				cli := mysqlcli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			case "postgresql", "postgres":
				// 使用 Config 配置
				config := &postgrescli.Config{
					Host:            connection.Host,
					Port:            connection.Port,
					Username:        connection.Username,
					Password:        connection.Password,
					Database:        dbConfig.DatabaseName,
					SSLMode:         "disable",
					ConnectTimeout:  10,
					ApplicationName: "psql",
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 30 * time.Minute,
				}

				cli := postgrescli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			case "redis":
				// 使用 Config 配置
				config := &rediscli.Config{
					Host:         connection.Host,
					Port:         connection.Port,
					Password:     connection.Password,
					DB:           0,
					MaxRetries:   3,
					DialTimeout:  10 * time.Second,
					ReadTimeout:  30 * time.Second,
					WriteTimeout: 30 * time.Second,
					PoolSize:     10,
					MinIdleConns: 5,
					PoolTimeout:  4 * time.Second,
					IdleTimeout:  5 * time.Minute,
				}

				cli := rediscli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			case "mongodb", "mongo":
				// 使用 Config 配置
				config := &mongodbcli.Config{
					Host:            connection.Host,
					Port:            connection.Port,
					Username:        connection.Username,
					Password:        connection.Password,
					Database:        dbConfig.DatabaseName,
					ConnectTimeout:  10 * time.Second,
					MinPoolSize:     5,
					MaxPoolSize:     10,
					MaxConnIdleTime: 30 * time.Minute,
				}

				cli := mongodbcli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			case "mssql", "sqlserver":
				// 使用 Config 配置
				config := &mssqlcli.Config{
					Host:            connection.Host,
					Port:            connection.Port,
					Username:        connection.Username,
					Password:        connection.Password,
					Database:        dbConfig.DatabaseName,
					Encrypt:         "disable",
					TrustServerCert: true,
					ConnectTimeout:  10,
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 30 * time.Minute,
					ApplicationName: "roma-mssql",
				}

				cli := mssqlcli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			case "clickhouse":
				// 使用 Config 配置
				config := &clickhousecli.Config{
					Host:            connection.Host,
					Port:            connection.Port,
					Username:        connection.Username,
					Password:        connection.Password,
					Database:        dbConfig.DatabaseName,
					Secure:          false,
					SkipVerify:      true,
					DialTimeout:     10 * time.Second,
					ReadTimeout:     30 * time.Second,
					WriteTimeout:    30 * time.Second,
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxLifetime: 30 * time.Minute,
					Compression:     "lz4",
				}

				cli := clickhousecli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			case "elasticsearch", "es":
				// 使用 Config 配置
				config := &elasticsearchcli.Config{
					Host:                 connection.Host,
					Port:                 connection.Port,
					Username:             connection.Username,
					Password:             connection.Password,
					Scheme:               "http",
					MaxRetries:           3,
					CompressRequestBody:  false,
					DiscoverNodesOnStart: false,
				}

				cli := elasticsearchcli.NewCLIWithConfig(*sess, config)
				if err := cli.Connect(); err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				defer cli.Close()

				// 启动 CLI，退出后直接返回
				err := cli.Start()
				if err != nil {
					fmt.Fprintf(*sess, "\n[-] CLI error: %v\n", err)
				}

				return nil

			default:
				// 未知数据库类型只测试 TCP 连接
				conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", connection.Host, connection.Port), 5*time.Second)
				if conn != nil {
					conn.Close()
				}
				if err != nil {
					return fmt.Errorf("[-] Connection failed: %v", err)
				}
				fmt.Fprintf(*sess, "[+] Connection successful (CLI not available for %s)\n\n", dbType)
				return nil
			}
		}
	}

	return nil
}

// handleWindowsConnection 处理 Windows 服务器连接（打印 RDP 信息）
func handleWindowsConnection(sess *ssh.Session, connections []*types.Connection, resModel model.Resource) error {
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	winConfig, ok := resModel.(*model.WindowsConfig)
	if !ok {
		return errors.New("资源类型不是 Windows 配置")
	}

	fmt.Fprintf(tw, "Windows: %s\n", winConfig.Hostname)
	fmt.Fprintf(tw, "------------------------------------------------------------\n")

	// 打印 RDP 连接信息
	for _, connection := range connections {
		if connection.Type == constants.ConnectRDP {
			fmt.Fprintf(tw, "Host: %s:%d\n", connection.Host, connection.Port)
			fmt.Fprintf(tw, "User: %s / %s\n", connection.Username, connection.Password)
			fmt.Fprintf(tw, "\nConnect:\n")
			fmt.Fprintf(tw, "  mstsc /v:%s:%d\n", connection.Host, connection.Port)
			fmt.Fprintf(tw, "------------------------------------------------------------\n")
		}
	}

	fmt.Fprintf(tw, "\n")

	tw.Flush()
	fmt.Fprint(*sess, buffer.String())

	return nil
}

// handleRouterConnection 处理路由器连接（打印 Web 信息 + SSH 连接）
func handleRouterConnection(sess *ssh.Session, connections []*types.Connection, resModel model.Resource) error {
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	routerConfig, ok := resModel.(*model.RouterConfig)
	if !ok {
		return errors.New("资源类型不是路由器配置")
	}

	fmt.Fprintf(tw, "Router: %s\n", routerConfig.RouterName)
	fmt.Fprintf(tw, "------------------------------------------------------------\n")

	// 打印 Web 管理界面信息
	hasWeb := false
	for _, connection := range connections {
		if connection.Type == constants.ConnectHTTP {
			hasWeb = true
			protocol := "http"
			if connection.Port == 443 {
				protocol = "https"
			}
			webURL := fmt.Sprintf("%s://%s:%d", protocol, connection.Host, connection.Port)

			fmt.Fprintf(tw, "Web: %s\n", webURL)
			fmt.Fprintf(tw, "User: %s / %s\n", connection.Username, connection.Password)
			break
		}
	}

	if hasWeb {
		fmt.Fprintf(tw, "------------------------------------------------------------\n")
	}

	tw.Flush()
	fmt.Fprint(*sess, buffer.String())

	// 尝试 SSH 连接
	sshConnections := []*types.Connection{}
	for _, conn := range connections {
		if conn.Type == constants.ConnectSSH {
			sshConnections = append(sshConnections, conn)
		}
	}

	if len(sshConnections) > 0 {
		// 建立 SSH 连接
		return handleLinuxConnection(sess, sshConnections)
	}

	return nil
}

// handleSwitchConnection 处理交换机连接（SSH + 命令提示）
func handleSwitchConnection(sess *ssh.Session, connections []*types.Connection, resModel model.Resource) error {
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	switchConfig, ok := resModel.(*model.SwitchConfig)
	if !ok {
		return errors.New("资源类型不是交换机配置")
	}

	fmt.Fprintf(tw, "Switch: %s\n", switchConfig.SwitchName)
	fmt.Fprintf(tw, "------------------------------------------------------------\n")

	tw.Flush()
	fmt.Fprint(*sess, buffer.String())

	// 建立 SSH 连接
	return handleLinuxConnection(sess, connections)
}
