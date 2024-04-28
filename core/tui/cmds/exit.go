package cmds

import (
	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

type Exit struct{}

func (e *Exit) Exit(sess ssh.Session) error {
	if sess != nil {
		err := sess.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Exit) Name() string {
	return "exit"
}

func (e *Exit) Usage() string {
	return "exit - Exit the program"
}

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Exit{}, Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Exit{}, Weight: 1})
}
