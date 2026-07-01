package sftp

import (
	"fmt"
	"io"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

type FileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

func (h *Handler) ListFiles(host string, port int, username, password, path string) ([]FileInfo, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return nil, err
	}
	defer sftpClient.Close()

	files, err := sftpClient.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var result []FileInfo
	for _, f := range files {
		result = append(result, FileInfo{
			Name:    f.Name(),
			Size:    f.Size(),
			Mode:    f.Mode().String(),
			ModTime: f.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   f.IsDir(),
		})
	}

	return result, nil
}

func (h *Handler) DownloadFile(host string, port int, username, password, remotePath string) (io.ReadCloser, error) {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, err
	}

	file, err := sftpClient.Open(remotePath)
	if err != nil {
		sftpClient.Close()
		client.Close()
		return nil, err
	}

	return file, nil
}

func (h *Handler) UploadFile(host string, port int, username, password, remotePath string, reader io.Reader) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	file, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	return err
}

func (h *Handler) DeleteFile(host string, port int, username, password, remotePath string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Remove(remotePath)
}

func (h *Handler) Mkdir(host string, port int, username, password, remotePath string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return err
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return sftpClient.Mkdir(remotePath)
}