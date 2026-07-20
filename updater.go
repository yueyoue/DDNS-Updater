package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

func updateFile(u *FileUpdater, newIP string) error {
	data, err := os.ReadFile(u.Path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	content := string(data)
	oldVal := u.Old
	newVal := strings.ReplaceAll(u.New, "{{.IP}}", newIP)

	if !strings.Contains(content, oldVal) {
		return nil
	}

	updated := strings.ReplaceAll(content, oldVal, newVal)
	return os.WriteFile(u.Path, []byte(updated), 0644)
}

func executeCommand(cmd *CommandConfig) error {
	timeout := time.Duration(cmd.Timeout) * time.Second
	c := exec.Command(cmd.Cmd, cmd.Args...)

	done := make(chan error, 1)
	go func() {
		done <- c.Run()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		c.Process.Kill()
		return fmt.Errorf("命令执行超时 (%d秒)", cmd.Timeout)
	}
}
