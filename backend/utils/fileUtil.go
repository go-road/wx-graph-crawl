package utils

import (
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
