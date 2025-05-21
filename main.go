package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"path/filepath"

	"github.com/pudongping/wx-graph-crawl/backend"
	"github.com/pudongping/wx-graph-crawl/backend/bootstrap"
	"github.com/pudongping/wx-graph-crawl/backend/configs"
	"github.com/pudongping/wx-graph-crawl/backend/global"
	"github.com/pudongping/wx-graph-crawl/backend/utils"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	bootstrap.PrintLogo()
	rootPath, err := bootstrap.RootPath() // 获取项目根目录
	if err != nil {
		log.Fatalf("获取项目根目录失败: %+v", err)
	}
	global.RootPath = rootPath
	cfg := configs.GetConfig() // 项目配置
	cfg.Log.LogDir = filepath.Join(rootPath, cfg.Log.LogDir)
	utils.ConsoleBlue(fmt.Sprintf("Run At: %s", rootPath))

	// Initialize the backend
	zapLogger := bootstrap.InitZapLog(cfg)
	defer zapLogger.Sync()

	// 业务
	backendBoot := backend.NewBoot()

	wailsLogLevel := app.logLevel(cfg.Log.Level)
	wailsOptions := &options.App{
		Title:                            cfg.App.Title, // 应用程序标题
		Width:                            1024,
		MinWidth:                         900,
		Height:                           840,
		MinHeight:                        800,
		HideWindowOnClose:                true,          // 关闭窗口时隐藏而不是退出
		EnableFraudulentWebsiteDetection: true,          // 启用针对欺诈内容（例如恶意软件或网络钓鱼尝试）的扫描服务
		LogLevel:                         wailsLogLevel, // 日志级别
		AssetServer: &assetserver.Options{
			Assets: assets, // 应用程序的前端资产
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.setContext(ctx) // 设置上下文
			app.startup(ctx)    // 启动时的回调

			// 将应用启动时的上下文传递给业务逻辑（方便在业务逻辑代码中使用运行时函数）
			backendBoot.SetContext(ctx)

		}, // 创建窗口并即将开始加载前端资源时的回调
		OnShutdown:    app.shutdown,        // 应用程序即将退出时的回调
		OnBeforeClose: app.beforeClose,     // 应用关闭前的回调
		Mac:           app.macOptions(cfg), // macOS特定的选项
		Bind: append([]interface{}{ // 我们希望向前端暴露的一部分结构体实例
			app,
		}, backendBoot.Binds()...), // 动态绑定所有 handler
	}

	// Create application with options
	if err := wails.Run(wailsOptions); err != nil {
		log.Fatalf("Wails run error: %+v \n", err)
	}
}
