package itface

// Helper 定义了命令帮助信息的接口
type Helper interface {
	Usage() string
}

type HelperWeight struct {
	Helper
	Weight int
}

// Helpers 用于收集所有命令的帮助信息
var Helpers []HelperWeight

type ByWeight []HelperWeight

func (h ByWeight) Len() int           { return len(h) }
func (h ByWeight) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h ByWeight) Less(i, j int) bool { return h[i].Weight > h[j].Weight }
