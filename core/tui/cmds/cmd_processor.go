package cmds

type CommandProcessor struct {
	args []string
}

func (cp *CommandProcessor) shift(n ...int) {
	shiftBy := 1 // 默认移动一个位置
	if len(n) > 0 {
		shiftBy = n[0]
	}
	if len(cp.args) > shiftBy {
		cp.args = cp.args[shiftBy:]
	} else {
		cp.args = []string{}
	}
}
