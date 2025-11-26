package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/permissions"
	"binrc.com/roma/core/tui/cmds/itface"
	"github.com/loganchef/ssh"
	"github.com/rs/zerolog/log"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewLs(nil, ""), Weight: 50})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewLs(nil, ""), Weight: 50})
}

type Ls struct {
	baseLen int // 基础命令长度
	flags   *Flags
	target  string
	sess    ssh.Session
}

func NewLs(sess ssh.Session, typo string) *Ls {
	flags := &Flags{}
	flags.AddOption("l", "list", "Display detailed information", BoolOption, false)
	flags.AddOption("a", "all", "Display all resource", BoolOption, false)
	flags.AddOption("h", "help", "Display this help message", BoolOption, false)
	return &Ls{baseLen: 2, flags: flags, target: typo, sess: sess}
}

// Name 返回 ls 命令的名称
func (cmd *Ls) Name() string {
	return "ls"
}
func (cmd *Ls) Execute(commands string) (interface{}, error) {
	argParts := commands[cmd.baseLen:]
	args := strings.Fields(strings.TrimSpace(argParts))

	// Use Parse to handle the arguments and set the options in cmd.flags
	target := cmd.flags.Parse(args)

	if cmd.flags.GetOption("help").IsSet {
		return cmd.Usage(), nil
	}

	resourceTypes := constants.GetResourceType()
	// 优先使用 Parse 返回的 target（用户输入的资源类型）
	if target != "" && target != "~" {
		cmd.target = target
	} else if cmd.target == "~" || cmd.target == "" {
		// 如果没有指定，使用默认的第一个资源类型
		cmd.target = resourceTypes[0]
	}
	// 如果 cmd.target 已经设置（从 typo 参数），且 target 为空，保持 cmd.target

	if !sliceContains(resourceTypes, cmd.target) {
		return cmd.error("invalid resource type: " + cmd.target)
	}

	log.Debug().Msgf("resource type: %s", cmd.target)
	resList, err := cmd.handleResources(cmd.target)
	if err != nil {
		return cmd.error(err.Error())
	}

	if cmd.flags.GetOption("list").IsSet {
		if cmd.flags.GetOption("all").IsSet {
			return cmd.ResourceLines(resList), nil
		}
		return cmd.ResourceLines(resList), nil
	}

	if cmd.flags.GetOption("all").IsSet {
		return cmd.Resources(resList), nil
	}

	return cmd.Resources(resList), nil
}

func (cmd *Ls) Resources(resList []model.Resource) string {
	// 创建一个字节缓冲区用于保存格式化后的输出
	nameList := []string{}
	if len(resList) > 0 {
		for _, res := range resList {
			// fmt.Fprintf(writer, "%s\t", res.GetName())
			nameList = append(nameList, res.GetName())
		}
	}
	// writer.Flush()
	return strings.Join(nameList, " ")
}

func (cmd *Ls) ResourceLines(resList []model.Resource) string {
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)

	if len(resList) > 0 {
		titleArr := resList[0].GetTitle()
		if len(titleArr) > 0 {
			// 打印标题
			for _, title := range titleArr {
				fmt.Fprintf(tw, "%s\t", title)
			}
			fmt.Fprintln(tw) // 换行

			// 打印内容
			for _, res := range resList {
				for _, item := range res.GetLine() {
					fmt.Fprintf(tw, "%s\t", item)
				}
				fmt.Fprintln(tw) // 换行
			}
		}
	}

	tw.Flush()
	return buffer.String()
}

func (cmd *Ls) error(msg string) (interface{}, error) {
	errMsg := msg + "\n"
	errMsg += "Try 'ls --help' for more information."
	return nil, errors.New(errMsg)
}

func (cmd *Ls) Usage() string {
	resourceTypes := constants.GetResourceType()
	usageMsg := cmd.flags.FormatUsagef("🍂 %s", green(cmd.Name()+" [OPTIONS] TYPE"))
	usageMsg += cmd.flags.FormatUsagef("List the specified TYPE of resource,TYPE is %s, etc.", cyan(strings.Join(resourceTypes, ", ")))
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

func (cmd *Ls) handleResources(resourceType string) ([]model.Resource, error) {
	opUser := operation.NewUserOperation()
	user, err := opUser.GetUserByUsername(cmd.sess.User())
	if err != nil {
		return nil, errors.New("permission denied: unable to get user")
	}

	var resListA []model.Resource
	op := operation.NewResourceOperation()

	roles, err := opUser.GetUserRolesByUsername(cmd.sess.User())
	if err != nil {
		return nil, errors.New("permission denied: unable to get user roles")
	}

	for _, role := range roles {
		resList, _ := op.GetResourceListByRoleId(role.ID, resourceType)
		for _, res := range resList {
			// 使用 CheckResourceAccessWithRoles 检查权限（角色 + 空间），避免重复查询用户角色
			allowed, reason := permissions.CheckResourceAccessWithRoles(user, roles, res.GetID(), resourceType, "list")
			if !allowed {
				log.Debug().Msgf("Resource %s (ID: %d) access denied: %s", res.GetName(), res.GetID(), reason)
				continue
			}

			switch resourceType {
			case constants.ResourceTypeLinux:
				if linuxRes, ok := res.(*model.LinuxConfig); ok {
					resListA = append(resListA, linuxRes)
				}
			case constants.ResourceTypeWindows:
				if windowsRes, ok := res.(*model.WindowsConfig); ok {
					resListA = append(resListA, windowsRes)
				}
			case constants.ResourceTypeDocker:
				if dockerRes, ok := res.(*model.DockerConfig); ok {
					resListA = append(resListA, dockerRes)
				}
			case constants.ResourceTypeDatabase:
				if databaseRes, ok := res.(*model.DatabaseConfig); ok {
					resListA = append(resListA, databaseRes)
				}
			case constants.ResourceTypeSwitch:
				if switchRes, ok := res.(*model.SwitchConfig); ok {
					resListA = append(resListA, switchRes)
				}
			case constants.ResourceTypeRouter:
				if routerRes, ok := res.(*model.RouterConfig); ok {
					resListA = append(resListA, routerRes)
				}
			}
		}
	}

	if len(resListA) == 0 {
		return nil, errors.New("resource of " + resourceType + " is empty or permission denied")
	}
	return resListA, nil
}
