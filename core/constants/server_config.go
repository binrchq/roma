package constants

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	DOMAIN = "https://roma.bitrec.ai"
)

var (
	BASE_DIR = "/usr/local/roma"
)

func init() {
	root := os.Getenv("ROMA_BASE_ROOT")
	if root != "" {
		BASE_DIR = root
	} else {
		// 执行`go list -m -f "{{.Dir}}"`命令来获取当前模块的根目录
		cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
		output, err := cmd.Output()
		if err != nil {
			//日志
			log.Println("获取当前模块的根目录失败:", err)
			return
		}

		// 将输出转换为字符串并去除可能的空白字符
		BASE_DIR = strings.TrimSpace(string(output))
	}

}
