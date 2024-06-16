package tui

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/tui/cmds"
	"github.com/brckubo/ssh"

	"github.com/fatih/color"
	"github.com/rs/zerolog/log"

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

// ShowMenu show menu
func (ui *TUI) ShowMenu(label string, menu error, BackOptionLabel string, selectedChain error) {
	if global.CONFIG.Banner.Show {
		ui.echo(color.GreenString(global.CONFIG.Banner.Banner))
	}
	page := "~"
	ui.echo((&CReader{}).AllCommandName())

	history := []string{"ls", "cdd", "sds", "whoami"} // 用于存储历史记录的切片

	l, err := readline.NewEx(&readline.Config{
		// Prompt: "",
		// HistoryFile:     "/tmp/readline.tmp",
		Prompt:              color.WhiteString((*ui.sess).User()) + color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt) + " " + color.CyanString(page) + " ",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
		Stdin:               *ui.sess,
		Stdout:              *ui.sess,
	})
	for _, h := range history {
		l.SaveHistory(h)
	}
	if err != nil {
		panic(err)
	}
	defer l.Close()
	l.CaptureExitSignal()
	// sb := screenbuf.New(l)
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		cmd := strings.Split(line, " ")[0]
		switch {
		case cmd == "ls":
			log.Debug().Msgf("ls %s", page)
			res, err := cmds.NewLs(*ui.sess, page).Execute(line)
			if err != nil {
				ui.echo(color.RedString("%s", err))
				continue
			}
			ui.echo_e(res.(string))
		case cmd == "use":
			fold, err := cmds.NewUse().Execute(line)
			if err != nil {
				ui.echo(color.RedString("%s", err))
				continue
			}
			page = fold
			ui.SetPrompt(l, page)
		case cmd == "quit":
		case cmd == "exit":
			err := cmds.NewExit().Exit(*ui.sess)
			if err != nil {

			}
		case cmd == "help":
			cmds.NewHelp().Execute(*ui.sess)
		case cmd == "whoami":
			cmds.NewWhoami().Whoami(*ui.sess)
		case cmd == "ln":
		default:
			if cmd != "" {
				res, err := cmds.NewLn(*ui.sess, page).Execute(line)
				if err != nil {
					ui.echo(color.RedString("%s", err))
					continue
				}
				ui.echo_e(res.(string))
			}
		}
		l.SaveHistory(line)
	}
}

func (ui *TUI) SetPrompt(l *readline.Instance, prompt string) {
	//根据prompt打印不同的ascii
	rand.Seed(time.Now().UnixNano()) // 初始化随机数种子
	// 根据 prompt 打印不同的 ASCII 艺术字符
	if asciiSlice, exists := constants.AsciiPrompts[prompt]; exists {
		randomIndex := rand.Intn(len(asciiSlice)) // 随机选择一个 ASCII 艺术字符
		ui.echo_e(color.GreenString(asciiSlice[randomIndex]))
	}
	l.SetPrompt(color.WhiteString((*ui.sess).User()) + color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt) + " " + color.CyanString(prompt) + " ")
}

// ShowMainMenu show main menu
func (ui *TUI) ShowMainMenu() {
	ui.ShowMenu("Please select the function you need", nil, "Quit", nil)
}
