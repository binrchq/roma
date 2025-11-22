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

var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("login"),
	readline.PcItem("say",
		readline.PcItemDynamic(listFiles("./"),
			readline.PcItem("with",
				readline.PcItem("following"),
				readline.PcItem("items"),
			),
		),
		readline.PcItem("hello"),
		readline.PcItem("bye"),
	),
	readline.PcItem("setprompt"),
	readline.PcItem("setpassword"),
	readline.PcItem("bye"),
	readline.PcItem("help"),
	readline.PcItem("go",
		readline.PcItem("build", readline.PcItem("-o"), readline.PcItem("-v")),
		readline.PcItem("install",
			readline.PcItem("-v"),
			readline.PcItem("-vv"),
			readline.PcItem("-vvv"),
		),
		readline.PcItem("test"),
	),
	readline.PcItem("sleep"),
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
	l, err := readline.NewEx(&readline.Config{
		Prompt:              "",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		Stdin:               *ui.sess,
		Stdout:              *ui.sess,
	})
	if err != nil {
		panic(err)
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
	l.SetPrompt(ui.setPrompt(page))
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
		if newPage != "" {
			page = newPage
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

func (ui *TUI) setPrompt(prompt string) string {
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
	user := color.WhiteString((*ui.sess).User())
	promptPrefix := color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt)
	promptSuffix := color.CyanString(prompt)
	return user + promptPrefix + " " + promptSuffix + " "
}

func (ui *TUI) ShowMainMenu(remainingCmd string, remainingArgs []string) {
	ui.ShowMenu(remainingCmd, remainingArgs)
}
