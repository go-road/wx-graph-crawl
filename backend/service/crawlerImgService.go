package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"github.com/pudongping/wx-graph-crawl/backend/utils"
	"go.uber.org/zap"
)

var (
	reImgURL      = regexp.MustCompile(`cdn_url: '([^']+)'`)                          // 抓取微信图片链接地址
	reContentData = regexp.MustCompile(`window\.__QMTPL_SSR_DATA__=\s*(\{.*?\})\s*;`) // 解析图文类型的一些文字信息
	reTitle       = regexp.MustCompile(`title:'(.*?)'`)                               // 抓取标题
	reDesc        = regexp.MustCompile(`desc:'(.*?)'`)                                // 抓取正文内容
)

type CrawlerImgService struct {
	WXTuWenIMGUrls      []string      // 需要被抓取的微信图文链接地址
	HttpClientTimeout   time.Duration // 网络请求超时时间
	ImgSavePath         string        // 图片保存路径
	TextContentFilePath string        // 文案保存文件地址
}

func NewCrawlerImgService(
	wxTuWenIMGUrls []string,
	httpClientTimeout time.Duration,
	imgSavePath string,
	textContentFilePath string,
) *CrawlerImgService {
	return &CrawlerImgService{
		WXTuWenIMGUrls:      wxTuWenIMGUrls,
		HttpClientTimeout:   httpClientTimeout,
		ImgSavePath:         imgSavePath,
		TextContentFilePath: textContentFilePath,
	}
}

func (svc *CrawlerImgService) RunSpiderImg() (spiderResults []types.CrawlResult, err error) {
	wg := sync.WaitGroup{}

	crawlResultChan := make(chan types.CrawlResult, len(svc.WXTuWenIMGUrls)) // 收集信息
	wg.Add(len(svc.WXTuWenIMGUrls))

	for i, wxTuWenIMGUrl := range svc.WXTuWenIMGUrls {
		go svc.work(i, wxTuWenIMGUrl, &wg, crawlResultChan)
	}

	// 等待所有子协程完成
	wg.Wait()
	close(crawlResultChan) // 关闭通道

	// 读取通道中所有的数据
	for workRes := range crawlResultChan {
		spiderResults = append(spiderResults, workRes)
	}

	// 将一些文案写入文件中
	if err := svc.WriteWenAnContent(spiderResults); err != nil {
		return spiderResults, errors.Wrap(err, "将文案写入时，出现异常")
	}

	return
}

func (svc *CrawlerImgService) work(i int, wxTuWenIMGUrl string, wg *sync.WaitGroup, crawlResultChan chan types.CrawlResult) {
	defer wg.Done()

	num := i + 1 // 标记每个子协程的序号
	var err error

	crawlRes := types.CrawlResult{
		URL:    wxTuWenIMGUrl,
		Number: num,
		Err:    nil,
	}
	// 先一个一个的抓取每一个链接地址对应的 html 内容
	html, err := svc.FetchWXHTMLContent(wxTuWenIMGUrl)
	if err != nil {
		crawlRes.Err = err
		crawlResultChan <- crawlRes
		return
	}
	crawlRes.Html = html
	// 从抓取后的 html 内容中解析出所有的图片链接地址
	imgUrls, err := svc.ParseImgUrls(html)
	if err != nil {
		crawlRes.Err = err
		crawlResultChan <- crawlRes
		return
	}
	// 批量下载图片
	imgFilePaths, err := svc.FastDownloadImgFiles(imgUrls, num)
	if err != nil {
		crawlRes.Err = err
		crawlResultChan <- crawlRes
		return
	}
	crawlRes.ImgSavePathSuccess = imgFilePaths
	// 提取想要记录的文本信息
	crawlRes.WriteContent = svc.GetWriteContent(html, num)

	crawlResultChan <- crawlRes
}

// 抓取每一个链接地址对应的 html 内容
func (svc *CrawlerImgService) FetchWXHTMLContent(wxTuWenUrl string) (string, error) {
	if "" == wxTuWenUrl {
		return "", errors.Wrap(errors.New("链接地址不能为空！"), "被抓取的链接地址不能为空！")
	}

	httpClient := &http.Client{
		Timeout: svc.HttpClientTimeout,
	}
	httpResp, err := utils.HttpGet(httpClient, wxTuWenUrl)
	if err != nil {
		return "", errors.Wrap(err, "httpGet 方法出现错误！")
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 解析 HTML 文档
	doc, err := goquery.NewDocumentFromReader(httpResp.Body)
	if err != nil {
		return "", errors.Wrap(err, "goquery 解析 HTML 文档出现错误")
	}

	html, err := doc.Html()
	if err != nil {
		return "", errors.Wrap(err, "获取 HTML 文档内容出现错误")
	}

	return html, err
}

func (svc *CrawlerImgService) ParseImgUrls(html string) ([]string, error) {
	// 提取 picture_page_info_list 内容块
	reList := regexp.MustCompile(`window\.picture_page_info_list\s*=\s*(\[[\s\S]*?\]);`)
	listMatch := reList.FindStringSubmatch(html)
	if len(listMatch) < 2 {
		return nil, errors.New("未找到 picture_page_info_list 内容")
	}
	listContent := listMatch[1]

	// 提取 cdn_url，排除 watermark_info 下的
	reImgNeed := regexp.MustCompile(`cdn_url:\s*'([^']+)'`)
	reWatermark := regexp.MustCompile(`watermark_info\s*:\s*\{[\s\S]*?cdn_url:\s*'([^']+)'[\s\S]*?\}`)

	// 先找所有 watermark_info 下的 cdn_url
	watermarkUrls := make(map[string]struct{})
	for _, wm := range reWatermark.FindAllStringSubmatch(listContent, -1) {
		if len(wm) > 1 {
			watermarkUrls[wm[1]] = struct{}{}
		}
	}

	// 再找所有 cdn_url，排除水印
	results := reImgNeed.FindAllStringSubmatch(listContent, -1)
	urls := make([]string, 0, len(results))
	for _, v := range results {
		if len(v) > 1 {
			if _, isWatermark := watermarkUrls[v[1]]; !isWatermark {
				urls = append(urls, v[1])
			}
		}
	}
	return urls, nil
}

func (svc *CrawlerImgService) FastDownloadImgFiles(imgUrls []string, num int) (imgFilePaths []string, err error) {
	savePath := fmt.Sprintf("%s/%d", svc.ImgSavePath, num) // 以批次分组成不同的文件夹
	// 先检查目录是否存在
	if err = utils.MkdirIfNotExist(savePath); err != nil {
		return nil, errors.Wrap(err, "批量下载图片时，检查目录是否存在出现错误")
	}

	// 并发下载
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, 10) // 最多同时10个并发下载
	type imgDownRes struct {
		imgUrl      string
		imgFilePath string
		err         error
	}
	filePathChan := make(chan imgDownRes, len(imgUrls))

	for i, imgUrl := range imgUrls {
		wg.Add(1)

		go func(i int, imgUrl, savePath string) {
			defer wg.Done()
			sem <- struct{}{} // 请求信号量
			// 释放信号量
			defer func() { <-sem }()

			imgFilePath := fmt.Sprintf("%s/%d.jpeg", savePath, i+1)
			// 一张一张的下载图片
			_, err := svc.DownloadImgFile(imgUrl, imgFilePath)
			filePathChan <- imgDownRes{
				imgUrl:      imgUrl,
				imgFilePath: imgFilePath,
				err:         err,
			}

		}(i, imgUrl, savePath)

	}

	wg.Wait()
	close(filePathChan)

	var errStr string
	for downloadRes := range filePathChan {
		if downloadRes.err != nil {
			errStr += "ImgUrl: " + downloadRes.imgUrl + " Err: " + downloadRes.err.Error() + " | "
		} else {
			imgFilePaths = append(imgFilePaths, downloadRes.imgFilePath)
		}
	}

	if "" != errStr {
		return nil, errors.Wrap(errors.New(errStr), "批量下载图片时，可能下载某一张图片时出现错误")
	}

	return imgFilePaths, nil
}

// 一张一张的下载图片
func (svc *CrawlerImgService) DownloadImgFile(imgUrl, imgFilePath string) (string, error) {
	httpClient := &http.Client{
		Timeout: svc.HttpClientTimeout,
	}
	httpResp, err := utils.HttpGet(httpClient, imgUrl)
	if err != nil {
		return "", errors.Wrap(err, "一张一张下载图片时，出现错误")
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 检查响应状态码
	if httpResp.StatusCode != http.StatusOK {
		return "", errors.Wrap(errors.Errorf("网络请求失败，错误码为：%d", httpResp.StatusCode), "HTTP状态码不为200")
	}

	zap.L().Info("正在下载图片", zap.String("imgFilePath", imgFilePath))

	file, err := os.Create(imgFilePath)
	if err != nil {
		return "", errors.Wrap(err, "下载图片时，创建文件失败")
	}
	defer file.Close()
	// 保存文件
	_, err = io.Copy(file, httpResp.Body)
	if err != nil {
		return "", errors.Wrap(err, "保存下载的图片文件失败")
	}

	return imgFilePath, nil
}

func (svc *CrawlerImgService) GetWriteContent(html string, num int) string {
	// 查找匹配的部分
	matches := reContentData.FindStringSubmatch(html)
	if len(matches) <= 1 {
		// 没有匹配到内容
		return ""
	}
	contentStr := matches[1] // 匹配到的内容

	// 提取 title 和 desc 的值
	// 因为提取的 jsonStr 内容中是一定会含有 title 和 desc 字段的，因此以下代码可不用做边界值的判断
	// 这里不能直接通过解析 json 字符串的方式来提取内容，因为这里的内容不是一个合法的 json 字符串，它仅仅是一个 js 代码（尤其注意）
	title := reTitle.FindStringSubmatch(contentStr)[1]
	desc := reDesc.FindStringSubmatch(contentStr)[1]

	content := fmt.Sprintf("第 %d ====> \r\n", num)
	content += "标题： " + title + "\r\n"
	content += "正文内容 --------------- \r\n " + desc + "\r\n ------------- \r\n"

	return content
}

func (svc *CrawlerImgService) WriteWenAnContent(contents []types.CrawlResult) error {
	if err := utils.CreateFileIfNotExist(svc.TextContentFilePath); err != nil {
		return errors.Wrap(err, "写入文案时，创建文本文件出现异常")
	}

	// 先按照 number 字段进行从小到大排序
	sort.Slice(contents, func(i, j int) bool {
		return contents[i].Number < contents[j].Number
	})

	result := ""
	for _, content := range contents {
		result += content.WriteContent + "\r\n\r\n"
	}

	if err := os.WriteFile(svc.TextContentFilePath, []byte(result), 0644); err != nil {
		return errors.Wrapf(err, "写入 %s 文件时，发生错误：%+v", svc.TextContentFilePath, err)
	}

	return nil
}
