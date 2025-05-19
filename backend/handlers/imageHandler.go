package handlers

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/pudongping/wx-graph-crawl/backend/constant"
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
	start := time.Now()
	httpClientTimeout := time.Duration(req.TimeoutSeconds) * time.Second
	textContentFilePath := filepath.Join(req.ImgSavePath, constant.TextContentFileName) // 文字内容保存的路径
	res.TextContentSavePath = textContentFilePath

	crawlerImgSvc := service.NewCrawlerImgService(req.ImgUrls, httpClientTimeout, req.ImgSavePath, textContentFilePath)
	var spiderResults []types.CrawlResult
	spiderResults, err = crawlerImgSvc.RunSpiderImg()
	if err != nil {
		zap.L().Error("爬取图片失败", zap.Error(err))
		return res, err
	}

	// 统计 抓取的链接地址数量 抓取成功并保存成功的图片数量
	for _, item := range spiderResults {
		res.CrawlUrlCount++
		res.CrawlImgCount += int64(len(item.ImgSavePathSuccess))
		if item.Err != nil {
			zap.L().Error("抓取失败", zap.Int("Num", item.Number), zap.String("Url", item.URL), zap.Error(item.Err))
			res.ErrContent += item.Err.Error() + " | \n"
		}
	}

	castTime := time.Since(start)
	res.CastTimeStr = castTime.String()

	zap.L().Info("爬取图片结束", zap.String("返回结果", fmt.Sprintf("%+v", res)))
	return res, nil
}

func (h *ImageHandler) Cropping(req types.CroppingRequest) (res types.CroppingResponse, err error) {
	zap.L().Info("开始裁剪图片", zap.String("请求参数", fmt.Sprintf("%+v", req)))
	start := time.Now()
	concurrencyMax := 10 // 并发数
	cropSvc := service.NewCropImgService(req.ImgSavePath, concurrencyMax, req.BottomPixel)
	cropResults, err := cropSvc.RunCropImg()
	if err != nil {
		zap.L().Error("裁剪失败", zap.Error(err))
		return
	}

	res.CropImgPath = req.ImgSavePath
	// 统计裁剪成功的图片数量
	for _, item := range cropResults {
		if item.Err != nil {
			zap.L().Error("裁剪失败", zap.String("ImgPath", item.ImgPath), zap.Error(item.Err))
			res.ErrContent += item.Err.Error() + " | \n"
		} else {
			res.CropImgCount++
		}
	}
	castTime := time.Since(start)
	res.CastTimeStr = castTime.String()

	zap.L().Info("裁剪图片结束", zap.String("返回结果", fmt.Sprintf("%+v", res)))
	return res, nil
}

func (h *ImageHandler) Shuffling(req types.ShufflingRequest) (res types.ShufflingResponse, err error) {
	zap.L().Info("开始移动图片", zap.String("请求参数", fmt.Sprintf("%+v", req)))
	start := time.Now()
	moveImgSvc := service.NewMoveImgService(req.ImgSavePath, req.MaxNumImage)
	err = moveImgSvc.RunMoveImg()
	if err != nil {
		zap.L().Error("移动图片失败", zap.Error(err))
		return
	}
	castTime := time.Since(start)
	res.ShuffleImgPath = req.ImgSavePath
	res.CastTimeStr = castTime.String()

	zap.L().Info("移动图片结束", zap.String("返回结果", fmt.Sprintf("%+v", res)))
	return res, nil
}
