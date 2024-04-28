package itface

// Helper 定义了命令帮助信息的接口
type Completer interface {
	Name() string
}

type CompleterWeight struct {
	Command
	Weight int
}

// Helpers 用于收集所有命令的帮助信息
var Completers []CompleterWeight

type ByCompleterWeight []CompleterWeight

func (h ByCompleterWeight) Len() int           { return len(h) }
func (h ByCompleterWeight) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h ByCompleterWeight) Less(i, j int) bool { return h[i].Weight > h[j].Weight }
