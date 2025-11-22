package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"binrc.com/roma/core/connect"
	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/tui/cmds/itface"
	"binrc.com/roma/core/utils"
	"github.com/loganchef/ssh"
	"github.com/rs/zerolog/log"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewLn(nil, ""), Weight: 50})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewLn(nil, ""), Weight: 50})
}

func (cmd *Ln) Name() string {
	return "ln"
}

type Ln struct {
	baseLen int
	flags   *Flags
	target  string
	sess    ssh.Session
}

func NewLn(sess ssh.Session, typo string) *Ln {
	flags := &Flags{}
	flags.AddOption("t", "type", "Resource type", StringOption, typo)
	flags.AddOption("h", "help", "Display this help message", BoolOption, false)
	return &Ln{baseLen: 2, flags: flags, target: "", sess: sess}
}

func (cmd *Ln) Execute(commands string) (interface{}, error) {
	var execCommand string // 要执行的命令（在资源标识符之后）

	//看看cmd是否是ln
	if !strings.HasPrefix(commands, "ln") {
		cmd.target = strings.TrimSpace(commands)
	} else {
		argParts := commands[cmd.baseLen:]
		// 使用更智能的解析方式，保留引号内的内容
		args := parseArgsWithQuotes(strings.TrimSpace(argParts))

		// 支持 kubectl 风格的命令分隔符：ln -t TYPE RESOURCE -- COMMAND
		// 或者传统方式：ln -t TYPE RESOURCE "COMMAND"
		var resourceIndex = -1
		var commandStartIndex = -1
		skipNext := false

		// 查找 -- 分隔符
		for i, arg := range args {
			if arg == "--" {
				commandStartIndex = i + 1
				break
			}
		}

		// 先解析 flags，找到资源标识符的位置
		for i, arg := range args {
			if skipNext {
				skipNext = false
				continue
			}

			// 如果遇到 -- 分隔符，停止查找资源标识符
			if arg == "--" {
				break
			}

			// 跳过 flags
			if strings.HasPrefix(arg, "-") {
				// 如果是 StringOption（-t 或 --type），跳过下一个参数（值）
				if arg == "-t" || arg == "--type" {
					skipNext = true
				}
				continue
			}

			// 找到第一个非 flag 参数，应该是资源标识符
			if resourceIndex == -1 {
				resourceIndex = i
				cmd.target = arg
			}
		}

		// 解析 flags（但不依赖 Parse 返回的 target，因为 Parse 会返回最后一个非 flag 参数）
		cmd.flags.Parse(args)

		// 提取命令
		if commandStartIndex > 0 && commandStartIndex < len(args) {
			// 使用 -- 分隔符后的所有参数作为命令
			execCommand = strings.Join(args[commandStartIndex:], " ")
		} else if resourceIndex >= 0 && resourceIndex+1 < len(args) {
			// 没有 -- 分隔符，使用资源标识符之后的所有参数作为命令
			execCommand = strings.Join(args[resourceIndex+1:], " ")
		}

		// 移除引号（如果有）
		if execCommand != "" {
			execCommand = strings.Trim(execCommand, "\"'")
		}
	}

	resourceTypes := constants.GetResourceType()
	if cmd.flags.GetOptionValue("type").(string) == "~" {
		cmd.flags.SetOptionValue("type", resourceTypes[0])
	}
	if !sliceContains(resourceTypes, cmd.flags.GetOptionValue("type").(string)) {
		return nil, errors.New("invalid resource type,please ln -h to get itfacece")
	}

	return cmd.handleWithCommand(execCommand)
}

// parseArgsWithQuotes 解析参数，保留引号内的内容
func parseArgsWithQuotes(s string) []string {
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := byte(0)

	for i := 0; i < len(s); i++ {
		char := s[i]

		if char == '"' || char == '\'' {
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteByte(char)
			}
		} else if char == ' ' && !inQuotes {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(char)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// 处理 Linux 资源类型的逻辑
func (cmd *Ln) handle() (interface{}, error) {
	return cmd.handleWithCommand("")
}

// handleWithCommand 处理连接，支持非交互式执行命令
func (cmd *Ln) handleWithCommand(execCommand string) (interface{}, error) {
	roles, err := operation.NewUserOperation().GetUserRolesByUsername(cmd.sess.User())
	if err != nil {
		log.Error().Err(err).Msg("unable to get user roles")
		return nil, err
	}
	searchType, resA := utils.DetermineSearchType(cmd.target)
	var resListA []model.Resource
	op := operation.NewResourceOperation()
	log.Info().Msg("roles:")
	for _, role := range roles {
		resList, _ := op.GetResourceListByRoleId(role.ID, cmd.flags.GetOptionValue("type").(string))
		for _, res := range resList {
			log.Info().Msgf("-------------------------%v", res.GetName())
			log.Info().Msgf("searchType: %v", searchType)
			log.Info().Msgf("resA: %v", resA)
			if matchResource(res, searchType, resA) {
				resListA = append(resListA, res)
			}
		}
	}
	if len(resListA) == 0 {
		return nil, errors.New("resource not found")
	}
	if len(resListA) > 1 {
		return NewLs(cmd.sess, "").Resources(resListA), nil
	}
	Res := resListA[0]
	log.Info().Msgf("connecting to %v", Res.GetName())

	// 如果有命令，使用非交互式执行
	if execCommand != "" {
		return connect.NewConnectionWithCommand(&cmd.sess, Res, cmd.flags.GetOptionValue("type").(string), execCommand)
	}

	// 否则使用交互式连接
	err = connect.NewConnectionLoop(&cmd.sess, Res, cmd.flags.GetOptionValue("type").(string))
	if err != nil {
		return nil, err
	}
	return "", nil
}

// 获取字段值的通用函数
func getFieldValue(res model.Resource, fieldName string) (string, bool) {
	val := reflect.ValueOf(res).Elem()
	field := val.FieldByName(fieldName)
	if !field.IsValid() {
		return "", false
	}
	switch field.Kind() {
	case reflect.String:
		return field.String(), true
	case reflect.Int:
		return strconv.Itoa(int(field.Int())), true
	default:
		return "", false
	}
}

// 匹配资源
func matchResource(res model.Resource, searchType, resA string) bool {
	fieldMappings := map[string][]string{
		utils.DETERMINE_IP:          {"IPv4Priv", "IPv4Pub", "IPv6"},
		utils.DETERMINE_DOMAIN:      {"IPv4Pub", "IPv6"},
		utils.DETERMINE_DOMAIN_PORT: {"IPv4Pub", "IPv6"},
		utils.DETERMINE_IP_PORT:     {"IPv4Pub", "IPv4Priv", "IPv6"},
		utils.DETERMINE_HOSTNAME:    {"GetName"},
	}

	fields, exists := fieldMappings[searchType]
	log.Info().Msgf("fields: %v", fields)
	if !exists {
		return false
	}

	// 对于 hostname 类型，使用精确匹配或包含匹配
	if searchType == utils.DETERMINE_HOSTNAME {
		resourceName := res.GetName()
		// 精确匹配优先
		if resourceName == resA {
			return true
		}
		// 部分匹配（向后兼容）
		if strings.Contains(resourceName, resA) {
			return true
		}
		// 也尝试反向包含（如果输入是资源名的子串）
		if strings.Contains(resA, resourceName) {
			return true
		}
		// hostname 类型只匹配 GetName，如果都不匹配就返回 false
		return false
	}
	for _, field := range fields {
		fieldValue, ok := getFieldValue(res, field)
		if !ok {
			continue
		}
		if searchType == utils.DETERMINE_DOMAIN_PORT || searchType == utils.DETERMINE_IP_PORT {
			parts := strings.Split(resA, ":")
			if len(parts) != 2 {
				return false
			}
			port, _ := getFieldValue(res, "Port")
			portActual, _ := getFieldValue(res, "PortActual")
			portIPv6, _ := getFieldValue(res, "PortIPv6")

			if fieldValue == parts[0] && (port == parts[1] || portActual == parts[1] || portIPv6 == parts[1]) {
				return true
			}
		} else if fieldValue == resA {
			return true
		}
	}

	return false
}

// Help 返回 ln 命令的帮助信息
func (cmd *Ln) Usage() string {
	resourceTypes := constants.GetResourceType()
	usageMsg := cmd.flags.FormatUsagef("🍂 %s", green(cmd.Name()+" [-t TYPE] RESOURCE [-- COMMAND]"))
	usageMsg += cmd.flags.FormatUsagef("Login the specified TYPE of resource,TYPE is %s;RESOURCE for ls Query, etc.", cyan(strings.Join(resourceTypes, ", ")))
	usageMsg += cmd.flags.FormatUsagef("")
	usageMsg += cmd.flags.FormatUsagef("Examples:")
	usageMsg += cmd.flags.FormatUsagef("  ln -t linux server1                    # 交互式登录")
	usageMsg += cmd.flags.FormatUsagef("  ln -t linux server1 -- 'df -h'         # 执行命令并退出")
	usageMsg += cmd.flags.FormatUsagef("  ln -t database links-mysql -- 'SHOW databases;'  # 执行 SQL")
	usageMsg += cmd.flags.FormatUsagef("Usage:")
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	// 写入Options
	log.Info().Msgf("flags: %v", cmd.flags.Options)
	tw = cmd.flags.ColorUsage(tw)
	fmt.Fprint(tw) // 换行
	tw.Flush()
	return usageMsg + buffer.String()
}
