package cmds

import (
	"binrc.com/roma/core/tui/cmds/itface"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewClear(), Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewClear(), Weight: 1})
}

type Clear struct {
	baseLen int // 基础命令长度
	flags   *Flags
}

func NewClear() *Clear {
	return &Clear{baseLen: 5, flags: &Flags{}}
}

func (cmd *Clear) Name() string {
	return "clear"
}

func (cmd *Clear) Execute() string {
	return "\033[H\033[2J"
}

func (cmd *Clear) Usage() string {
	usageMsg := cmd.flags.FormatUsageln("🧹 %s - Clear the screen", green(cmd.Name()))
	return usageMsg
}
