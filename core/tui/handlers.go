package tui

import (
	"errors"
	"fmt"
	"strings"

	"binrc.com/roma/core/tui/cmds"
	"github.com/chzyer/readline"
	"github.com/fatih/color"
	"github.com/loganchef/ssh"
	"github.com/rs/zerolog/log"
)

// Function to handle the "ls" command
func handleLs(ui *TUI, args []string, page, input string) (string, error) {
	log.Debug().Msgf("ls %s", page)
	res, err := cmds.NewLs(*ui.sess, page).Execute(strings.Join(args, " "))
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

// Function to handle the "use" command
func handleUse(ui *TUI, l *readline.Instance, args []string, input string) (string, error) {
	fold, err := cmds.NewUse().Execute(strings.Join(args, " "))
	if err != nil {
		return "~", err
	}
	// 切换页面时不显示 banner，只更新提示符
	l.SetPrompt(ui.setPrompt(fold, false))
	return fold, nil
}

// Function to handle the "exit" command
func handleExit(ui *TUI, sess ssh.Session) bool {
	err := cmds.NewExit().Exit(sess)
	if err != nil {
		ui.echo(color.RedString("%s", err))
		return false
	}
	return true
}

// Function to handle the "clear" command
func handleClear(ui *TUI) {
	ui.echo_e(cmds.NewClear().Execute())
}

// Function to handle the "history" command
func handleHistory(ui *TUI, history []string) (string, error) {
	output := cmds.NewHistory().Execute(history)
	return output, nil
}

// Function to handle the "help" command
func handleHelp(ui *TUI, args []string, input string) (string, error) {
	output, err := cmds.NewHelp().Execute(*ui.sess)
	if err != nil {
		return "", err
	}
	return output, nil
}

// Function to handle the "whoami" command
func handleWhoami(ui *TUI, args []string, input string) (string, error) {
	output, err := cmds.NewWhoami().Whoami(*ui.sess)
	if err != nil {
		return "", err
	}
	return output, nil
}

// getNonEmptyLines 返回非空行的切片
func getNonEmptyLines(input string) []string {
	var lines []string
	for _, line := range strings.Split(input, "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// handleGrep 函数用于在输入文本中进行模式匹配和高亮
func handleGrep(ui *TUI, input string, args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("grep requires a pattern to search")
	}

	pattern := args[0]

	// 根据输入内容的不同，执行不同的搜索

	if strings.Contains(input, "\n") || len(getNonEmptyLines(input)) > 1 {
		// 如果输入包含换行符，则将其视为多行文本进行 grep
		var filteredLines []string
		lines := strings.Split(input, "\n")
		fmt.Print(lines)
		for _, line := range lines {
			if strings.Contains(line, pattern) {
				// 高亮匹配的部分
				highlightedLine := strings.ReplaceAll(line, pattern, color.YellowString(pattern))
				filteredLines = append(filteredLines, highlightedLine)
			}
		}
		return strings.Join(filteredLines, "\n"), nil
	}

	fmt.Println(input)
	// 否则视为单行文本,则先分割，再显示风格之后能匹配的字段
	if strings.Contains(input, pattern) {
		// 高亮匹配的部分
		inputParts := strings.Split(input, " ")
		patternInputParts := []string{}
		for _, part := range inputParts {
			if part != "" && strings.Contains(part, pattern) {
				patternInputParts = append(patternInputParts, color.YellowString(part))
			}
		}
		highlightedOutput := strings.ReplaceAll(strings.Join(patternInputParts, " "), pattern, color.YellowString(pattern))
		return highlightedOutput, nil
	}

	return "", errors.New("pattern not found in input")
}

// handleAwk 函数用于在输入文本中进行 awk 操作
func handleAwk(ui *TUI, input string, args []string) (string, error) {
	if len(args) < 1 {
		return "", errors.New("awk requires a pattern and an action to perform")
	}

	// awk 命令的模式部分和动作部分
	pattern := args[0]
	action := strings.Join(args[1:], " ")

	// 根据输入内容的不同，执行不同的操作
	var filteredLines []string
	lines := strings.Split(input, "\n")
	for _, line := range lines {
		if strings.Contains(line, pattern) {
			// 将匹配的行按空格分割为字段
			fields := strings.Fields(line)
			// 执行指定的 awk 动作（这里只支持简单的打印字段动作）
			if action == "{print $1}" && len(fields) > 0 {
				filteredLines = append(filteredLines, fields[0])
			} else if action == "{print $2}" && len(fields) > 1 {
				filteredLines = append(filteredLines, fields[1])
			} else {
				return "", errors.New("unsupported action")
			}
		}
	}

	return strings.Join(filteredLines, "\n"), nil
}

// Function to handle default commands
func handleDefault(ui *TUI, cmd string, page, input string) (string, error) {
	res, err := cmds.NewLn(*ui.sess, page).Execute(cmd)
	if err != nil {
		return "", err
	}
	return res.(string), nil
}

// Function to highlight matches in the grep output
func highlightMatches(output, pattern string) string {
	highlighted := strings.ReplaceAll(output, pattern, color.YellowString(pattern))
	return highlighted
}
