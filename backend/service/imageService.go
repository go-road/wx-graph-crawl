package service

import (
	"context"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/constant"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"go.uber.org/zap"
)

type ImageService struct {
}

func NewImageService() *ImageService {
	return &ImageService{}
}

func (svc *ImageService) Crawling(ctx context.Context, req types.CrawlingRequest) (res types.CrawlingResponse, err error) {
	start := time.Now()
	httpClientTimeout := time.Duration(req.TimeoutSeconds) * time.Second
	textContentFilePath := filepath.Join(req.ImgSavePath, constant.TextContentFileName) // 文字内容保存的路径
	res.TextContentSavePath = textContentFilePath

	textContentFileDir := filepath.Join(req.ImgSavePath, constant.TextContentFileDir) // 文字内容保存的目录
	res.TextContentSaveDir = textContentFileDir

	crawlerImgSvc := NewCrawlerImgService(req.ImgUrls, httpClientTimeout, req.ImgSavePath, textContentFilePath, textContentFileDir)
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

	return res, nil
}

func (svc *ImageService) Cropping(ctx context.Context, req types.CroppingRequest) (res types.CroppingResponse, err error) {
	start := time.Now()
	concurrencyMax := 10 // 并发数
	cropSvc := NewCropImgService(req.ImgSavePath, concurrencyMax, req.BottomPixel)
	cropResults, err := cropSvc.RunCropImg()
	if err != nil {
		zap.L().Error("裁剪失败", zap.Error(err))
		return res, errors.Wrap(err, "裁剪图片失败")
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

	return res, nil
}

func (svc *ImageService) Shuffling(ctx context.Context, req types.ShufflingRequest) (res types.ShufflingResponse, err error) {
	start := time.Now()
	moveImgSvc := NewMoveImgService(req.ImgSavePath, req.MaxNumImage)
	err = moveImgSvc.RunMoveImg()
	if err != nil {
		zap.L().Error("移动图片失败", zap.Error(err))
		return res, errors.Wrap(err, "移动图片失败")
	}
	castTime := time.Since(start)
	res.ShuffleImgPath = req.ImgSavePath
	res.CastTimeStr = castTime.String()

	return res, nil
}
