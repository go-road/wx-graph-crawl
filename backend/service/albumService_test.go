package service

import (
	"fmt"
	"go.uber.org/zap"
	"testing"
)

var urls []string = []string{
	//一安未来_干货分享集(490)
	"https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&__biz=Mzg5MzgxMTIyOQ==&scene=1&album_id=2544487917101039623&count=3#wechat_redirect",
	//一安未来_第三方工具(32)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=Mzg5MzgxMTIyOQ==&action=getalbum&album_id=2544480919240359937&scene=173&subscene=&sessionid=svr_61c1fe9a7f8&enterid=1759862015&from_msgid=2247501849&from_itemidx=1&count=3&nolastread=1#wechat_redirect",
	//一安未来_面试专题集(66)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=Mzg5MzgxMTIyOQ==&action=getalbum&album_id=2544478198781231106&scene=173&subscene=&sessionid=svr_61c1fe9a7f8&enterid=1759862015&from_msgid=2247501849&from_itemidx=1&count=3&nolastread=1#wechat_redirect",
}

func TestRunSpiderAlbum(t *testing.T) {
	fmt.Println("开始执行TestRunSpiderAlbum...")
	// 示例用法
	albumHomeURL := urls[0]
	urls, articleInfos, err := GetWechatAlbumAllArticleURLs(albumHomeURL)
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
		// 使用文章详细信息（可用于跟踪下载状态）
		for i := range articleInfos {
			// 处理文章信息或更新状态
			// articleInfos[i].Status = "已下载"
			fmt.Printf("文章 %d: %s, URL: %s, 状态: %s\n", articleInfos[i].Index, articleInfos[i].Title, articleInfos[i].URL, articleInfos[i].Status)
		}
	}
}
