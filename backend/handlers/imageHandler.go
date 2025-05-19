package handlers

import (
	"context"
	"fmt"

	"github.com/pudongping/wx-graph-crawl/backend/service"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"go.uber.org/zap"
)

var _ ContextSetter = (*ImageHandler)(nil)

type ImageHandler struct {
	ctx context.Context
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

func (h *ImageHandler) SetContext(ctx context.Context) {
	h.ctx = ctx
}

func (h *ImageHandler) Crawling(req types.CrawlingRequest) (res types.CrawlingResponse, err error) {
	zap.L().Info("开始爬取图片", zap.String("请求参数", fmt.Sprintf("%+v", req)))
	res, err = service.NewImageService().Crawling(h.ctx, req)
	zap.L().Info("爬取图片结束", zap.String("返回结果", fmt.Sprintf("%+v", res)))
	return
}

func (h *ImageHandler) Cropping(req types.CroppingRequest) (res types.CroppingResponse, err error) {
	zap.L().Info("开始裁剪图片", zap.String("请求参数", fmt.Sprintf("%+v", req)))
	res, err = service.NewImageService().Cropping(h.ctx, req)
	zap.L().Info("裁剪图片结束", zap.String("返回结果", fmt.Sprintf("%+v", res)))
	return
}

func (h *ImageHandler) Shuffling(req types.ShufflingRequest) (res types.ShufflingResponse, err error) {
	zap.L().Info("开始移动图片", zap.String("请求参数", fmt.Sprintf("%+v", req)))
	res, err = service.NewImageService().Shuffling(h.ctx, req)
	zap.L().Info("移动图片结束", zap.String("返回结果", fmt.Sprintf("%+v", res)))
	return
}
