package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"

	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
)

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: NewAwk(), Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: NewAwk(), Weight: 1})
}

type Awk struct {
	baseLen int // 基础命令长度
	flags   *Flags
}

func NewAwk() *Awk {
	flags := &Flags{}
	flags.AddOption("F", "field-separator", "Specify the field separator", StringOption, " ")
	flags.AddOption("h", "help", "Display this help message", BoolOption, false)
	return &Awk{baseLen: 3, flags: flags}
}

func (cmd *Awk) Name() string {
	return "awk"
}

func (cmd *Awk) Execute(input string, args []string) (string, error) {
	if len(args) < 2 {
		return "", errors.New("awk requires a pattern and an action to perform")
	}

	// 获取字段分隔符
	fs := cmd.flags.GetOptionValue("F").(string)

	// awk 命令的模式部分和动作部分
	pattern := args[0]
	action := strings.Join(args[1:], " ")

	// 根据输入内容的不同，执行不同的操作
	var filteredLines []string
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if strings.Contains(line, pattern) {
			// 将匹配的行按指定分隔符分割为字段
			fields := strings.Split(line, fs)
			// 执行指定的 awk 动作（这里只支持简单的打印字段动作）
			result, err := evalAwkAction(fields, action)
			if err != nil {
				return "", err
			}
			filteredLines = append(filteredLines, result)
		}
	}

	output := strings.Join(filteredLines, "\n")
	return highlightMatches(output, pattern), nil
}

// Function to evaluate awk actions
func evalAwkAction(fields []string, action string) (string, error) {
	if action == "{print $1}" && len(fields) > 0 {
		return fields[0], nil
	} else if action == "{print $2}" && len(fields) > 1 {
		return fields[1], nil
	} else if strings.HasPrefix(action, "{print $") && strings.HasSuffix(action, "}") {
		indexStr := strings.TrimPrefix(strings.TrimSuffix(action, "}"), "{print $")
		index, err := strconv.Atoi(indexStr)
		if err != nil {
			return "", err
		}
		if index > 0 && index <= len(fields) {
			return fields[index-1], nil
		}
	}
	return "", errors.New("unsupported action")
}

// Function to highlight matches in the awk output
func highlightMatches(output, pattern string) string {
	return strings.ReplaceAll(output, pattern, color.YellowString(pattern))
}

func (cmd *Awk) Usage() string {
	usageMsg := cmd.flags.FormatUsagef("🍂 %s", green(cmd.Name()+" [OPTIONS] PATTERN ACTION"))
	usageMsg += cmd.flags.FormatUsagef("Process the input text according to the specified PATTERN and ACTION.")
	usageMsg += cmd.flags.FormatUsagef("Usage:")

	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	// 写入Options
	log.Info().Msgf("flags: %v", cmd.flags.Options)
	tw = cmd.flags.ColorUsage(tw)
	fmt.Fprint(tw) // 换行
	tw.Flush()

	return usageMsg + buffer.String()
}
