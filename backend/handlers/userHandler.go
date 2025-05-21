package handlers

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/service"
	"github.com/pudongping/wx-graph-crawl/backend/types"
)

var _ ContextSetter = (*UserHandler)(nil)

type UserHandler struct {
	ctx context.Context
}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

// SetPreferenceInfo 设置用户偏好信息
func (h *UserHandler) SetPreferenceInfo(req types.SetPreferenceInfoRequest) (types.SetPreferenceInfoResponse, error) {
	return service.NewUserService().SetPreferenceInfo(h.ctx, req)
}

// GetPreferenceInfo 获取用户偏好信息
func (h *UserHandler) GetPreferenceInfo() (*types.GetPreferenceInfoResponse, error) {
	return service.NewUserService().GetPreferenceInfo()
}
