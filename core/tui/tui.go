package tui

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"strings"

	"bitrec.ai/roma/core/global"
	"bitrec.ai/roma/core/tui/cmds"
	"github.com/brckubo/ssh"

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

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
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
	// fmt.Fprintf(*ui.sess, "\n%s", output)
	// 刷新 stdout 缓冲区
	// (*ui.sess).CloseWrite()
	// rcode, err := (*ui.sess).Write([]byte(output))
	// if err != nil {
	// 	fmt.Println("write error:", err)
	// }
	// fmt.Println("rcode error:", rcode)
	// fmt.Fprintf(*ui.sess, "\n")
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
	log.SetOutput(l.Stderr())
	// sb := screenbuf.New(l)
	for {
		// sb.Reset()
		// sb.Write([]byte(color.WhiteString((*ui.sess).User()) + color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt) + " " + color.CyanString(page) + " \n"))
		// sb.Flush()
		// ui.echo(color.WhiteString((*ui.sess).User()) + color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt) + " " + color.CyanString(page) + " ")
		// ui.echo_e(color.WhiteString((*ui.sess).User()) + color.YellowString(".") + color.GreenString(global.CONFIG.Common.Prompt) + " " + color.CyanString(page) + " ")
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
			res, err := cmds.NewLs(*ui.sess).Execute(line)
			if err != nil {
				ui.echo(color.RedString("%s", err))
				continue
			}
			ui.echo_e(res.(string))
			// case strings.HasPrefix(line, "use"):
			// 	var typeName string
			// 	if len(command) == 1 {
			// 		typeName = "~"
			// 	} else {
			// 		typeName = command[1]
			// 	}
			// 	if typeName == "-h" {
			// 		fmt.Fprintln(*ui.sess, (&cmds.Use{}).Help())
			// 		continue
			// 	}
			// 	t, err := (&cmds.Use{}).Use(typeName)
			// 	if err != nil {
			// 		fmt.Fprintf(*ui.sess, "type not found: %s\n", typeName)
			// 		fmt.Fprintln(*ui.sess, "Type 'use -h' for more information")
			// 		continue
			// 	}
			// 	page = t
			// case strings.HasPrefix(line, "ln"):
			// 	var typeName string
			// 	if len(command) == 1 {
			// 		fmt.Fprintln(*ui.sess, (&cmds.Ln{}).Help())
			// 		continue
			// 	} else {
			// 		typeName = command[1]
			// 	}
			// 	if typeName == "-h" {
			// 		fmt.Fprintln(*ui.sess, (&cmds.Ln{}).Help())
			// 		continue
			// 	}
			// 	err := (&cmds.Ln{}).Execute((ui.sess), command[1:])
			// 	if err != nil {
			// 		fmt.Fprintf((*ui.sess), color.RedString("error: %s\n"), err)
			// 		continue
			// 	}
			// case strings.HasPrefix(line, "quit"):
			// case strings.HasPrefix(line, "exit"):
			// 	err := (&cmds.Exit{}).Exit(*ui.sess)
			// 	if err != nil {
			// 		log.Println("Error closing session:", err)
			// 	}
			// case strings.HasPrefix(line, "help"):
			// 	(&cmds.Help{}).Execute(*ui.sess)
			// case strings.HasPrefix(line, "whoami"):
			// 	(&cmds.Whoami{}).Whoami(*ui.sess)
			// default:
			// 	err := (&cmds.Ln{}).Execute(ui.sess, []string{page, command[0]})
			// 	if err != nil {
			// 		fmt.Fprintf(*ui.sess, color.RedString("error: %s\n"), err)
			// 		continue
			// 	}
			// }
		}
	}
}

// ShowMainMenu show main menu
func (ui *TUI) ShowMainMenu() {
	ui.ShowMenu("Please select the function you need", nil, "Quit", nil)
}
