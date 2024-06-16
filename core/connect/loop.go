package connect

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"text/tabwriter"

	"bitrec.ai/roma/core/constants"
	"bitrec.ai/roma/core/model"
	"bitrec.ai/roma/core/sshd"
	"github.com/brckubo/ssh"
)

func NewConnectionLoop(sess *ssh.Session, resModel model.Resource, resType string) error {
	// 将 r 转换为相应的资源类型并创建资源
	ConnectionLoop := resModel.GetConnect()
	if ConnectionLoop == nil {
		return errors.New("缺少连接方式")
	}
	ssh_flag := false
	sccess_flag := false
	var buffer bytes.Buffer
	tw := tabwriter.NewWriter(&buffer, 0, 0, 2, ' ', 0)
	for _, connection := range ConnectionLoop {
		if connection.Type == constants.ConnectSSH {
			ssh_flag = true
			if connection.Host == "" || connection.Port == 0 {
				continue
			}
			err := sshd.NewTerminal(sess, connection.Host, connection.Port, connection.Username, connection.PrivateKey, resType)
			if err == nil {
				sccess_flag = true
				break
			}
		} else if connection.Type == constants.ConnectHTTP {
			fmt.Fprintf(tw, "%s:  %s %s\t%s\n", connection.Type, connection.Host+":"+strconv.Itoa(connection.Port), connection.Username, connection.Password)
		} else if connection.Type == constants.ConnectRDP {
			fmt.Fprintf(tw, "%s:  %s %s\t%s\n", connection.Type, connection.Host+":"+strconv.Itoa(connection.Port), connection.Username, connection.Password)
		} else if connection.Type == constants.ConnectVNC {
			fmt.Fprintf(tw, "%s:  %s %s\t%s\n", connection.Type, connection.Host+":"+strconv.Itoa(connection.Port), connection.Username, connection.Password)
		} else if connection.Type == constants.ConnectDatabase {
			fmt.Fprintf(tw, "%s:  %s %s\t%s\n", connection.Type, connection.Host+":"+strconv.Itoa(connection.Port), connection.Username, connection.Password)
		}
	}
	if ssh_flag && !sccess_flag {
		return errors.New("SSH连接失败")
	} else {
		tw.Flush()
		fmt.Fprint(*sess, buffer.String())
	}
	return nil
}
