package utils

import (
	"os"

	"github.com/pkg/errors"
)

// MkdirIfNotExist 检查目录是否存在，如果不存在则创建
// 该函数会创建所有必要的父目录
// 例如，如果路径是 "/a/b/c"，而 "/a" 和 "/a/b" 不存在，则会创建 "/a" 和 "/a/b"
// 如果路径已经存在，则不会执行任何操作
func MkdirIfNotExist(path string) error {
	// os.Stat 用于检查目录是否存在，os.IsNotExist 判断错误类型是否是因为文件或目录不存在
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return errors.Wrap(err, "创建目录失败")
		}
		// 这一行代码必须要有，因为此时的 err 本来就不会为 nil 且此时的 err 就为文件或者目录不存在时的 error
		// 避免外层判断遭受干扰
		return nil
	}

	return err
}

// CreateFileIfNotExist 检查文件是否存在，如果不存在则创建
func CreateFileIfNotExist(filePath string) error {
	_, err := os.Stat(filePath)
	if err != nil && os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return errors.Wrap(err, "创建文件失败")
		}
		defer file.Close()
		return nil
	}

	return err
}
