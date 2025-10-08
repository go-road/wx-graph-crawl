package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveFile 将给定内容保存到指定路径的文件中
func SaveFile(content, filePath string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// 写入文件
	return os.WriteFile(filePath, []byte(content), 0644)
}

// GetDefaultDownloadsDir 获取默认的下载目录
func GetDefaultDownloadsDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "downloads"
	}
	fmt.Printf("用户主目录：%s\n", homeDir)

	wd, err := os.Getwd()
	if err != nil {
		return "downloads"
	}
	fmt.Printf("当前工作目录：%s\n", wd)

	// 获取当前盘符
	drive := filepath.VolumeName(wd)
	fmt.Printf("当前盘符：%s\n", drive)
	return filepath.Join(drive, "\\", "downloads")
}
