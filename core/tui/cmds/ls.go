package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Ls{}, Weight: 50})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Ls{}, Weight: 50})
}

type Ls struct {
	baseLen int // 基础命令长度
	flags   *Flags
	target  string
	sess    ssh.Session
}

func NewLs(sess ssh.Session) *Ls {
	flags := &Flags{}
	flags.AddOption("l", "list", "Display detailed information", BoolOption, false)
	flags.AddOption("a", "all", "Display all resource", BoolOption, false)
	flags.AddOption("h", "help", "Display this help message", BoolOption, false)
	return &Ls{baseLen: 2, flags: flags, target: "~", sess: sess}
}

// Name 返回 ls 命令的名称
func (u *Ls) Name() string {
	return "ls"
}

func (cmd *Ls) Execute(commands string) (interface{}, error) {
	cp := CommandProcessor{}
	argParts := commands[cmd.baseLen:]
	cp.args = strings.Fields(strings.TrimSpace(argParts))
	for len(cp.args) > 0 {
		arg := cp.args[0]
		switch {
		case strings.HasPrefix(arg, "--"):
			switch arg {
			case "--help":
				cmd.flags.GetOption("help").IsSet = true
			case "--list":
				cmd.flags.GetOption("list").IsSet = true
			case "--all":
				cmd.flags.GetOption("all").IsSet = true
			default:
				return cmd.error("unknown option: " + arg)
			}
			cp.shift()
		case strings.HasPrefix(arg, "-"):
			for _, char := range arg[1:] {
				switch char {
				case 'h':
					cmd.flags.GetOption("help").IsSet = true
				case 'l':
					cmd.flags.GetOption("list").IsSet = true
				case 'a':
					cmd.flags.GetOption("all").IsSet = true
				default:
					return cmd.error("unknown option: " + string(char))
				}
			}
			cp.shift()
		default:
			cmd.target = arg
			cp.shift()
		}
	}
	if cmd.flags.GetOption("help").IsSet {
		return cmd.Usage(), nil
	}
	resourceTypes := constants.GetResourceType()
	if cmd.target == "~" || cmd.target == "" {
		cmd.target = resourceTypes[0]
	}
	if !sliceContains(resourceTypes, cmd.target) {
		return cmd.error("invalid resource type: " + cmd.target)
	}
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
	var buffer bytes.Buffer
	// 使用 tabwriter 创建一个新的写入器，并设置格式化选项
	resList = append(resList, resList...)
	resList = append(resList, resList...)
	resList = append(resList, resList...)
	resList = append(resList, resList...)
	writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	if len(resList) > 0 {
		for _, res := range resList {
			fmt.Fprintf(writer, "%s\t", res.GetName())
		}
		fmt.Fprintln(writer, "") // 最后一行换行
	}
	writer.Flush()
	return buffer.String()
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

func (u *Ls) Usage() string {
	resourceTypes := constants.GetResourceType()
	usageMsg := "Usage: ls [OPTIONS] TYPE\n"
	usageMsg += "List the specified TYPE of resource,TYPE is " + strings.Join(resourceTypes, ",") + ", etc.\n"
	usageMsg += "Options:\n"

	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	// 写入Options
	for _, opt := range u.flags.Options {
		switch opt.Type {
		case BoolOption:
			fmt.Fprintf(tw, "  -%s, --%s\t%s\n", opt.Short, opt.Long, opt.Help)
		case StringOption:
			fmt.Fprintf(tw, "  -%s, --%s=%s\t%s\n", opt.Short, opt.Long, opt.GetNameToUpper(), opt.Help)
		case ListOption:
			fmt.Fprintf(tw, "  -%s, --%s=[]%s\t%s\n", opt.Short, opt.Long, opt.GetNameToUpper(), opt.Help)
		}
	}
	fmt.Fprint(tw) // 换行
	tw.Flush()
	return usageMsg + buffer.String()
}

func (cmd *Ls) handleResources(resourceType string) ([]model.Resource, error) {
	roles, err := operation.NewUserOperation().GetUserRolesByUsername(cmd.sess.User())
	if err != nil {
		return nil, errors.New("permission denied")
	}
	var resListA []model.Resource
	op := operation.NewResourceOperation()

	for _, role := range roles {
		resList, _ := op.GetResourceListByRoleId(role.ID, resourceType)
		for _, res := range resList {
			switch resourceType {
			case constants.ResourceTypeLinux:
				if linuxRes, ok := res.(*model.LinuxConfig); ok {
					resListA = append(resListA, linuxRes)
				}
			case constants.ResourceTypeWindows:
				if windowsRes, ok := res.(*model.WindowsConfig); ok {
					resListA = append(resListA, windowsRes)
				}
			}
		}
	}

	if len(resListA) == 0 {
		return nil, errors.New("resource of " + resourceType + " is empty")
	}
	if len(resListA) > 1 {
		// 将格式化后的输出写入到标准输出
		return resListA, nil
	}
	return nil, errors.New("permission denied")
}
