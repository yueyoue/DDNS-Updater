package main

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// updateFile replaces IP in a text config file
func updateFile(u *FileUpdater, newIP string) error {
	data, err := os.ReadFile(u.Path)
	if err != nil {
		return fmt.Errorf("读取文件失败: %w", err)
	}

	content := string(data)
	oldVal := u.Old
	newVal := strings.ReplaceAll(u.New, "{{.IP}}", newIP)

	if !strings.Contains(content, oldVal) {
		// Already updated or pattern not found
		return nil
	}

	updated := strings.ReplaceAll(content, oldVal, newVal)
	return os.WriteFile(u.Path, []byte(updated), 0644)
}

// updateDatabase updates IP in SQLite database
func updateDatabase(u *DBUpdater, newIP string) error {
	db, err := sql.Open("sqlite3", u.Path)
	if err != nil {
		return fmt.Errorf("打开数据库失败: %w", err)
	}
	defer db.Close()

	for _, q := range u.Queries {
		sqlStr := strings.ReplaceAll(q.SQL, "{{.IP}}", newIP)
		result, err := db.Exec(sqlStr)
		if err != nil {
			return fmt.Errorf("执行SQL失败 [%s]: %w", q.Desc, err)
		}
		rows, _ := result.RowsAffected()
		if rows > 0 {
			// Updated successfully
		}
	}

	return nil
}

// executeCommand runs an external command
func executeCommand(cmd *CommandConfig) error {
	timeout := time.Duration(cmd.Timeout) * time.Second
	c := exec.Command(cmd.Cmd, cmd.Args...)
	c.SysProcAttr = getSysProcAttr()

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
