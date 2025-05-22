# Makefile

WAILS := wails

# Targets
.PHONY: dev build build-macos build-windows doctor gm v

# 本地运行
dev:
	$(WAILS) dev -browser

build:
	$(WAILS) build -clean

# 打包到 macOS
build-macos:
	$(WAILS) build -clean -platform darwin/amd64,darwin/arm64

# 打包到 Windows
build-windows:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(WAILS) build -clean -platform windows/amd64

# 检查环境
doctor:
	$(WAILS) doctor

# 生成 frontend/wailsjs 目录下的代码
gm:
	$(WAILS) generate module

# 打印 wails 版本
v:
	$(WAILS) version