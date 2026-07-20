package main

import (
	"os"
	"strings"
)

func loadState(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveState(path, ip string) error {
	return os.WriteFile(path, []byte(ip), 0644)
}
