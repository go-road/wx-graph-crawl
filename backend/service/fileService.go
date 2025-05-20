package service

import (
	"bufio"
	"context"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/constant"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type FileService struct {
}

func NewFileService() *FileService {
	return &FileService{}
}

// ValidateURL 检查 URL 是否符合要求（是否为小绿书的 URL）
func (svc *FileService) validateIfWXURL(url string) bool {
	url = strings.TrimSpace(url)

	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		return false
	}

	return strings.Contains(url, constant.WXMPTWDomain)
}

// readURLFile 读取文件内容并返回 URL 列表
func (svc *FileService) readURLFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "读取URL文件，打开文件时")
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // 跳过空行
		}
		urls = append(urls, line)
	}

	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "读取URL文件，扫描文件时")
	}

	return urls, nil
}

// SelectFile 选择文件并返回文件路径和内容
// 返回给 js 的方法，只能返回 2 个值，第二个值必须是错误，（第一个返回值会被 resolve 接收，第二个返回值会被 reject 接收）
// 详见 https://wails.io/zh-Hans/docs/howdoesitwork/#method-binding
func (svc *FileService) SelectFile(ctx context.Context) (res types.SelectFileResponse, err error) {
	// 打开文件选择对话框
	filePath, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "请选择URL文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "文本文件 (*.txt)",
				Pattern:     "*.txt",
			},
		},
	})
	if err != nil {
		err = errors.Wrap(err, "SelectFile打开文件Dialog时")
		return
	}
	if filePath == "" {
		// 表明用户取消了选择
		return
	}

	// 读取文件内容
	urls, err := svc.readURLFile(filePath)
	if err != nil {
		err = errors.Wrap(err, "SelectFile读取文件内容时")
		return
	}

	// 验证 URL
	var validURLs []string
	for _, u := range urls {
		if svc.validateIfWXURL(u) {
			validURLs = append(validURLs, u)
		}
	}

	res.FilePath = filePath
	res.ValidURLs = validURLs

	return
}

// SelectDirectory 选择目录并返回目录路径
func (svc *FileService) SelectDirectory(ctx context.Context) (string, error) {
	// 打开目录选择对话框
	dirPath, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title: "请选择图片保存目录",
	})
	if err != nil {
		return "", errors.Wrap(err, "选择目录，打开目录Dialog时")
	}

	if dirPath == "" {
		// 用户取消了选择
		return "", nil
	}

	return dirPath, nil
}
