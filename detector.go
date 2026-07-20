package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func detectPublicIP(urls []string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	for _, url := range urls {
		ip, err := fetchIP(client, url)
		if err == nil && ip != "" {
			return ip, nil
		}
	}

	return "", fmt.Errorf("所有IP检测源均失败")
}

func fetchIP(client *http.Client, url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	ip := strings.TrimSpace(string(body))
	// Remove IPv6 brackets if present
	ip = strings.Trim(ip, "[]")
	// Take first line only
	if idx := strings.IndexByte(ip, '\n'); idx >= 0 {
		ip = ip[:idx]
	}

	return ip, nil
}
