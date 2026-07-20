package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
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

	// Check if double-clicked (no args, parent is explorer)

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "init":
			if err := generateDefaultConfig(configPath); err != nil {
				fatalPause("生成配置文件失败: %v", err)
			}
			fmt.Printf("✅ 配置文件已生成: %s\n", configPath)
			fmt.Println("请编辑配置文件后重新运行程序")
			pause()
			return
		case "version":
			fmt.Printf("DDNS-Updater %s (built: %s)\n", Version, BuildTime)
			return
		case "check":
			runOnce(configPath)
			return
		case "status":
			showStatus(configPath)
			pause()
			return
		case "help", "-h", "--help":
			printUsage()
			return
		}
	}

	// Default: run as daemon
	// Auto-generate config if missing
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fmt.Println("══════════════════════════════════════════════════")
		fmt.Println("  DDNS-Updater 首次运行")
		fmt.Println("══════════════════════════════════════════════════")
		fmt.Println()
		fmt.Println("未找到配置文件，正在自动生成...")
		if err := generateDefaultConfig(configPath); err != nil {
			fatalPause("生成配置文件失败: %v", err)
		}
		fmt.Printf("✅ 配置文件已生成: %s\n\n", configPath)
		fmt.Println("请按以下步骤操作：")
		fmt.Println("  1. 用记事本打开 config.yaml")
		fmt.Println("  2. 修改 file_updaters.path 为你的引擎配置文件路径")
		fmt.Println("  3. 修改 file_updaters.old 为你当前的公网IP")
		fmt.Println("  4. 修改 commands 为你的微端更新命令")
		fmt.Println("  5. 保存后重新运行本程序")
		fmt.Println()
		fmt.Println("正在打开配置文件...")
		fmt.Println()
		openNotepad(configPath)
		pause()
		return
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

func getStatePath(cfg *Config) string {
	if cfg.StateFile != "" {
		return cfg.StateFile
	}
	exe, err := os.Executable()
	if err != nil {
		return ".ddns-state.json"
	}
	return filepath.Join(filepath.Dir(exe), ".ddns-state.json")
}

func printUsage() {
	fmt.Println(`DDNS-Updater - 动态IP自动更新工具
适用于传奇服务端引擎控制台和微端网关的IP自动更新

用法:
  ddns-updater.exe              启动后台监控（默认）
  ddns-updater.exe init         生成默认配置文件
  ddns-updater.exe check        立即检测一次IP并更新
  ddns-updater.exe status       查看当前状态和日志
  ddns-updater.exe version      显示版本号
  ddns-updater.exe help         显示帮助信息

配置文件: config.yaml（与程序同目录）`)
}

func showStatus(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fmt.Printf("❌ 加载配置失败: %v\n", err)
		return
	}
	statePath := getStatePath(cfg)
	s := loadState(statePath)

	fmt.Println("╔══════════════════════════════════════════════════╗")
	fmt.Println("║          DDNS-Updater 运行状态                  ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")

	fmt.Printf("║  🌐 当前外网IP:     %-28s ║\n", orDash(s.CurrentIP))

	if !s.LastChangeTime.IsZero() {
		fmt.Printf("║  🔄 最近更换时间:   %-28s ║\n", s.LastChangeTime.Format("2006-01-02 15:04:05"))
	} else {
		fmt.Printf("║  🔄 最近更换时间:   %-28s ║\n", "暂无记录")
	}

	if !s.NextExpected.IsZero() && s.NextExpected.After(time.Now()) {
		remaining := time.Until(s.NextExpected)
		days := int(remaining.Hours()) / 24
		hours := int(remaining.Hours()) % 24
		fmt.Printf("║  ⏰ 预计下次更换:   %-28s ║\n", s.NextExpected.Format("2006-01-02 15:04:05"))
		fmt.Printf("║      剩余约:        %d天%d小时                        ║\n", days, hours)
	} else if !s.NextExpected.IsZero() {
		fmt.Printf("║  ⏰ 预计下次更换:   %-28s ║\n", s.NextExpected.Format("2006-01-02 15:04:05")+" (已到期)")
	} else {
		fmt.Printf("║  ⏰ 预计下次更换:   %-28s ║\n", "暂无记录")
	}

	if !s.LastCheckTime.IsZero() {
		fmt.Printf("║  📡 最近检测时间:   %-28s ║\n", s.LastCheckTime.Format("2006-01-02 15:04:05"))
	}

	fmt.Printf("║  📊 累计更换次数:   %-28d ║\n", s.ChangeCount)

	fmt.Println("╠══════════════════════════════════════════════════╣")
	fmt.Println("║  📋 最近一次更新详情                             ║")
	fmt.Println("╠══════════════════════════════════════════════════╣")

	if len(s.LastUpdateLog) > 0 {
		last := s.LastUpdateLog[len(s.LastUpdateLog)-1]
		fmt.Printf("║  IP变化: %s -> %s\n", last.OldIP, last.NewIP)
		fmt.Printf("║  时间:   %s\n", last.Time.Format("2006-01-02 15:04:05"))
		fmt.Println("║")
		for _, r := range last.Results {
			icon := "✅"
			if !r.Success {
				icon = "❌"
			}
			fmt.Printf("║  %s [%s] %s\n", icon, r.Type, r.Name)
			if r.Message != "" {
				fmt.Printf("║     %s\n", r.Message)
			}
		}
	} else {
		fmt.Println("║  暂无更新记录")
	}

	fmt.Println("╚══════════════════════════════════════════════════╝")
}

func orDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func runOnce(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fatalPause("加载配置失败: %v", err)
	}

	ip, err := detectPublicIP(cfg.DetectURLs)
	if err != nil {
		fatalPause("检测公网IP失败: %v", err)
	}

	fmt.Printf("当前公网IP: %s\n", ip)

	statePath := getStatePath(cfg)
	state := loadState(statePath)

	if state.HasChanged(ip) {
		oldIP := state.LastIP
		fmt.Printf("IP变化: %s -> %s，开始更新...\n", oldIP, ip)
		results := runUpdates(cfg, ip, true)
		state.RecordChange(oldIP, ip, results)
		fmt.Printf("✅ IP已更新: %s -> %s\n", oldIP, ip)
	} else {
		fmt.Println("IP未变化，无需更新")
	}

	state.RecordCheck(ip)
	saveState(statePath, state)
}

func runDaemon(configPath string) {
	cfg, err := loadConfig(configPath)
	if err != nil {
		fatalPause("加载配置失败: %v", err)
	}

	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("════════════════════════════════════════")
	log.Printf("  DDNS-Updater %s 启动", Version)
	log.Printf("════════════════════════════════════════")
	log.Printf("检测间隔: %d秒", cfg.Interval)
	log.Printf("公网IP检测源: %d个", len(cfg.DetectURLs))
	log.Printf("文件更新器: %d个", len(cfg.FileUpdaters))
	log.Printf("自动命令: %d个", len(cfg.Commands))
	log.Printf("────────────────────────────────────────")

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

	statePath := getStatePath(cfg)
	state := loadState(statePath)
	state.RecordCheck(ip)

	if state.HasChanged(ip) {
		oldIP := state.LastIP
		log.Printf("🔄 IP变化: %s -> %s，开始更新...", oldIP, ip)

		results := runUpdates(cfg, ip, false)
		state.RecordChange(oldIP, ip, results)

		log.Printf("✅ IP已更新: %s -> %s", oldIP, ip)
	}

	saveState(statePath, state)
}

func runUpdates(cfg *Config, ip string, printOutput bool) []LogEntry {
	var results []LogEntry

	for i := range cfg.FileUpdaters {
		u := &cfg.FileUpdaters[i]
		entry := LogEntry{Name: u.Name, Type: "文件"}
		if err := updateFile(u, ip); err != nil {
			if printOutput {
				fmt.Printf("  ❌ 文件更新失败 [%s]: %v\n", u.Name, err)
			}
			log.Printf("  ❌ 文件更新失败 [%s]: %v", u.Name, err)
			entry.Success = false
			entry.Message = err.Error()
		} else {
			if printOutput {
				fmt.Printf("  ✅ 文件已更新 [%s]: %s\n", u.Name, u.Path)
			}
			log.Printf("  ✅ 文件已更新 [%s]: %s", u.Name, u.Path)
			entry.Success = true
			entry.Message = u.Path
		}
		results = append(results, entry)
	}

	for i := range cfg.Commands {
		cmd := &cfg.Commands[i]
		entry := LogEntry{Name: cmd.Name, Type: "命令"}
		if err := executeCommand(cmd); err != nil {
			if printOutput {
				fmt.Printf("  ❌ 命令执行失败 [%s]: %v\n", cmd.Name, err)
			}
			log.Printf("  ❌ 命令执行失败 [%s]: %v", cmd.Name, err)
			entry.Success = false
			entry.Message = err.Error()
		} else {
			if printOutput {
				fmt.Printf("  ✅ 命令已执行 [%s]\n", cmd.Name)
			}
			log.Printf("  ✅ 命令已执行 [%s]", cmd.Name)
			entry.Success = true
		}
		results = append(results, entry)
	}

	return results
}

// pause waits for user to press Enter (for interactive mode)
func pause() {
	fmt.Println()
	fmt.Print("按回车键退出...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
}

// fatalPause shows error and waits for user input
func fatalPause(format string, args ...interface{}) {
	fmt.Printf("❌ "+format+"\n", args...)
	pause()
	os.Exit(1)
}

// openNotepad opens a file in notepad (Windows)
func openNotepad(path string) {
	// Use start command to open notepad
	cmd := exec.Command("cmd", "/c", "start", "", "notepad", path)
	cmd.SysProcAttr = getSysProcAttr()
	cmd.Start()
}
