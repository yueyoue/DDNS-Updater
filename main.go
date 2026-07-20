package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	configPath := getConfigPath()

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := generateDefaultConfig(configPath); err != nil {
				log.Fatalf("生成配置文件失败: %v", err)
			}
			fmt.Printf("✅ 配置文件已生成: %s\n", configPath)
			fmt.Println("请编辑配置文件后重新运行程序")
			return
		case "version":
			fmt.Printf("DDNS-Updater %s (built: %s)\n", Version, BuildTime)
			return
		case "check":
			runOnce(configPath)
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}

	runDaemon(configPath)
}

func getConfigPath() string {
	if p := os.Getenv("DDNS_CONFIG"); p != "" {
		return p
	}
	exe, err := os.Executable()
	if err != nil {
		return "config.yaml"
	}
	return filepath.Join(filepath.Dir(exe), "config.yaml")
}

func printUsage() {
	fmt.Println(`DDNS-Updater - 动态IP自动更新工具
适用于传奇服务端引擎控制台和微端网关的IP自动更新

用法:
  ddns-updater.exe              启动后台监控（默认）
  ddns-updater.exe init         生成默认配置文件
  ddns-updater.exe check        立即检测一次IP并更新
  ddns-updater.exe version      显示版本号
  ddns-updater.exe help         显示帮助信息

配置文件: config.yaml（与程序同目录）`)
}

func runOnce(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	ip, err := detectPublicIP(cfg.DetectURLs)
	if err != nil {
		log.Fatalf("检测公网IP失败: %v", err)
	}

	fmt.Printf("当前公网IP: %s\n", ip)

	changed, oldIP, err := checkAndUpdate(cfg, ip)
	if err != nil {
		log.Fatalf("更新失败: %v", err)
	}

	if changed {
		fmt.Printf("✅ IP已更新: %s -> %s\n", oldIP, ip)
	} else {
		fmt.Println("IP未变化，无需更新")
	}
}

func runDaemon(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	log.Printf("DDNS-Updater %s 启动", Version)
	log.Printf("检测间隔: %d秒", cfg.Interval)
	log.Printf("公网IP检测源: %d个", len(cfg.DetectURLs))
	log.Printf("文件更新器: %d个", len(cfg.FileUpdaters))
	log.Printf("自动命令: %d个", len(cfg.Commands))

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer ticker.Stop()

	runCheck(cfg)

	for {
		select {
		case <-ticker.C:
			runCheck(cfg)
		case sig := <-sigCh:
			log.Printf("收到信号 %v，正在退出...", sig)
			return
		}
	}
}

func runCheck(cfg *Config) {
	ip, err := detectPublicIP(cfg.DetectURLs)
	if err != nil {
		log.Printf("⚠️ 检测公网IP失败: %v", err)
		return
	}

	changed, oldIP, err := checkAndUpdate(cfg, ip)
	if err != nil {
		log.Printf("⚠️ 更新失败: %v", err)
		return
	}

	if changed {
		log.Printf("✅ IP已更新: %s -> %s", oldIP, ip)
	}
}

func checkAndUpdate(cfg *Config, newIP string) (changed bool, oldIP string, err error) {
	statePath := getStatePath(cfg)
	oldIP = loadState(statePath)

	if oldIP == newIP {
		return false, oldIP, nil
	}

	log.Printf("🔄 IP变化: %s -> %s，开始更新...", oldIP, newIP)

	for i := range cfg.FileUpdaters {
		u := &cfg.FileUpdaters[i]
		if err := updateFile(u, newIP); err != nil {
			log.Printf("  ❌ 文件更新失败 [%s]: %v", u.Name, err)
		} else {
			log.Printf("  ✅ 文件已更新 [%s]: %s", u.Name, u.Path)
		}
	}

	for i := range cfg.Commands {
		cmd := &cfg.Commands[i]
		if err := executeCommand(cmd); err != nil {
			log.Printf("  ❌ 命令执行失败 [%s]: %v", cmd.Name, err)
		} else {
			log.Printf("  ✅ 命令已执行 [%s]", cmd.Name)
		}
	}

	saveState(statePath, newIP)
	return true, oldIP, nil
}

func getStatePath(cfg *Config) string {
	if cfg.StateFile != "" {
		return cfg.StateFile
	}
	exe, err := os.Executable()
	if err != nil {
		return ".ddns-state"
	}
	return filepath.Join(filepath.Dir(exe), ".ddns-state")
}
