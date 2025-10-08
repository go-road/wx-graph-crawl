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
	//楼仔_技术派(13)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=Mzg3OTU5NzQ1Mw==&action=getalbum&album_id=2878590662869303297&scene=126&sessionid=1759929957820#wechat_redirect",
	//楼仔_硬核技术(45)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=Mzg3OTU5NzQ1Mw==&action=getalbum&album_id=2206501082485358594&scene=126&sessionid=1759929957820#wechat_redirect",
	//youcongtech_开源项目(108)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzUxODk0ODQ3Ng==&action=getalbum&album_id=2296681499552710656&subscene=&scenenote=https%3A%2F%2Fmp.weixin.qq.com%2Fs%3F__biz%3DMzUxODk0ODQ3Ng%3D%3D%26mid%3D2247487827%26idx%3D1%26sn%3D82d44c8078a1233fb7fbc016b6fba0cd%26click_id%3D8%26key%3Ddaf9bdc5abc4e8d0cff20aa977979f507228a43e7cebe8be597e2ca86a64ad5931455889d76cff68efad850146f7d574e787118e5bbd68a1cf670365caf10240825744bc2cf2f6c88b9ac514c1d93c8ac086b716e2947b5dcf4f1811fa2e2e648172b71c596faf499f0482785910f63bd5e2ceac5ccf81aff0ef7a27c473b626%26ascene%3D1%26uin%3DMTA1MzE0OTc4MA%253D%253D%26devicetype%3DUnifiedPCWindows%26version%3Df2541022%26lang%3Dzh_CN%26countrycode%3DCN%26exportkey%3Dn_ChQIAhIQsJeKe6nDL%252BlHh7GBt1XeShLbAQIE97dBBAEAAAAAAAHCDSVAds8AAAAOpnltbLcz9gKNyK89dVj0kDE1KVZK1pB0LHviZtLy%252Bxv7bfofGhhmd%252F%252F390cz5u8Psi1pK%252B%252Bhh2o33f27iyJ3%252F2LFuTc5kFqb%252Bjv09PACgp19fQH2otqguS89l8cb1D0s8HOKlm41gnQgX0yBV3MDAGJLWX23tzNCN3WXbzSq5UZ9wg3EaRI1R%252Fa2KX7tEd1NDN%252FisjivQYcWkJMc3YMdPaG8rqPmWBM2yzJrqcKXoszf3SkRa15Rmo67AMy9DbgUgoyC8A%253D%253D%26acctmode%3D0%26pass_ticket%3Djx21yd3DAidw0RbvIFEmal%252FbKeECf62LIaU%252FbSRqGtq9%252Bb4j%252FiLN01Qytqz3GB8r%26wx_header%3D0%26fasttmpl_type%3D0%26fasttmpl_fullversion%3D7931669-zh_CN-html%26from_xworker%3D1&nolastread=1&sessionid=#wechat_redirect",
	//youcongtech_架构(97)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzUxODk0ODQ3Ng==&action=getalbum&album_id=2174448056010604548&scene=173&subscene=&sessionid=undefined&enterid=0&from_msgid=2247487827&from_itemidx=1&count=3&nolastread=1#wechat_redirect",
	//youcongtech_分布式(95)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzUxODk0ODQ3Ng==&action=getalbum&album_id=2164900987990245380&scene=173&subscene=&sessionid=undefined&enterid=0&from_msgid=2247487827&from_itemidx=1&count=3&nolastread=1#wechat_redirect",
	//youcongtech_微服务(90)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzUxODk0ODQ3Ng==&action=getalbum&album_id=2164900987587592201&scene=173&subscene=&sessionid=undefined&enterid=0&from_msgid=2247487827&from_itemidx=1&count=3&nolastread=1#wechat_redirect",
	//youcongtech_框架(99)
	"https://mp.weixin.qq.com/mp/appmsgalbum?__biz=MzUxODk0ODQ3Ng==&action=getalbum&album_id=2164900987554037763&scene=173&subscene=&sessionid=undefined&enterid=0&from_msgid=2247487827&from_itemidx=1&count=3&nolastread=1#wechat_redirect",
}

func TestRunSpiderAlbum(t *testing.T) {
	fmt.Println("开始执行TestRunSpiderAlbum...")
	// 示例用法
	albumHomeURL := urls[len(urls)-1]
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
