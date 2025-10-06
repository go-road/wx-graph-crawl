package service

import (
	"github.com/pudongping/wx-graph-crawl/backend/constant"
	"path/filepath"
	"testing"
	"time"
)

func TestRunSpiderImg(t *testing.T) {
	imgSavePath := "E:\\Downloads\\wx_graph_crawl_downloads"
	textContentFilePath := "E:\\Downloads\\wx_graph_crawl_downloads\\content.txt"
	textContentFileDir := filepath.Join(imgSavePath, constant.TextContentFileDir)
	wxTuWenIMGUrls := []string{
		"https://mp.weixin.qq.com/s/hQf0N8P4vaaCaxt8OFzwfw",
		"https://mp.weixin.qq.com/s/FCuFttc5rVfxn7sPxmAsMg",
	}

	httpClientTimeout := 10 * time.Minute

	start := time.Now()
	crawlerImgSvc := NewCrawlerImgService(wxTuWenIMGUrls, httpClientTimeout, imgSavePath, textContentFilePath, textContentFileDir)
	spiderResults, err := crawlerImgSvc.RunSpiderImg()
	if err != nil {
		t.Fatalf("爬虫失败: %v", err)
	}

	castTime := time.Since(start)
	t.Logf("爬虫结束，总耗时: %s", castTime.String())
	for _, item := range spiderResults {
		if item.Err != nil {
			t.Errorf("抓取失败: Num: %d, Url: %s, Err: %v", item.Number, item.URL, item.Err)
		}
	}
}

func TestRunCropImg(t *testing.T) {
	imgSavePath := "E:\\Downloads\\wx_graph_crawl_downloads"

	start := time.Now()
	cropSvc := NewCropImgService(imgSavePath, 10, 65)
	cropResults, err := cropSvc.RunCropImg()
	if err != nil {
		t.Fatalf("裁剪失败: %v", err)
	}
	castTime := time.Since(start)
	t.Logf("裁剪结束，总耗时: %s", castTime.String())
	for _, item := range cropResults {
		if item.Err != nil {
			t.Errorf("裁剪失败: ImgPath: %s, Err: %v", item.ImgPath, item.Err)
		}
	}
}

func TestRunMoveImg(t *testing.T) {
	imgSavePath := "E:\\Downloads\\wx_graph_crawl_downloads"
	start := time.Now()
	moveSvc := NewMoveImgService(imgSavePath, 5)
	err := moveSvc.RunMoveImg()
	if err != nil {
		t.Fatalf("移动图片失败: %v", err)
	}
	castTime := time.Since(start)
	t.Logf("移动图片结束，总耗时: %s", castTime.String())
}
