package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"text/tabwriter"

	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Whoami{}, Weight: 10})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Whoami{}, Weight: 10})
}

type Whoami struct {
	baseLen int // 基础命令长度
	flags   *Flags
}

func NewWhoami() *Whoami {
	return &Whoami{
		baseLen: 7,
		flags:   &Flags{},
	}
}

// Name 返回命令名称
func (cmd *Whoami) Name() string {
	return "whoami"
}

func (cmd *Whoami) Whoami(sess ssh.Session) (string, error) {
	op := operation.NewUserOperation()
	userInfo, err := op.GetUserByUsername(sess.User())
	if err != nil {
		return "", errors.New("获取用户信息失败" + sess.User())
	}

	// 使用 tabwriter 创建一个新的 tabwriter.Writer

	var buffer bytes.Buffer
	w := tabwriter.NewWriter(&buffer, 0, 0, 4, ' ', 0) // 去掉 AlignRight 参数

	// 将用户信息以表格形式写入到 tabwriter.Writer
	fmt.Fprintf(w, "Username\t:%s\n", userInfo.Username)
	fmt.Fprintf(w, "Name\t:%s\n", userInfo.Name)
	fmt.Fprintf(w, "Nickname\t:%s\n", userInfo.Nickname)
	fmt.Fprintf(w, "Email\t:%s\n", userInfo.Email)
	fmt.Fprintf(w, "PublicKey\t:%s********************\n", userInfo.PublicKey[:22])
	fmt.Fprintf(w, "CreatedAt\t:%s\n", userInfo.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "UpdatedAt\t:%s\n", userInfo.UpdatedAt.Format("2006-01-02 15:04:05"))

	// 输出用户角色信息
	if len(userInfo.Roles) > 0 {
		for _, role := range userInfo.Roles {
			fmt.Fprintf(w, "Role \t:%s - %s\n", role.Name, role.Desc)
			// 输出其他角色属性...
		}
	} else {
		fmt.Fprintln(w, "User has no roles assigned.")
	}

	// 刷新 tabwriter.Writer，以便将缓冲区中的数据输出到 ssh.Session
	w.Flush()
	return buffer.String(), nil
}

func (cmd *Whoami) Usage() string {
	usageMsg := cmd.flags.FormatUsageln("🍂 %s - Get user(me) information", green(cmd.Name()))
	return usageMsg
}
