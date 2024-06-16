package cmds

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/fatih/color"
)

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
)

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

// 根据选项的名称获取对应的选项
func (f *Flags) GetOptionValue(name string) interface{} {
	for _, opt := range f.Options {
		if opt.Short == name || opt.Long == name {
			return opt.Value
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

// Parse method to process command line arguments and set the options accordingly
func (f *Flags) Parse(argsI interface{}) string {
	var args []string
	switch argsT := argsI.(type) {
	case []string:
		args = argsT
	case string:
		args = strings.Split(argsT, " ")
	default:
		args = []string{}
	}

	target := ""
	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			if strings.HasPrefix(arg, "--") {
				// Long option
				arg = strings.TrimPrefix(arg, "--")
				opt := f.GetOption(arg)
				if opt != nil {
					switch opt.Type {
					case BoolOption:
						opt.Value = true
						opt.IsSet = true
					case StringOption:
						if i+1 < len(args) {
							opt.Value = args[i+1]
							opt.IsSet = true
							i++
						}
					case ListOption:
						var list []string
						for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
							list = append(list, args[i+1])
							i++
						}
						opt.Value = list
						opt.IsSet = true
					}
				}
			} else {
				// Short option(s)
				arg = strings.TrimPrefix(arg, "-")
				for _, char := range arg {
					opt := f.GetOption(string(char))
					if opt != nil {
						switch opt.Type {
						case BoolOption:
							opt.Value = true
							opt.IsSet = true
						case StringOption:
							if i+1 < len(args) {
								opt.Value = args[i+1]
								opt.IsSet = true
								i++
							}
						case ListOption:
							var list []string
							for i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
								list = append(list, args[i+1])
								i++
							}
							opt.Value = list
							opt.IsSet = true
						}
					}
				}
			}
		} else {
			target = arg
		}
		i++
	}
	return target
}

func (f *Flags) FormatUsagef(format string, args ...interface{}) string {
	return fmt.Sprintf(format+"\n", args...)
}

func (f *Flags) FormatUsageln(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

// 返回渲染的usage
func (f *Flags) ColorUsage(tw *tabwriter.Writer) *tabwriter.Writer {
	// 设置颜色
	for _, opt := range f.Options {
		switch opt.Type {
		case BoolOption:
			fmt.Fprintf(tw, "  %s, %s\t%s\n", cyan("-"+opt.Short), green("--"+opt.Long), yellow(opt.Help))
		case StringOption:
			fmt.Fprintf(tw, "  %s, %s\t%s\n", cyan("-"+opt.Short), green("--"+opt.Long+"="+opt.GetNameToUpper()), yellow(opt.Help))
		case ListOption:
			fmt.Fprintf(tw, "  %s, %s\t%s\n", cyan("-"+opt.Short), green("--"+opt.Long+"=[]"+opt.GetNameToUpper()), yellow(opt.Help))
		}
	}
	return tw
}
