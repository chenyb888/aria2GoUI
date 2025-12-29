package main

import (
	"log"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2/app"
	"github.com/chenyb888/aria2GoUI/internal/config"
	"github.com/chenyb888/aria2GoUI/internal/aria2"
	"github.com/chenyb888/aria2GoUI/internal/ui"
)

func main() {
	// 设置环境变量以支持中文字体，使用单个字体文件而非字体集合
	fontPath := "C:\\Windows\\Fonts\\simhei.ttf" // 黑体（单个字体文件）
	if _, err := os.Stat(fontPath); err != nil {
		// 如果黑体不存在，尝试其他中文字体
		fontPath = "C:\\Windows\\Fonts\\simsun.ttc" // 宋体（字体集合，作为备选）
		if _, err := os.Stat(fontPath); err != nil {
			fontPath = "" // 使用系统默认字体
		}
	}
	
	if fontPath != "" {
		os.Setenv("FYNE_FONT", fontPath)
	}
	
	// 创建 Fyne 应用程序
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(ui.NewChineseFontTheme())

	// 加载配置
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("加载配置失败: %v，使用默认配置", err)
		cfg = config.DefaultConfig()
	}

	// 创建 UI 应用
	uiApp := ui.NewApp()
	uiApp.SetConfig(cfg)

	// 创建 aria2 客户端
	aria2Client := aria2.NewClient(
		cfg.RPC.Host,
		cfg.RPC.Port,
		cfg.RPC.Token,
		cfg.RPC.Protocol,
		cfg.RPC.Path,
	)
	uiApp.SetAria2Client(aria2Client)

	// 测试连接
	if err := testConnection(aria2Client); err != nil {
		log.Printf("连接 aria2 失败: %v", err)
		// 显示连接设置对话框
		uiApp.ShowConnectionDialog()
	} else {
		log.Println("成功连接到 aria2")
	}

	// 创建主界面
	uiApp.CreateMainUI()

	// 显示并运行
	uiApp.ShowAndRun()
}

// testConnection 测试 aria2 连接
func testConnection(client *aria2.Client) error {
	_, err := client.GetVersion()
	return err
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "config.json"
	}
	return filepath.Join(homeDir, ".aria2goui", "config.json")
}