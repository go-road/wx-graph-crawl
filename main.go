package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
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
	rootPath, err := getRootPath() // 获取项目根目录
	if err != nil {
		log.Fatalf("获取项目根目录失败: %+v", err)
	}
	global.RootPath = rootPath
	cfg := configs.GetConfig() // 项目配置
	cfg.Log.LogDir = filepath.Join(rootPath, cfg.Log.LogDir)
	utils.ConsoleBlue(fmt.Sprintf("Run At: %s", rootPath))

	// 初始化日志
	zapLogger := bootstrap.InitZapLog(cfg)
	defer zapLogger.Sync()
	// 初始化数据库
	dbPath := filepath.Join(rootPath, cfg.DB.SQLite3FilePath)
	db, err := bootstrap.InitDB(dbPath, cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %+v", err)
	}
	defer bootstrap.CloseDB()
	global.DB = db

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

// getRootPath 获取项目根目录
func getRootPath() (string, error) {
	var (
		exePath, rootPathByExecutable, rootPathByCaller, tmpDir string
		err                                                     error
	)
	// 第一种方式：获取当前执行程序所在的绝对路径
	// 这种仅在 `go build` 时，才可以获取正确的路径
	// 获取当前执行的二进制文件的全路径，包括二进制文件名
	exePath, err = os.Executable()
	if err != nil {
		return "", errors.Wrap(err, "获取当前执行文件路径失败 Executable")
	}
	rootPathByExecutable, err = filepath.EvalSymlinks(filepath.Dir(exePath))
	if err != nil {
		return "", errors.Wrap(err, "获取当前执行文件路径失败 EvalSymlinks")
	}

	// 第二种方式：获取当前执行文件绝对路径
	// 这种方式在 `go run` 和 `go build` 时，都可以获取到正确的路径
	// 但是交叉编译后，执行的结果是错误的结果
	_, filename, _, ok := runtime.Caller(0)
	if ok {
		rootPathByCaller = path.Dir(filename)
	}

	// 可以通过 `echo $TMPDIR` 查看当前系统临时目录
	tmpDir, err = filepath.EvalSymlinks(os.TempDir())
	if err != nil {
		return "", errors.Wrap(err, "获取当前系统临时目录失败")
	}

	// 对比通过 `os.Executable()` 获取到的路径是否与 `TMPDIR` 环境变量设置的路径相同
	// 相同，则说明是通过 `go run` 命令启动的
	// 不同，则是通过 `go build` 命令启动的
	if strings.Contains(rootPathByExecutable, tmpDir) {
		return rootPathByCaller, nil
	}

	return rootPathByExecutable, nil
}
