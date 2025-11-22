package sshd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"binrc.com/roma/core/constants"
	"binrc.com/roma/core/model"
	"binrc.com/roma/core/operation"
	"binrc.com/roma/core/utils"
	"github.com/loganchef/ssh"
	gossh "golang.org/x/crypto/ssh"
)

const (
	flagCopyFile       = "C"
	flagStartDirectory = "D"
	flagEndDirectory   = "E"
	flagTime           = "T"
)

const (
	responseOk        uint8 = 0
	responseError     uint8 = 1
	responseFailError uint8 = 2
)

type response struct {
	Type    uint8
	Message string
}

// ParseResponse Reads from the given reader (assuming it is the output of the remote) and parses it into a Response structure
func parseResponse(reader io.Reader) (response, error) {
	buffer := make([]uint8, 1)
	_, err := reader.Read(buffer)
	if err != nil {
		return response{}, err
	}

	responseType := buffer[0]
	message := ""
	if responseType > 0 {
		bufferedRader := bufio.NewReader(reader)
		message, err = bufferedRader.ReadString('\n')
		if err != nil {
			return response{}, err
		}
	}

	return response{responseType, message}, nil
}

func (r *response) IsOk() bool {
	return r.Type == responseOk
}

func (r *response) IsError() bool {
	return r.Type == responseError
}

// Returns true when the remote responded with an error
func (r *response) FailError() bool {
	return r.Type == responseFailError
}

// Returns true when the remote answered with a warning or an error
func (r *response) IsFailure() bool {
	return r.Type > 0
}

// Returns the message the remote sent back
func (r *response) GetMessage() string {
	return r.Message
}

// ExecuteSCP ExecuteSCP
func ExecuteSCP(args []string, clientSess *ssh.Session) error {
	flag := args[0]
	switch flag {
	case "-t":
		err := copyToServer(args, clientSess)
		if err != nil {
			replyErr(*clientSess, err)
			return err
		}
	case "-f":
		err := copyFromServer(args, clientSess)
		if err != nil {
			replyErr(*clientSess, err)
			return err
		}
		(*clientSess).Close()
	default:
		return errors.New("This feature is not currently supported")
	}

	return nil
}

func copyToServer(args []string, clientSess *ssh.Session) error {
	err := replyOk(*clientSess)
	if err != nil {
		return err
	}

	bufferedReader := bufio.NewReader(*clientSess)
	b, err := bufferedReader.ReadByte()
	if err != nil {
		return err
	}

	flag := string(b)
	switch flag {
	case flagCopyFile:
		var perm string
		var size int64
		var filename string
		n, err := fmt.Fscanf(bufferedReader, "%s %d %s\n", &perm, &size, &filename)

		if err != nil {
			return err
		}
		if n != 3 {
			return fmt.Errorf("unexpected count in reading start directory message header: n=%d", 3)
		}

		err = copyFileToServer(bufferedReader, size, filename, args[1], perm, clientSess)
		if err != nil {
			return err
		}
		return nil
	case flagEndDirectory:
	case flagStartDirectory:
		return errors.New("Folder transfer is not yet supported. You can try to compress the folder and upload it. ")
	default:
		return fmt.Errorf("expected control record")
	}

	return nil
}

func copyFromServer(args []string, clientSess *ssh.Session) error {
	resource, resourceType, filePath, err := parseResourcePath(args[1], (*clientSess).User())
	if err != nil {
		return err
	}

	// 获取凭证
	passportOp := operation.NewPassportOperation()
	passports, err := passportOp.GetPassportByType(resourceType)
	if err != nil || len(passports) == 0 {
		return fmt.Errorf("no passport found for resource type: %s", resourceType)
	}

	// 获取资源连接信息
	var ip string
	var port int

	switch r := resource.(type) {
	case *model.LinuxConfig:
		ip = r.IPv4Priv
		if ip == "" {
			ip = r.IPv4Pub
		}
		port = r.Port
		if port == 0 {
			port = 22
		}
	case *model.WindowsConfig:
		ip = r.IPv4Priv
		if ip == "" {
			ip = r.IPv4Pub
		}
		port = r.Port
		if port == 0 {
			port = 22
		}
	default:
		return fmt.Errorf("unsupported resource type: %T", resource)
	}

	upstream, err := NewSSHClient(ip, port, passports[0].ServiceUser, passports[0].Passport, resourceType)
	if err != nil {
		return err
	}

	upstreamSess, err := upstream.NewSession()
	if err != nil {
		return err
	}

	errCh := make(chan error, 2)
	defer func() {
		select {
		case <-errCh:
			return
		default:
		}
		close(errCh)
	}()

	stdout, err := upstreamSess.StdoutPipe()
	if err != nil {
		return err
	}

	stdin, err := upstreamSess.StdinPipe()
	if err != nil {
		return err
	}

	err = upstreamSess.Start(fmt.Sprintf("scp -f %s", filePath))
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()

		err := replyOk(stdin)
		if err != nil {
			errCh <- err
			return
		}

		stdOutReader := bufio.NewReader(stdout)
		b, err := stdOutReader.ReadByte()
		if err != nil {
			errCh <- err
			return
		}

		if b == responseError {
			message, err := stdOutReader.ReadString('\n')
			if err != nil {
				errCh <- err
				return
			}
			errCh <- errors.New(message)
			return
		}

		flag := string(b)
		switch flag {
		case flagCopyFile:
			var perm string
			var size int64
			var filename string
			n, err := fmt.Fscanf(stdOutReader, "%s %d %s\n", &perm, &size, &filename)
			if err != nil {
				errCh <- err
				return
			}
			if n != 3 {
				errCh <- fmt.Errorf("unexpected count in reading start directory message header: n=%d", 3)
			}
			err = replyOk(stdin)
			if err != nil {
				errCh <- err
				return
			}
			err = copyFileFromServer(stdOutReader, size, filename, perm, clientSess)
			if err != nil {
				errCh <- err
				return
			}
			return
		case flagEndDirectory:
		case flagStartDirectory:
			errCh <- errors.New("Folder transfer is not yet supported. You can try to compress the folder and upload it. ")
			return
		default:
			errCh <- fmt.Errorf("expected control record")
			return
		}

	}()

	upstreamSess.Wait()

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func copyFileFromServer(bfReader *bufio.Reader, size int64, filename string, perm string, clientSess *ssh.Session) error {
	tmpFilePath, tmp, err := createTmpFile(bfReader, perm, size)
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		if utils.FileExited(tmpFilePath) {
			os.Remove(tmpFilePath)
		}
	}()

	tmpReader := bufio.NewReader(tmp)
	err = copyToClientSession(tmpReader, clientSess, perm, filename, size)
	if err != nil {
		return err
	}

	return nil
}

func copyToClientSession(tmpReader *bufio.Reader, clientSess *ssh.Session, perm, filename string, size int64) error {
	if err := checkResponse(*clientSess); err != nil {
		return err
	}

	_, err := fmt.Fprintln(*clientSess, flagCopyFile+perm, size, filename)
	if err != nil {
		return err
	}

	if err := checkResponse(*clientSess); err != nil {
		return err
	}

	io.Copy(*clientSess, tmpReader)

	_, err = fmt.Fprint(*clientSess, "\x00")
	if err != nil {
		return err
	}

	return nil
}

// parseResourcePath 解析 SCP 路径格式: user@hostname:/remote/path
// 返回资源配置、远程路径和错误
func parseResourcePath(fullPath, currentUsername string) (model.Resource, string, string, error) {
	// 解析格式: user@hostname:/remote/path
	args := strings.SplitN(fullPath, ":", 2)
	invalidPathErr := errors.New(
		"Please input your resource path, like 'scp -P 2222 /tmp/tmp.file user@jumpserver:user@server1:/tmp/tmp.file'",
	)

	if len(args) < 2 {
		return nil, "", "", invalidPathErr
	}

	inputServer, remotePath := args[0], args[1]
	serverArgs := strings.SplitN(inputServer, "@", 2)
	if len(serverArgs) < 2 {
		return nil, "", "", invalidPathErr
	}

	_, hostname := serverArgs[0], serverArgs[1] // serviceUser 暂时未使用，未来可用于权限检查

	// 从数据库查找资源（先尝试 Linux）
	resourceOp := operation.NewResourceOperation()

	// 尝试 Linux 资源
	var resource model.Resource
	var resourceType string

	linuxResource := &model.LinuxConfig{}
	err := resourceOp.DB.Where("hostname = ?", hostname).First(linuxResource).Error
	if err == nil {
		resource = linuxResource
		resourceType = constants.ResourceTypeLinux
	} else {
		// 尝试 Windows 资源
		windowsResource := &model.WindowsConfig{}
		err = resourceOp.DB.Where("hostname = ?", hostname).First(windowsResource).Error
		if err == nil {
			resource = windowsResource
			resourceType = constants.ResourceTypeWindows
		} else {
			return nil, "", "", fmt.Errorf("resource not found: hostname '%s'", hostname)
		}
	}

	// TODO: 检查用户权限
	// userOp := operation.NewUserOperation()
	// user, err := userOp.GetUserByUsername(currentUsername)
	// if err != nil {
	// 	return nil, "", "", fmt.Errorf("user not found: %s", currentUsername)
	// }
	// 检查用户是否有权限访问此资源...

	return resource, resourceType, remotePath, nil
}

func checkResponse(r io.Reader) error {
	response, err := parseResponse(r)
	if err != nil {
		return err
	}

	if response.IsFailure() {
		return errors.New(response.GetMessage())
	}

	return nil

}

func copyFileToServer(bfReader *bufio.Reader, size int64, filename, filePath string, perm string, clientSess *ssh.Session) error {
	resource, resourceType, remotePath, err := parseResourcePath(filePath, (*clientSess).User())
	if err != nil {
		return err
	}
	err = replyOk(*clientSess)
	if err != nil {
		return err
	}

	// 获取凭证
	passportOp := operation.NewPassportOperation()
	passports, err := passportOp.GetPassportByType(resourceType)
	if err != nil || len(passports) == 0 {
		return fmt.Errorf("no passport found for resource type: %s", resourceType)
	}

	// 获取资源连接信息
	var ip string
	var port int

	switch r := resource.(type) {
	case *model.LinuxConfig:
		ip = r.IPv4Priv
		if ip == "" {
			ip = r.IPv4Pub
		}
		port = r.Port
		if port == 0 {
			port = 22
		}
		filePath = remotePath
	case *model.WindowsConfig:
		ip = r.IPv4Priv
		if ip == "" {
			ip = r.IPv4Pub
		}
		port = r.Port
		if port == 0 {
			port = 22
		}
		filePath = remotePath
	default:
		return fmt.Errorf("unsupported resource type: %T", resource)
	}

	upstream, err := NewSSHClient(ip, port, passports[0].ServiceUser, passports[0].Passport, resourceType)
	if err != nil {
		return err
	}

	upstreamSess, err := upstream.NewSession()
	if err != nil {
		return err
	}

	err = copyToUpstreamSession(bfReader, upstreamSess, perm, filePath, filename, size)
	if err != nil {
		return err
	}

	err = replyOk(*clientSess)
	if err != nil {
		return err
	}

	return nil
}

func copyToUpstreamSession(r *bufio.Reader, upstreamSess *gossh.Session, perm, filePath, filename string, size int64) error {
	errCh := make(chan error, 2)
	defer func() {
		select {
		case <-errCh:
			return
		default:
		}
		close(errCh)
	}()
	stdout, err := upstreamSess.StdoutPipe()
	if err != nil {
		return err
	}

	stdin, err := upstreamSess.StdinPipe()
	if err != nil {
		return err
	}

	err = upstreamSess.Start(fmt.Sprintf("scp -t %s", filePath))
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}

		_, err = fmt.Fprintln(stdin, flagCopyFile+perm, size, filename)
		if err != nil {
			errCh <- err
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}

		// Create a temp file
		tmpFilePath, tmp, err := createTmpFile(r, perm, size)
		defer func() {
			tmp.Close()
			if utils.FileExited(tmpFilePath) {
				os.Remove(tmpFilePath)
			}
		}()

		if err != nil {
			errCh <- err
			return
		}
		defer func() {
			tmp.Close()
		}()

		tmpReader := bufio.NewReader(tmp)
		io.Copy(stdin, tmpReader)

		_, err = fmt.Fprint(stdin, "\x00")
		if err != nil {
			errCh <- err
			return
		}

		if err = checkResponse(stdout); err != nil {
			errCh <- err
			return
		}
	}()

	upstreamSess.Wait()

	close(errCh)
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func createTmpFile(r *bufio.Reader, perm string, size int64) (string, *os.File, error) {
	fileMode, err := strconv.ParseUint(perm, 8, 0)
	if err != nil {
		return "", nil, err
	}

	tmpFilePath := fmt.Sprintf("/tmp/roma-tmp-file-%d", time.Now().UnixNano())
	f, err := os.OpenFile(tmpFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(fileMode))
	if err != nil {
		return "", nil, err
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	var off int64
	buf := make([]byte, 2048)
	for {
		n, err := r.Read(buf)
		buffSize := int64(n)

		if err != nil && err != io.EOF {
			return "", nil, err
		}

		if off+buffSize > size && buf[n-1] == '\x00' {
			_, err := f.WriteAt(buf[:n-1], off)
			if err != nil {
				return "", nil, err
			}
			break
		} else if off+buffSize > size && buf[n-1] != '\x00' {
			return "", nil, errors.New("File size not match. ")
		}

		_, err = f.WriteAt(buf, off)
		if err != nil {
			return "", nil, err
		}
		off = off + buffSize
	}

	tmp, err := os.Open(tmpFilePath)
	if err != nil {
		return tmpFilePath, nil, err
	}

	return "", tmp, nil
}

func replyOk(w io.Writer) error {
	bufferedWriter := bufio.NewWriter(w)
	_, err := bufferedWriter.Write([]byte{responseOk})

	if err != nil {
		return err
	}

	err = bufferedWriter.Flush()
	if err != nil {
		return err
	}
	return nil
}

func replyErr(w io.Writer, replyErr error) error {
	bufferedWriter := bufio.NewWriter(w)
	if _, err := bufferedWriter.Write([]byte{responseError}); err != nil {
		return err
	}
	if _, err := bufferedWriter.Write([]byte(strings.ReplaceAll(replyErr.Error(), "\n", " "))); err != nil {
		return err
	}
	if _, err := bufferedWriter.Write([]byte{'\n'}); err != nil {
		return err
	}

	if err := bufferedWriter.Flush(); err != nil {
		return err
	}
	return nil
}
