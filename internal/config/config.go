package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config 应用程序配置
type Config struct {
	RPC      RPCConfig      `json:"rpc"`
	UI       UIConfig       `json:"ui"`
	General  GeneralConfig  `json:"general"`
	Download DownloadConfig `json:"download"`
	Advanced AdvancedConfig `json:"advanced"`
	Display  DisplayConfig  `json:"display"`
	Notify   NotifyConfig   `json:"notify"`
}

// RPCConfig aria2 RPC 连接配置
type RPCConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Token        string `json:"token"`
	Path         string `json:"path"`
	Protocol     string `json:"protocol"`     // http, https, ws, wss
	AutoReconnect bool   `json:"auto_reconnect"`
	Timeout      int    `json:"timeout"`       // 连接超时时间（秒）
}

// UIConfig 界面配置
type UIConfig struct {
	Theme        string `json:"theme"`         // light, dark, auto
	Language     string `json:"language"`      // zh_CN, en_US
	WindowWidth  int    `json:"window_width"`
	WindowHeight int    `json:"window_height"`
	RefreshInterval int  `json:"refresh_interval"` // 刷新间隔（秒）
	PageTitle    string `json:"page_title"`
}

// GeneralConfig 通用配置
type GeneralConfig struct {
	AutoStart      bool `json:"auto_start"`
	MinimizeToTray bool `json:"minimize_to_tray"`
	ContinueTasks  bool `json:"continue_tasks"` // 启动时恢复未完成任务
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	DefaultDirectory    string `json:"default_directory"`
	MaxConcurrentDownloads int  `json:"max_concurrent_downloads"`
	MaxConnectionPerServer int  `json:"max_connection_per_server"`
	GlobalSpeedLimit    int    `json:"global_speed_limit"`    // KB/s
	UploadSpeedLimit    int    `json:"upload_speed_limit"`    // KB/s
}

// AdvancedConfig 高级配置
type AdvancedConfig struct {
	UserAgent    string `json:"user_agent"`
	HTTProxy     string `json:"http_proxy"`
	FTPProxy     string `json:"ftp_proxy"`
	BTPortRange  string `json:"bt_port_range"`  // "6881-6999"
	DHTEnabled   bool   `json:"dht_enabled"`
	PEXEnabled   bool   `json:"pex_enabled"`
	SeedDownload bool   `json:"seed_download"`
}

// DisplayConfig 显示配置
type DisplayConfig struct {
	ViewMode          string `json:"view_mode"`           // list, card
	SortBy            string `json:"sort_by"`             // name, size, progress, speed
	ShowFileList      bool   `json:"show_file_list"`
	ShowBTInfo        bool   `json:"show_bt_info"`
	ProgressBarStyle  string `json:"progress_bar_style"`  // normal, detailed
	AutoScrollToActive bool   `json:"auto_scroll_to_active"`
}

// NotifyConfig 通知配置
type NotifyConfig struct {
	SoundEnabled   bool `json:"sound_enabled"`
	SystemNotify   bool `json:"system_notify"`
	BrowserNotify  bool `json:"browser_notify"`
	ErrorNotify    bool `json:"error_notify"`
	CompleteNotify bool `json:"complete_notify"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		RPC: RPCConfig{
			Host:         "localhost",
			Port:         6800,
			Token:        "",
			Path:         "/jsonrpc",
			Protocol:     "http",
			AutoReconnect: true,
			Timeout:      30,
		},
		UI: UIConfig{
			Theme:           "light",
			Language:        "zh_CN",
			WindowWidth:     800,
			WindowHeight:    600,
			RefreshInterval: 5,
			PageTitle:       "aria2GoUI",
		},
		General: GeneralConfig{
			AutoStart:     false,
			MinimizeToTray: true,
			ContinueTasks: true,
		},
		Download: DownloadConfig{
			DefaultDirectory:       "",
			MaxConcurrentDownloads: 5,
			MaxConnectionPerServer: 16,
			GlobalSpeedLimit:       0,  // 0 表示无限制
			UploadSpeedLimit:       0,  // 0 表示无限制
		},
		Advanced: AdvancedConfig{
			UserAgent:    "aria2GoUI/1.0",
			HTTProxy:     "",
			FTPProxy:     "",
			BTPortRange:  "6881-6999",
			DHTEnabled:   true,
			PEXEnabled:   true,
			SeedDownload: true,
		},
		Display: DisplayConfig{
			ViewMode:            "list",
			SortBy:              "name",
			ShowFileList:        true,
			ShowBTInfo:          true,
			ProgressBarStyle:    "normal",
			AutoScrollToActive:  true,
		},
		Notify: NotifyConfig{
			SoundEnabled:   true,
			SystemNotify:   true,
			BrowserNotify:  false,
			ErrorNotify:    true,
			CompleteNotify: true,
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	// 如果配置文件不存在，返回默认配置
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveConfig 保存配置到文件
func (c *Config) SaveConfig(path string) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}