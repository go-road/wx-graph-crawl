package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:                            "小绿书爬虫",
		Width:                            1024,
		MinWidth:                         1000,
		Height:                           768,
		MinHeight:                        600,
		HideWindowOnClose:                true,        // 关闭窗口时隐藏而不是退出
		EnableFraudulentWebsiteDetection: true,        // 启用针对欺诈内容（例如恶意软件或网络钓鱼尝试）的扫描服务
		LogLevel:                         logger.INFO, // 日志级别
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,     // 创建窗口并即将开始加载前端资源时的回调
		OnShutdown:       app.shutdown,    // 应用程序即将退出时的回调
		OnBeforeClose:    app.beforeClose, // 应用关闭前的回调
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title:   "wxGraphCrawler",
				Message: "专用于抓取微信公众号图文类型（俗称：小绿书）中的图片小工具 \r\n @Copyright 2025 by Alex",
				Icon:    nil,
			},
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
