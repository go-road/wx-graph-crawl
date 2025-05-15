package main

import (
	"context"
	"fmt"

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

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
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

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
