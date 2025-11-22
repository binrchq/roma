package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/tui/cmds/itface"
	"github.com/rs/zerolog/log"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewUse(), Weight: 100})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewUse(), Weight: 100})
}

// Name 返回 Use 命令的名称。
func (cmd *Use) Name() string {
	return "use"
}

type Use struct {
	baseLen int
	flags   *Flags
	target  string
}

func NewUse() *Use {
	flags := &Flags{}
	flags.AddOption("h", "help", "Display this help message", BoolOption, false)
	return &Use{baseLen: 3, flags: flags, target: "~"}
}

func (cmd *Use) Execute(commands string) (string, error) {
	resourceTypes := constants.GetResourceType()
	// 解析命令行参数
	cmd.target = cmd.flags.Parse(commands[cmd.baseLen:])
	if cmd.target == "~" {
		if len(resourceTypes) == 0 {
			return "", errors.New("no resource types available")
		}
		return "~", nil
	}
	if cmd.target == "" {
		return "", errors.New("no resource type specified")
	}
	if !sliceContains(resourceTypes, cmd.target) {
		return "", errors.New("invalid resource type")
	}
	return cmd.target, nil
}

func sliceContains(slice []string, elem string) bool {
	for _, e := range slice {
		if e == elem {
			return true
		}
	}
	return false
}

// Help 返回 Use 命令的帮助信息，该命令用于更改当前资源类型。
// 参数 <type> 可以是以下任一值：
//   - linux: 切换到 Linux 资源类型
//   - windows: 切换到 Windows 资源类型
//   - database: 切换到数据库资源类型
//   - router: 切换到路由器资源类型
//   - switch: 切换到交换机资源类型
//   - docker: 切换到 Docker 资源类型
func (cmd *Use) Usage() string {
	// 获取所有资源类型
	resourceTypes := constants.GetResourceType()
	usageMsg := cmd.flags.FormatUsagef("🍂 %s", green(cmd.Name()+" [OPTIONS] TYPE"))
	usageMsg += cmd.flags.FormatUsagef("Switch to specified TYPE of resource,TYPE is %s, etc.", cyan(strings.Join(resourceTypes, ", ")))
	usageMsg += cmd.flags.FormatUsagef("Usage:")
	// 如果资源类型列表为空，直接返回帮助信息
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	// 写入Options
	log.Info().Msgf("flags: %v", cmd.flags.Options)
	tw = cmd.flags.ColorUsage(tw)
	fmt.Fprint(tw) // 换行
	tw.Flush()
	return usageMsg + buffer.String()
}
