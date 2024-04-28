// package cmds

// import (
// 	"fmt"

// 	"github.com/brckubo/ssh"
// )

//	func Help(sess ssh.Session) {
//		fmt.Fprintln(sess, "Available commands:")
//		fmt.Fprintln(sess, "<resource> - jump to the resource page")
//		fmt.Fprintln(sess, "ls - List all resources")
//		fmt.Fprintln(sess, "cd <type> - Change the current resource")
//		fmt.Fprintln(sess, "use <type> - Change the current resource")
//		fmt.Fprintln(sess, "to <type> <IP:Port|Hostname> - Change the current resource")
//		fmt.Fprintln(sess, "exit - Exit the program")
//	}
package cmds

import (
	"fmt"
	"sort"

	"bitrec.ai/roma/core/tui/cmds/itface"
	"github.com/brckubo/ssh"
)

type Help struct{}

func (h *Help) Execute(sess ssh.Session) {
	// 将帮助信息列表按照权重进行排序
	sort.Sort(itface.ByWeight(itface.Helpers))
	for _, h := range itface.Helpers {
		fmt.Fprintln(sess, h.Usage())
	}
}

func (h *Help) Usage() string {
	return "help - Gets more help messages for commands"
}

func (e *Help) Name() string {
	return "help"
}

func init() {
	itface.Helpers = append(itface.Helpers, itface.HelperWeight{Helper: &Help{}, Weight: 1})
	itface.Commands = append(itface.Commands, itface.CommandWeight{Command: &Help{}, Weight: 1})
}
