package cmds

import (
	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewExit(), Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewExit(), Weight: 1})
}

func (cmd *Exit) Name() string {
	return "exit"
}

type Exit struct {
	baseLen int // 基础命令长度
	flags   *Flags
}

func NewExit() *Exit {
	return &Exit{baseLen: 4, flags: &Flags{}}
}

func (cmd *Exit) Exit(sess ssh.Session) error {
	if sess != nil {
		err := sess.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
func (cmd *Exit) Usage() string {
	usageMsg := cmd.flags.FormatUsageln("🍂 %s - Exit the program", green(cmd.Name()))
	return usageMsg
}
