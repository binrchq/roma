package itface

// Helper 定义了命令帮助信息的接口
type Command interface {
	Name() string
}

type CommandWeight struct {
	Command
	Weight int
}

// Helpers 用于收集所有命令的帮助信息
var Commands []CommandWeight

type ByCommandWeight []CommandWeight

func (h ByCommandWeight) Len() int           { return len(h) }
func (h ByCommandWeight) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h ByCommandWeight) Less(i, j int) bool { return h[i].Weight > h[j].Weight }
