package tui

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/global"
	"binrc.com/roma/core/tui/cmds"
	"github.com/loganchef/ssh"

	"github.com/fatih/color"

	// "github.com/manifoldco/promptui"
	"github.com/chzyer/readline"
)

// TUI pui
type TUI struct {
	sess *ssh.Session
}

// SetSession SetSession
func (ui *TUI) SetSession(s *ssh.Session) {
	ui.sess = s
}

// Function constructor - constructs new function for listing given directory
func listFiles(path string) func(string) []string {
	return func(line string) []string {
		names := make([]string, 0)
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			names = append(names, f.Name())
		}
		return names
	}
}

// 动态获取资源类型列表用于自动完成
func getResourceTypes() []string {
	return constants.GetResourceType()
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("ls",
		readline.PcItem("-h", readline.PcItem("--help")),
		readline.PcItem("-a", readline.PcItem("--all")),
		readline.PcItem("-l", readline.PcItem("--list")),
		readline.PcItemDynamic(func(line string) []string {
			return getResourceTypes()
		}),
	),
	readline.PcItem("use",
		readline.PcItem("-h", readline.PcItem("--help")),
		readline.PcItemDynamic(func(line string) []string {
			types := getResourceTypes()
			types = append(types, "~")
			return types
		}),
	),
	readline.PcItem("help"),
	readline.PcItem("whoami"),
	readline.PcItem("clear"),
	readline.PcItem("history"),
	readline.PcItem("grep"),
	readline.PcItem("awk"),
	readline.PcItem("quit"),
	readline.PcItem("exit"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func (ui *TUI) echo(output string) {
	fmt.Fprintln(*ui.sess, output)
}

func (ui *TUI) echo_e(output string) {
	fmt.Fprint(*ui.sess, output)
}

func (ui *TUI) ShowMenu(remainingCmd string, remainingArgs []string) {
	page := "~"

	// 强制启用颜色输出（在 Docker 容器中也需要颜色）
	// fatih/color 库默认会检测是否为 TTY，在 Docker 中可能被禁用
	color.NoColor = false

	// 获取当前用户名
	username := (*ui.sess).User()

	// 创建历史管理器
	historyManager := NewHistoryManager(username)

	// 从文件加载历史记录
	history := historyManager.LoadHistory()

	// 如果没有历史记录，使用默认的
	if len(history) == 0 {
		history = []string{"ls", "cdd", "whoami"}
	}

	// Initialize readline configuration
	// 在 Docker 容器中，readline 可能检测不到终端特性
	// 关键：确保 Stderr 被正确设置，readline 使用 Stderr 显示提示符和输入字符
	// 如果 Stderr 未设置或设置不正确，提示符和输入字符将不会显示
	l, err := readline.NewEx(&readline.Config{
		Prompt:              "",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		Stdin:               *ui.sess,
		Stdout:              *ui.sess,
		Stderr:              *ui.sess, // readline 使用 Stderr 显示提示符和输入，这在 Docker 容器中尤其重要
	})
	if err != nil {
		// 在 Docker 容器中，如果 readline 初始化失败，输出错误信息并返回
		// 而不是 panic，这样可以避免会话立即关闭
		fmt.Fprintf(*ui.sess, "错误: 无法初始化终端: %v\n", err)
		fmt.Fprintf(*ui.sess, "提示: 请确保使用 -t 参数连接 SSH (ssh -t user@host)\n")
		return
	}
	defer l.Close()

	// Capture exit signals
	l.CaptureExitSignal()

	// Save initial history to readline
	for _, h := range history {
		l.SaveHistory(h)
	}
	// 如果传递了命令，则直接执行该命令并退出

	if remainingCmd != "" {
		fmt.Println("执行命令:", strings.Join(append([]string{remainingCmd}, remainingArgs...), " "))
		// 执行命令并返回结果
		line := strings.Join(append([]string{remainingCmd}, remainingArgs...), " ")
		output, _, lastErr := ui.executeCommand(l, line, page, history)
		if lastErr != nil {
			ui.echo_e(lastErr.Error())
		} else {
			ui.echo_e(output)
		}
		// 保存命令到历史文件
		if line != "" {
			if err := historyManager.AppendHistory(line); err != nil {
				// 记录错误但不中断执行
				fmt.Fprintf(*ui.sess, "Warning: failed to save history: %v\n", err)
			}
		}
		return
	}
	// Main loop to read and process input
	// 只在首次设置提示符时显示 banner 和命令列表
	promptStr := ui.setPrompt(page, true)
	l.SetPrompt(promptStr)
	// 在 Docker 容器中，刷新以确保提示符立即显示
	l.Refresh()
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			}
			continue
		} else if err == io.EOF {
			break
		}

		// Process the input command
		line = strings.TrimSpace(line)
		output, newPage, lastErr := ui.executeCommand(l, line, page, history)
		// Update page if it changed
		if newPage != "" && newPage != page {
			page = newPage
			// 页面改变时更新提示符（但不重复显示 banner）
			promptStr = ui.setPrompt(page, false)
			l.SetPrompt(promptStr)
			l.Refresh() // 刷新 readline 以显示新的提示符
		}
		// Save the command in history
		if line != "" {
			history = append(history, line)
			l.SaveHistory(line)
			// 立即同步到文件
			if err := historyManager.AppendHistory(line); err != nil {
				// 记录错误但不中断执行
				fmt.Fprintf(*ui.sess, "Warning: failed to save history: %v\n", err)
			}
		}
		// Finally, print the output of the last command in the chain if there was no error
		if lastErr == nil && output != "" {
			ui.echo(output)
		}
	}
}

func (ui *TUI) executeCommand(l *readline.Instance, cmd string, page string, history []string) (string, string, error) {

	var output string
	var lastErr error
	var previousOutput string
	cmdParts := strings.Split(cmd, " | ") // Support pipes
	for _, cmd := range cmdParts {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}
		args := strings.Split(cmd, " ")

		// Reset output and lastErr for each command
		output = ""
		lastErr = nil

		switch args[0] {
		case "ls":
			output, lastErr = handleLs(ui, args, page, previousOutput)
		case "use":
			page, lastErr = handleUse(ui, l, args, previousOutput)
		case "quit", "exit":
			handleExit(ui, *ui.sess)
		case "help":
			output, lastErr = handleHelp(ui, args, previousOutput)
		case "whoami":
			output, lastErr = handleWhoami(ui, args, previousOutput)
		case "clear":
			handleClear(ui)
		case "history":
			output, lastErr = handleHistory(ui, history)
		case "awk":
			output, lastErr = cmds.NewAwk().Execute(output, args[1:])
		case "grep":
			if previousOutput == "" {
				previousOutput, lastErr = handleLs(ui, args, page, previousOutput)
				if lastErr != nil {
					ui.echo(color.RedString("%s", lastErr))
					break
				}
				output, lastErr = handleGrep(ui, previousOutput, args[1:])
			} else {
				output, lastErr = handleGrep(ui, previousOutput, args[1:])
			}
		default:
			output, lastErr = handleDefault(ui, cmd, page, previousOutput)
		}

		// If there is an error, print it and stop the chain
		if lastErr != nil {
			ui.echo(color.RedString("%s", lastErr))
			return "", page, lastErr
		}

		// Update previousOutput to current command's output
		previousOutput = output
	}
	return previousOutput, page, lastErr
}

func (ui *TUI) setPrompt(prompt string, showBanner bool) string {
	// 只在首次显示时输出 banner 和命令列表
	if showBanner {
		if prompt == "~" {
			if global.CONFIG.Banner.Show {
				ui.echo(color.GreenString(global.CONFIG.Banner.Banner))
			}
			ui.echo((&CReader{}).AllCommandName())
		} else {
			rng := rand.New(rand.NewSource(time.Now().UnixNano()))
			if asciiSlice, exists := constants.AsciiPrompts[prompt]; exists {
				randomIndex := rng.Intn(len(asciiSlice)) // Randomly select an ASCII art index
				ui.echo_e(color.GreenString(asciiSlice[randomIndex]))
			}
		}
	}
	user := color.WhiteString((*ui.sess).User())
	promptPrefix := color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt)
	promptSuffix := color.CyanString(prompt)
	return user + promptPrefix + " " + promptSuffix + " "
}

func (ui *TUI) ShowMainMenu(remainingCmd string, remainingArgs []string) {
	ui.ShowMenu(remainingCmd, remainingArgs)
}
