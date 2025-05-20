package handlers

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/service"
	"github.com/pudongping/wx-graph-crawl/backend/types"
)

var _ ContextSetter = (*FileHandler)(nil)

type FileHandler struct {
	ctx context.Context
}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (h *FileHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

// SelectFile 选择文件并返回文件路径和内容
func (h *FileHandler) SelectFile() (res types.SelectFileResponse, err error) {
	return service.NewFileService().SelectFile(h.ctx)
}

// SelectDirectory 选择目录并返回目录路径
func (h *FileHandler) SelectDirectory() (string, error) {
	return service.NewFileService().SelectDirectory(h.ctx)
}
