package cmds

import (
	"strings"

	"bitrec.ai/roma/core/tui/cmds/itface"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewGrep(), Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewGrep(), Weight: 1})
}

type Grep struct {
	baseLen int // 基础命令长度
	flags   *Flags
}

func NewGrep() *Grep {
	return &Grep{baseLen: 4, flags: &Flags{}}
}

func (cmd *Grep) Name() string {
	return "grep"
}

func (cmd *Grep) Execute(input, pattern string) string {
	lines := strings.Split(input, "\n")
	matchedLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, pattern) {
			matchedLines = append(matchedLines, line)
		}
	}
	return strings.Join(matchedLines, "\n")
}

func (cmd *Grep) Usage() string {
	usageMsg := cmd.flags.FormatUsageln("🔍 %s - Search for PATTERN in input", green(cmd.Name()))
	return usageMsg
}
