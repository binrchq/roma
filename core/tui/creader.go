package tui

import (
	"sort"
	"strings"

	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

type CReader struct{}

func (h *CReader) AllCommandName() string {
	var builder strings.Builder

	sort.Sort(itface.ByCommandWeight(itface.Commands))
	builder.WriteString("commands: ")

	for _, c := range itface.Commands {
		builder.WriteString(c.Name())
		builder.WriteString(" ")
	}

	builder.WriteString("\n")

	return builder.String()
}

func (h *CReader) AllCommandCompleter(sess ssh.Session) {

}
