package cmds

import "strings"

type OptionType int

const (
	BoolOption   OptionType = iota // 布尔选项，不带值
	StringOption                   // 字符串选项，带一个值
	ListOption                     // 列表选项，带多个值
)

type Option struct {
	Short string
	Long  string
	Type  OptionType
	IsSet bool
	Help  string
	Value interface{}
}

// 获取参数大写
func (opt *Option) GetNameToUpper() string {
	if opt != nil {
		return strings.ToUpper(opt.Long)
	}
	return ""
}

type Flags struct {
	Options []*Option
}

// 添加一个选项到 Flags
func (f *Flags) AddOption(short, long, help string, optType OptionType, value interface{}) {
	option := &Option{
		Short: short,
		Long:  long,
		Type:  optType,
		IsSet: false,
		Help:  help,
		Value: value,
	}
	f.Options = append(f.Options, option)
}

// 根据选项的名称获取对应的选项
func (f *Flags) GetOption(name string) *Option {
	for _, opt := range f.Options {
		if opt.Short == name || opt.Long == name {
			return opt
		}
	}
	return nil
}

// 根据选项的名称设置对应的值
func (f *Flags) SetOptionValue(name string, value interface{}) {
	opt := f.GetOption(name)
	if opt != nil {
		opt.Value = value
	}
}
