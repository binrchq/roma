package cmds

import (
	"strings"

	"bitrec.ai/roma/core/tui/cmds/itface"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewHistory(), Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewHistory(), Weight: 1})
}

type History struct {
	baseLen int // 基础命令长度
	flags   *Flags
}

func NewHistory() *History {
	return &History{baseLen: 7, flags: &Flags{}}
}

func (cmd *History) Name() string {
	return "history"
}

func (cmd *History) Execute(history []string) string {
	return strings.Join(history, "\n")
}

func (cmd *History) Usage() string {
	usageMsg := cmd.flags.FormatUsageln("📜 %s - Display command history", green(cmd.Name()))
	return usageMsg
}
