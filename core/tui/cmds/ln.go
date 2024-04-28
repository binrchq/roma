package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"text/tabwriter"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/sshd"
	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

type Ln struct{}

func (l *Ln) Execute(sess *ssh.Session, args []string) error {
	resourceTypes := constants.GetResourceType()
	if len(args) == 0 {
		return errors.New("args no resource type specified,please ln -h to get itface")
	}
	resType := args[0]
	if resType == "~" {
		resType = resourceTypes[0]
	}
	if !sliceContains(resourceTypes, resType) {
		return errors.New("invalid resource type,please ln -h to get itfacece")
	}

	// 根据资源类型选择不同的处理函数
	switch resType {
	case constants.ResourceTypeLinux:
		return l.handleLinux(sess, args)
	case constants.ResourceTypeDatabase:
		return errors.New("invalid resource type,please ln -h to get itface")
	case constants.ResourceTypeWindows:
		return errors.New("invalid resource type,please ln -h to get itface")
	case constants.ResourceTypeDocker:
		return errors.New("invalid resource type,please ln -h to get itfacecece")
	case constants.ResourceTypeRouter:
		return errors.New("invalid resource type,please ln -h to get help")
	case constants.ResourceTypeSwitch:
		return errors.New("invalid resource type,please ln -h to get help")
	default:
		return errors.New("invalid resource type,please ln -h to get help")
	}
}

// 处理 Linux 资源类型的逻辑
func (l *Ln) handleLinux(sess *ssh.Session, args []string) error {
	roles, err := operation.NewUserOperation().GetUserRolesByUsername((*sess).User())
	if err != nil {
		return err
	}

	searchType, resA := determineSearchType(args[1])

	var resListA []model.Resource
	op := operation.NewResourceOperation()
	for _, role := range roles {
		resList, _ := op.GetResourceListByRoleId(role.ID, constants.ResourceTypeLinux)
		for _, res := range resList {
			linuxRes, ok := res.(*model.LinuxConfig)
			if !ok {
				continue
			}
			if matchResource(linuxRes, searchType, resA) {
				resListA = append(resListA, linuxRes)
			}
		}
	}
	if len(resListA) == 0 {
		return errors.New("resource not found")
	}
	if len(resListA) > 1 {

		// 创建一个字节缓冲区用于保存格式化后的输出
		var buffer bytes.Buffer

		// 使用 tabwriter 创建一个新的写入器，并设置格式化选项
		writer := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
		// 定义每行的最大列数
		maxColumns := 4
		currentColumns := 0

		// 打印资源列表中的每个主机名
		for _, res := range resListA {
			// 打印主机名并增加当前列数
			fmt.Fprintf(writer, "%s\t", res.(*model.LinuxConfig).Hostname)
			currentColumns++

			// 如果达到最大列数或者遍历到最后一个主机名，则换行
			if currentColumns == maxColumns || res == resListA[len(resListA)-1] {
				fmt.Fprintln(writer, "")
				currentColumns = 0
			}
		}

		// 刷新写入器，确保所有数据都写入缓冲区
		writer.Flush()

		// 将格式化后的输出写入到标准输出
		fmt.Fprintln((*sess), buffer.String())
		return nil
	}

	// 更新 Linux 配置
	linuxRes := resListA[0].(*model.LinuxConfig)
	sshd.NewTerminal(sess, linuxRes.IPv4Pub, linuxRes.Port, linuxRes.Username, linuxRes.PrivateKey, constants.ResourceTypeLinux)
	return nil
}

// 确定搜索类型和资源
func determineSearchType(resA string) (string, string) {
	searchType := "hostname"
	if strings.Contains(resA, ":") {
		parts := strings.Split(resA, ":")
		if IsIP(parts[0]) {
			searchType = "ipport"
		} else if IsDomain(parts[0]) {
			searchType = "domainport"
		}
	} else {
		if IsIP(resA) {
			searchType = "ip"
		} else if IsDomain(resA) {
			searchType = "domain"
		}
	}
	return searchType, resA
}

// 匹配资源
func matchResource(linuxRes *model.LinuxConfig, searchType, resA string) bool {
	switch searchType {
	case "ip":
		return linuxRes.IPv4Priv == resA || linuxRes.IPv4Pub == resA || linuxRes.IPv6 == resA
	case "domain":
		return linuxRes.IPv4Pub == resA || linuxRes.IPv6 == resA
	case "domainport":
		parts := strings.Split(resA, ":")
		return (linuxRes.IPv4Pub == parts[0] && strconv.Itoa(linuxRes.Port) == parts[1]) ||
			(linuxRes.IPv6 == parts[0] && strconv.Itoa(linuxRes.PortIPv6) == parts[1])
	case "ipport":
		parts := strings.Split(resA, ":")
		return (linuxRes.IPv4Pub == parts[0] && strconv.Itoa(linuxRes.Port) == parts[1]) ||
			(linuxRes.IPv4Priv == parts[0] && strconv.Itoa(linuxRes.PortActual) == parts[1]) ||
			(linuxRes.IPv6 == parts[0] && strconv.Itoa(linuxRes.PortIPv6) == parts[1])
	case "hostname":
		return strings.Contains(linuxRes.Hostname, resA)
	default:
		return false
	}
}

// 检查是否是合法的IP地址
func IsIP(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil
}

// 检查是否是合法的IP:port地址
func IsIPPort(str string) bool {
	parts := strings.Split(str, ":")
	if len(parts) != 2 {
		return false
	}
	ip := net.ParseIP(parts[0])
	if ip == nil {
		return false
	}
	_, err := net.LookupPort("tcp", parts[1])

	return err == nil
}

// 检查是否是合法的域名
func IsDomain(hostname string) bool {
	// 使用net.ParseIP尝试解析，如果解析成功说明不是域名
	ip := net.ParseIP(hostname)
	if ip != nil {
		return false
	}

	// 使用net.LookupHost尝试解析，如果解析失败说明不是域名
	_, err := net.LookupHost(hostname)
	return err == nil
}

func IsDomainPort(hostname string) bool {
	// 使用net.ParseIP尝试解析，如果解析成功说明不是域名
	ip := net.ParseIP(hostname)
	if ip != nil {
		return false
	}

	// 使用net.LookupHost尝试解析，如果解析失败说明不是域名
	_, err := net.LookupHost(hostname)
	return err == nil
}

// Help 返回 ln 命令的帮助信息
func (u *Ln) Usage() string {
	// 获取所有资源类型
	resourceTypes := constants.GetResourceType()
	helpItems := []string{
		"ln <type> <resource> -  login to the specified type of resource",
		" <resource> -  login to the resource simple way;but resource name not equel (" + strings.Join(resourceTypes, ",") + ")",
	}

	// 如果资源类型列表为空，直接返回帮助信息
	if len(resourceTypes) == 0 {
		return strings.Join(helpItems, "\n")
	}
	// 将资源类型添加到帮助信息中
	for _, resType := range resourceTypes {
		helpItems = append(helpItems, fmt.Sprintf("    %s\t: Login to %s 's resource", resType, resType))
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

func (u *Ln) Name() string {
	return "ln"
}
func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Ln{}, Weight: 50})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Ln{}, Weight: 50})
}
