package main

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/configs"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

func (a *App) setContext(ctx context.Context) {
	a.ctx = ctx
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	runtime.LogInfo(ctx, "项目启动成功")
}

func (a *App) shutdown(ctx context.Context) {
	runtime.LogInfo(ctx, "项目关闭成功")
}

func (a *App) beforeClose(ctx context.Context) (prevent bool) {
	dialog, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "提示",
		Message:       "是否确定退出？",
		Buttons:       []string{"取消", "确定退出"},
		DefaultButton: "取消",
	})

	if err != nil {
		runtime.LogFatalf(ctx, "关闭应用程序时，出错： %s", err.Error())
	}

	if dialog == "确定退出" {
		envInfo := runtime.Environment(ctx)
		runtime.LogInfof(ctx, "关闭应用程序，环境信息：%+v", envInfo)
	}

	// 如果用户点击了“取消”，则阻止应用程序关闭
	prevent = "确定退出" != dialog

	return
}

// 设置一些参数选项 https://wails.io/zh-Hans/docs/reference/options
func (a *App) macOptions(cfg *configs.Config) *mac.Options {
	return &mac.Options{
		About: &mac.AboutInfo{
			Title:   cfg.App.MacTitle,
			Message: cfg.App.MacMessage,
			Icon:    nil,
		},
	}
}

func (a *App) logLevel(input string) logger.LogLevel {
	// debug、info、warn、error
	allow := map[string]logger.LogLevel{
		"debug": logger.DEBUG,
		"info":  logger.INFO,
		"warn":  logger.WARNING,
		"error": logger.ERROR,
	}
	result, ok := allow[input]
	if !ok {
		return logger.DEBUG
	}
	return result
}
