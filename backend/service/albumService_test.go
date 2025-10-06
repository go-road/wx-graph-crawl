package service

import (
	"fmt"
	"go.uber.org/zap"
	"testing"
)

func TestRunSpiderAlbum(t *testing.T) {
	fmt.Println("开始执行TestRunSpiderAlbum...")
	// 示例用法
	albumHomeURL := "https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&__biz=Mzg5MzgxMTIyOQ==&scene=1&album_id=2544487917101039623&count=3#wechat_redirect"
	urls, err := GetWechatAlbumAllArticleURLs(albumHomeURL)
	if err != nil {
		fmt.Println("获取专辑文章URL失败:", err)
		zap.L().Error("获取专辑文章URL失败", zap.Error(err))
	} else {
		fmt.Printf("获取的专辑文章URL列表：%v", urls)
		zap.L().Info("成功获取所有文章URL", zap.Int("count", len(urls)))
		// 处理获取到的URL列表
		for _, url := range urls {
			fmt.Println(url)
		}
	}
}
