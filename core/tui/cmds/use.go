package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/tabwriter"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/tui/cmds/itface"
)

type Use struct{}

func (u *Use) Use(resType string) (string, error) {
	resourceTypes := constants.GetResourceType()
	if resType == "~" {
		if len(resourceTypes) == 0 {
			return "", errors.New("no resource types available")
		}
		return "~", nil
	}
	if !sliceContains(resourceTypes, resType) {
		return "", errors.New("invalid resource type")
	}
	return resType, nil
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
func (u *Use) Usage() string {
	helpItems := []string{
		"use <type> - Change the current resource type",
	}
	// 获取所有资源类型
	resourceTypes := constants.GetResourceType()
	// 如果资源类型列表为空，直接返回帮助信息
	if len(resourceTypes) == 0 {
		return strings.Join(helpItems, "\n")
	}
	// 设置默认切换的目标为第一个资源类型
	defaultSwitchTarget := resourceTypes[0]
	// 如果输入的资源类型为 "~"，则将默认切换的目标显示为帮助信息中的第一个资源类型
	helpItems = append(helpItems, fmt.Sprintf("    ~\t: lnSwitch to %s resource type", defaultSwitchTarget))

	// 将资源类型添加到帮助信息中
	for _, resType := range resourceTypes {
		helpItems = append(helpItems, fmt.Sprintf("    %s\t: Switch to %s resource type", resType, resType))
	}
	helpItems = append(helpItems, "    -h\t: Get help for this command")

	// 使用tabwriter来格式化输出
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	for _, item := range helpItems {
		//如果是最后一行就不添加换行符
		if item == helpItems[len(helpItems)-1] {
			fmt.Fprintf(tw, "%s", item)
		} else {
			fmt.Fprintf(tw, "%s\n", item)
		}
	}
	tw.Flush()

	return buffer.String()
}

//Name 返回 Use 命令的名称。
func (u *Use) Name() string {
    return "use"
}

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Use{}, Weight: 100})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Use{}, Weight: 100})
}
