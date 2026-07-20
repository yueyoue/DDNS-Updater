package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Interval    int              `yaml:"interval"`
	DetectURLs  []string         `yaml:"detect_urls"`
	StateFile   string           `yaml:"state_file,omitempty"`
	FileUpdaters []FileUpdater   `yaml:"file_updaters"`
	DBUpdaters   []DBUpdater     `yaml:"db_updaters"`
	Commands     []CommandConfig `yaml:"commands"`
}

type FileUpdater struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Old     string `yaml:"old"`
	New     string `yaml:"new"`
}

type DBUpdater struct {
	Name    string       `yaml:"name"`
	Path    string       `yaml:"path"`
	Queries []DBQuery    `yaml:"queries"`
}

type DBQuery struct {
	SQL  string `yaml:"sql"`
	Desc string `yaml:"desc,omitempty"`
}

type CommandConfig struct {
	Name    string `yaml:"name"`
	Cmd     string `yaml:"cmd"`
	Args    []string `yaml:"args,omitempty"`
	Timeout int    `yaml:"timeout,omitempty"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// Set defaults
	if cfg.Interval <= 0 {
		cfg.Interval = 60
	}
	if len(cfg.DetectURLs) == 0 {
		cfg.DetectURLs = defaultDetectURLs()
	}
	for i := range cfg.Commands {
		if cfg.Commands[i].Timeout <= 0 {
			cfg.Commands[i].Timeout = 30
		}
	}

	return cfg, nil
}

func generateDefaultConfig(path string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("配置文件已存在: %s（如需重新生成请先删除）", path)
	}

	cfg := defaultConfig()
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func defaultDetectURLs() []string {
	return []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"https://ipecho.net/plain",
	}
}

func defaultConfig() *Config {
	return &Config{
		Interval:   60,
		DetectURLs: defaultDetectURLs(),
		FileUpdaters: []FileUpdater{
			{
				Name: "引擎控制台-外网IP",
				Path: `D:\MirServer\Mir200\Envir\MapQuest.txt`,
				Old:  `YOUR_OLD_IP`,
				New:  `{{.IP}}`,
			},
		},
		DBUpdaters: []DBUpdater{
			{
				Name: "微端网关-服务器地址",
				Path: `D:\MirServer\微端网关\wd.db`,
				Queries: []DBQuery{
					{
						SQL:  `UPDATE server_list SET address = '{{.IP}}'`,
						Desc: "更新微端服务器地址",
					},
				},
			},
		},
		Commands: []CommandConfig{
			{
				Name:    "重启微端网关",
				Cmd:     `cmd`,
				Args:    []string{"/c", "echo IP已更新，请手动重启微端网关"},
				Timeout: 30,
			},
		},
	}
}
