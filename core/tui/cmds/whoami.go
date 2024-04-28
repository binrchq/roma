package cmds

import (
	"fmt"
	"text/tabwriter"

	"bitrec.ai/roma/core/operation"
	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

type Whoami struct{}

func (c *Whoami) Whoami(sess ssh.Session) {
	op := operation.NewUserOperation()
	userInfo, err := op.GetUserByUsername(sess.User())
	if err != nil {
		fmt.Fprintln(sess, "Error:", err)
		return
	}

	// 使用 tabwriter 创建一个新的 tabwriter.Writer
	// 设置 tabwriter 格式参数，以制作一个漂亮的表格
	// 这里的 \t 表示 tab，- 表示左对齐，4 表示列之间的间隔
	w := tabwriter.NewWriter(sess, 0, 0, 4, ' ', 0) // 去掉 AlignRight 参数

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
}

func (w *Whoami) Usage() string {
	return "whoami - Get user information"
}

// Name 返回命令名称
func (w *Whoami) Name() string {
	return "whoami"
}

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Whoami{}, Weight: 10})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Whoami{}, Weight: 10})
}
