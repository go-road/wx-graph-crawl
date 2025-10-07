package service

import (
	"crypto/md5"
	"fmt"
	"github.com/labstack/gommon/log"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"github.com/pudongping/wx-graph-crawl/backend/utils"
	"go.uber.org/zap"
)

var (
	/**
	对于HTML解析，建议使用专门的HTML解析库如 goquery 而不是正则表达式，因为正则表达式在处理嵌套结构时容易出现问题。
	因为对于HTML解析任务，使用专门的HTML解析库（如 goquery）比正则表达式更加可靠和易于维护。
	Go语言的正则表达式引擎不支持Perl风格的负向先行断言 (?! 语法。Go使用的是RE2正则表达式引擎，它不支持某些高级的Perl正则表达式特性。
	*/

	reImgURL = regexp.MustCompile(`cdn_url: '([^']+)'`) // 抓取微信图片链接地址
	//reContentData = regexp.MustCompile(`window\.__QMTPL_SSR_DATA__=\s*(\{.*?\})\s*;`)   // 解析图文类型的一些文字信息

	// 匹配 id 为 js_article 的 div 容器
	//reContentData = regexp.MustCompile(`<div[^>]*id="js_article"[^>]*>([\s\S]*?)</div>`)
	//reContentData = regexp.MustCompile(`<div[^>]*id="js_article"[^>]*>[\s\S]*?</div>`)

	// 更精确的匹配方式，考虑嵌套的div标签
	reContentData = regexp.MustCompile(`<div[^>]*id="js_article"[^>]*>([\s\S]*?)</div>`)

	// 使用贪婪匹配，但需要确保HTML结构相对简单
	//reContentData = regexp.MustCompile(`<div[^>]*id="js_article"[^>]*>[\s\S]*</div>`)

	// 使用更复杂的正则表达式来正确匹配嵌套的div结构
	//reContentData = regexp.MustCompile(`<div[^>]*id="js_article"[^>]*>([\s\S]*?)</div\s*>`)

	// 匹配 section 标签内容
	//reContentData = regexp.MustCompile(`<section[^>]*>([\s\S]*?)</section>`)

	// 或者匹配特定 class 的 div 容器
	//reContentData = regexp.MustCompile(`<div[^>]*class="[^"]*rich_media_content[^"]*"[^>]*>([\s\S]*?)</div>`)

	reTitle = regexp.MustCompile(`<meta\s+property="og:title"\s+content="(.*?)"`) // 抓取标题
	reDesc  = regexp.MustCompile(`<meta\s+name="description"\s+content="(.*?)"`)  // 抓取正文内容
)

// 定义全局的资源文件映射，用于跟踪已下载的资源
var (
	// 用于线程安全的资源映射访问
	resourceMutex sync.Mutex
	// 存储已下载的资源文件路径，key为原始URL，value为本地文件路径
	downloadedResources = make(map[string]string)
)

type CrawlerImgService struct {
	WXTuWenIMGUrls      []string      // 需要被抓取的微信图文链接地址
	HttpClientTimeout   time.Duration // 网络请求超时时间
	ImgSavePath         string        // 图片保存路径
	TextContentFilePath string        // 文案保存文件地址（所有文案保存到一个文件中）
	TextContentFileDir  string        // 文案保存文件目录（每个文章文案保存到一个文件中）
	CommonJSDir         string        // JS公共目录
	CommonCSSDir        string        // CSS公共目录
	FailedDownloadDir   string        // 下载失败地址保存目录
}

func NewCrawlerImgService(
	wxTuWenIMGUrls []string,
	httpClientTimeout time.Duration,
	imgSavePath string,
	textContentFilePath string,
	textContentFileDir string,
) *CrawlerImgService {
	// 初始化公共目录路径
	commonJSDir := filepath.Join(imgSavePath, "js")
	commonCSSDir := filepath.Join(imgSavePath, "css")
	failedDownloadDir := filepath.Join(imgSavePath, "failed_downloads")

	// 确保公共目录存在
	utils.MkdirIfNotExist(commonJSDir)
	utils.MkdirIfNotExist(commonCSSDir)
	utils.MkdirIfNotExist(failedDownloadDir)

	return &CrawlerImgService{
		WXTuWenIMGUrls:      wxTuWenIMGUrls,
		HttpClientTimeout:   httpClientTimeout,
		ImgSavePath:         imgSavePath,
		TextContentFilePath: textContentFilePath,
		TextContentFileDir:  textContentFileDir,
		CommonJSDir:         commonJSDir,
		CommonCSSDir:        commonCSSDir,
		FailedDownloadDir:   failedDownloadDir,
	}
}

func (svc *CrawlerImgService) RunSpiderImg() (spiderResults []types.CrawlResult, err error) {
	wg := sync.WaitGroup{}

	crawlResultChan := make(chan types.CrawlResult, len(svc.WXTuWenIMGUrls)) // 收集信息
	wg.Add(len(svc.WXTuWenIMGUrls))

	for i, wxTuWenIMGUrl := range svc.WXTuWenIMGUrls {
		// 随机休眠1~10秒，避免被封 => http2: Transport received GOAWAY from server ErrCode:COMPRESSION_ERROR
		d := time.Millisecond * time.Duration(utils.GenRandomNumber(1000, 10000))
		fmt.Printf("准备随机休眠：%v", d)
		time.Sleep(d)
		log.Info("开始抓取", zap.String("链接地址", wxTuWenIMGUrl), zap.Int("序号", i+1))
		go svc.work(i, wxTuWenIMGUrl, &wg, crawlResultChan)
	}

	// 等待所有子协程完成
	wg.Wait()
	close(crawlResultChan) // 关闭通道

	// 读取通道中所有的数据
	for workRes := range crawlResultChan {
		spiderResults = append(spiderResults, workRes)
	}

	// 将文案写入文件中
	if err := svc.WriteWenAnContent(spiderResults); err != nil {
		return spiderResults, errors.Wrap(err, "将文案写入时，出现异常")
	}

	// 为每篇文章生成Word文档
	if err := NewWordService(svc.TextContentFileDir).GenerateWordForEachArticle(spiderResults); err != nil {
		return spiderResults, errors.Wrap(err, "生成Word文档时出现异常")
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
		Title:  "未命名标题",
	}
	// 先一个一个的抓取每一个链接地址对应的 html 内容
	html, err := svc.FetchWXHTMLContent(wxTuWenIMGUrl)
	if err != nil {
		crawlRes.Err = err
		crawlResultChan <- crawlRes
		// 记录文章抓取失败
		svc.SaveFailedDownloadUrl(wxTuWenIMGUrl, fmt.Sprintf("抓取文章失败: %v", err))
		return
	}
	log.Info("抓取微信图片链接地址成功", zap.String("链接地址", wxTuWenIMGUrl), zap.Int("序号", num))
	log.Info("微信HTML内容长度：", zap.Int("HTML内容长度", len(html)))
	crawlRes.Html = html
	// 从抓取后的 html 内容中解析出所有的图片链接地址
	imgUrls, err := svc.ParseImgUrls(html)
	if err != nil {
		crawlRes.Err = err
		crawlResultChan <- crawlRes
		return
	}
	// 批量下载图片(GetWriteContent函数已下载图片，这里无需重复下载)
	/*
		imgFilePaths, err := svc.FastDownloadImgFiles(imgUrls, num)
		if err != nil {
			crawlRes.Err = err
			crawlResultChan <- crawlRes
			return
		}
		crawlRes.ImgSavePathSuccess = imgFilePaths
	*/
	crawlRes.ImgSavePathSuccess = imgUrls // 模拟返回图片保存成功的图片地址
	//zap.L().Info("无需重复下载的图片数量：", zap.Int("count", len(imgUrls)))
	zap.L().Info(fmt.Sprintf("无需重复下载的图片数量：%d", len(imgUrls)))
	// 提取想要记录的内容
	crawlRes.Title, crawlRes.WriteContent = svc.GetWriteContent(html, num)

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
	// 提取 window.picture_page_info_list 内容块
	reList := regexp.MustCompile(`window\.picture_page_info_list\s*=\s*(\[[\s\S]*?\]);`)
	listMatch := reList.FindStringSubmatch(html)
	if len(listMatch) < 2 {
		return []string{}, nil // 返回空列表
		//return nil, errors.New("未找到 picture_page_info_list 内容")
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
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(imgUrl, fmt.Sprintf("下载图片时出现错误: %v", err))
		return "", errors.Wrap(err, "一张一张下载图片时，出现错误")
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 检查响应状态码
	if httpResp.StatusCode != http.StatusOK {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(imgUrl, fmt.Sprintf("网络请求失败，错误码为：%d", httpResp.StatusCode))
		return "", errors.Wrap(errors.Errorf("网络请求失败，错误码为：%d", httpResp.StatusCode), "HTTP状态码不为200")
	}

	zap.L().Info("正在下载图片", zap.String("imgFilePath", imgFilePath))

	file, err := os.Create(imgFilePath)
	if err != nil {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(imgUrl, fmt.Sprintf("创建文件失败: %v", err))
		return "", errors.Wrap(err, "下载图片时，创建文件失败")
	}
	defer file.Close()
	// 保存文件
	_, err = io.Copy(file, httpResp.Body)
	if err != nil {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(imgUrl, fmt.Sprintf("保存文件失败: %v", err))
		return "", errors.Wrap(err, "保存下载的图片文件失败")
	}

	return imgFilePath, nil
}

// 下载CSS和JS等资源文件
func (svc *CrawlerImgService) DownloadResourceFile(resourceUrl, resourceFilePath string) (string, error) {
	// 检查文件是否已存在
	if _, err := os.Stat(resourceFilePath); err == nil {
		zap.L().Info("资源文件已存在，跳过下载", zap.String("resourceFilePath", resourceFilePath))
		return resourceFilePath, nil
	}

	httpClient := &http.Client{
		Timeout: svc.HttpClientTimeout,
	}
	httpResp, err := utils.HttpGet(httpClient, resourceUrl)
	if err != nil {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(resourceUrl, fmt.Sprintf("下载资源文件时出现错误: %v", err))
		return "", errors.Wrap(err, "下载资源文件时，出现错误")
	}
	if httpResp != nil {
		defer httpResp.Body.Close()
	}

	// 检查响应状态码
	if httpResp.StatusCode != http.StatusOK {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(resourceUrl, fmt.Sprintf("网络请求失败，错误码为：%d", httpResp.StatusCode))
		return "", errors.Wrap(errors.Errorf("网络请求失败，错误码为：%d", httpResp.StatusCode), "HTTP状态码不为200")
	}

	zap.L().Info("正在下载资源文件", zap.String("resourceFilePath", resourceFilePath))

	// 确保目录存在
	dir := filepath.Dir(resourceFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(resourceUrl, fmt.Sprintf("创建资源文件目录失败: %v", err))
		return "", errors.Wrap(err, "创建资源文件目录失败")
	}

	file, err := os.Create(resourceFilePath)
	if err != nil {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(resourceUrl, fmt.Sprintf("创建文件失败: %v", err))
		return "", errors.Wrap(err, "下载资源文件时，创建文件失败")
	}
	defer file.Close()
	// 保存文件
	_, err = io.Copy(file, httpResp.Body)
	if err != nil {
		// 记录失败的下载地址
		svc.SaveFailedDownloadUrl(resourceUrl, fmt.Sprintf("保存文件失败: %v", err))
		return "", errors.Wrap(err, "保存下载的资源文件失败")
	}

	return resourceFilePath, nil
}

// 从URL中提取基础文件名（如从appmsg.mg0vycs343acb927.js提取appmsg.js）
func extractBaseFilename(url string) string {
	// 获取URL中的文件名部分
	filename := filepath.Base(url)

	// 正则表达式匹配文件名中的基础部分（如appmsg.mg0vycs343acb927.js中的appmsg）
	baseNameRegex := regexp.MustCompile(`^([a-zA-Z0-9_]+)`)
	matches := baseNameRegex.FindStringSubmatch(filename)
	if len(matches) > 1 {
		// 获取文件扩展名
		ext := filepath.Ext(filename)
		return matches[1] + ext
	}

	return filename
}

// 获取资源文件的公共路径，如果文件已存在则直接返回路径（绝对路径和相对路径）
func (svc *CrawlerImgService) getCommonResourcePath(url, fileType string) (string, string) {
	resourceMutex.Lock()
	defer resourceMutex.Unlock()

	var commonDir, relativePath string
	// 根据文件类型确定保存目录
	if fileType == "js" {
		commonDir = svc.CommonJSDir
	} else if fileType == "css" {
		commonDir = svc.CommonCSSDir
	} else {
		// 默认为原路径逻辑
		return url, url
	}

	// 提取基础文件名
	baseFilename := extractBaseFilename(url)
	// 构建完整的本地路径
	fullPath := filepath.Join(commonDir, baseFilename)

	// 检查资源是否已经下载过
	if relativePath, exists := downloadedResources[url]; exists {
		return fullPath, relativePath
	}

	if fileType == "js" {
		relativePath = filepath.Join("js", baseFilename)
	} else if fileType == "css" {
		relativePath = filepath.Join("css", baseFilename)
	}

	// 记录已下载的资源
	downloadedResources[url] = relativePath

	return fullPath, relativePath
}

func (svc *CrawlerImgService) GetWriteContent(html string, num int) (title string, content string) {
	// 提取 title 和 desc 的值
	// 因为提取的 jsonStr 内容中是一定会含有 title 和 desc 字段的，因此以下代码可不用做边界值的判断
	// 这里不能直接通过解析 json 字符串的方式来提取内容，因为这里的内容不是一个合法的 json 字符串，它仅仅是一个 js 代码（尤其注意）
	titleMatch := reTitle.FindStringSubmatch(html)
	descMatch := reDesc.FindStringSubmatch(html)

	if len(titleMatch) < 2 || len(descMatch) < 2 {
		zap.L().Error("未找到标题或描述信息")
		//return title, "未找到标题或描述信息"
		return title, html
	}
	for _, titleStr := range titleMatch {
		zap.L().Info("匹配到的标题内容：" + titleStr)
	}

	title = titleMatch[1]
	desc := descMatch[1]

	// 清理标题，使其适合作为文件名
	title = sanitizeFilename(title)

	// 使用清理后的标题创建文件路径，保存 html 文件
	filePath := fmt.Sprintf("%s/%s.html", svc.ImgSavePath, title)
	zap.L().Info("保存 html 文件供后续分析，文件路径为：" + filePath)

	// 创建goquery文档对象用于解析HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		zap.L().Error("解析 HTML 时出现错误", zap.Error(err))
		return title, "解析 HTML 时出现错误"
	} else {
		// 设置HTML文档的title标签
		titleTag := doc.Find("title")
		if titleTag.Length() > 0 {
			// 如果已经有title标签，则更新内容
			titleTag.SetText(title)
		} else {
			// 如果没有title标签，则在head中添加一个新的title标签
			head := doc.Find("head")
			if head.Length() > 0 {
				head.AppendHtml(fmt.Sprintf("<title>%s</title>", title))
			}
		}

		// 找到所有的img标签并下载图片，然后更新其data-src和src属性
		// 创建资源目录
		resourceDir := fmt.Sprintf("%s/%s", svc.ImgSavePath, title)
		if err := os.MkdirAll(resourceDir, 0755); err != nil {
			zap.L().Error("创建资源目录失败", zap.Error(err))
		} else {
			/**
				- 所有图片资源文件都保存在与HTML文件同名的子目录中
			    - 图片命名格式：`文件名/图片序号.jpeg`
			    - CSS命名格式：`CSS公共目录/xxx.css`
			    - JS命名格式：`JS公共目录/xxx.js`
			以//开头的相对协议URL被正确转换为带有https:前缀的绝对URL
			*/
			// 1. 处理图片文件
			// 下载每个图片并更新路径
			doc.Find("img").Each(func(i int, selection *goquery.Selection) {
				// 尝试获取data-src属性
				dataSrc, exists := selection.Attr("data-src")
				if exists && strings.Contains(dataSrc, "http") {
					// 构建本地图片路径：文件名/图片序号
					localImgPath := fmt.Sprintf("%s/%d.jpeg", title, i)
					fullImgPath := fmt.Sprintf("%s/%s", svc.ImgSavePath, localImgPath)

					// 下载图片
					if _, err := svc.DownloadImgFile(dataSrc, fullImgPath); err != nil {
						zap.L().Error("下载图片失败", zap.String("imgUrl", dataSrc), zap.Error(err))
					} else {
						// 同时更新data-src和src属性为本地路径
						selection.SetAttr("data-src", localImgPath)
						selection.SetAttr("src", localImgPath)
						// 添加日志记录，确认属性被设置
						zap.L().Info("设置图片属性",
							zap.String("data-src", localImgPath),
							zap.String("src", localImgPath))
					}
				}
			})

			// 2. 处理CSS文件
			doc.Find("link[rel='stylesheet']").Each(func(i int, selection *goquery.Selection) {
				// 获取href属性
				href, exists := selection.Attr("href")
				zap.L().Info("处理CSS文件", zap.String("href", href))
				// 检查包含"http"的链接和使用相对协议URL（以//开头）的CSS链接
				if exists {
					// 处理相对协议URL (以//开头)
					if strings.HasPrefix(href, "//") {
						href = "https:" + href // 或 "http:"，建议使用https
					}
					if strings.Contains(href, "http") {
						// 构建本地CSS路径
						//localCssPath := fmt.Sprintf("%s/style_%d.css", title, i)
						//fullCssPath := fmt.Sprintf("%s/%s", svc.ImgSavePath, localCssPath)

						// 获取CSS文件的公共路径
						fullCssPath, relativeCssPath := svc.getCommonResourcePath(href, "css")

						// 下载CSS文件（如果不存在）
						if _, err := svc.DownloadResourceFile(href, fullCssPath); err != nil {
							zap.L().Error("下载CSS文件失败", zap.String("cssUrl", href), zap.Error(err))
						} else {
							// 更新href属性为本地路径
							selection.SetAttr("href", relativeCssPath)
							// 对于小型CSS和JS文件，可以考虑直接内联到HTML中，减少文件数量
						}
					}
				}
			})

			// 3. 处理JS文件
			doc.Find("script[src]").Each(func(i int, selection *goquery.Selection) {
				// 获取src属性
				src, exists := selection.Attr("src")
				zap.L().Info("处理JS文件", zap.String("src", src))
				if exists {
					// 处理相对协议URL (以//开头)
					if strings.HasPrefix(src, "//") {
						src = "https:" + src // 或 "http:"，建议使用https
					}
					if strings.Contains(src, "http") {
						// 构建本地JS路径
						//localJsPath := fmt.Sprintf("%s/script_%d.js", title, i)
						//fullJsPath := fmt.Sprintf("%s/%s", svc.ImgSavePath, localJsPath)

						// 获取JS文件的公共路径
						fullJsPath, relativeJsPath := svc.getCommonResourcePath(src, "js")

						// 下载JS文件（如果不存在）
						if _, err := svc.DownloadResourceFile(src, fullJsPath); err != nil {
							zap.L().Error("下载JS文件失败", zap.String("jsUrl", src), zap.Error(err))
						} else {
							// 更新src属性为本地路径
							selection.SetAttr("src", relativeJsPath)
						}
					}
				}
			})

			// 获取更新后的HTML内容
			updatedHtml, err := doc.Html()
			if err == nil {
				html = updatedHtml
			} else {
				zap.L().Error("获取更新后的HTML内容失败", zap.Error(err))
			}
		}
	}

	/*zap.L().Info("保存 html 文件供后续分析",
	zap.String("reason", "content data not found"),
	zap.String("save_path", svc.ImgSavePath),
	zap.Int("number", num),
	zap.String("file_path", filePath))*/
	if err := utils.SaveFile(html, filePath); err != nil {
		zap.L().Error("保存html 文件时，出现错误", zap.Error(err))
	}

	// 查找 id 为 js_article 的 div
	articleDiv := doc.Find("div#js_article")
	if articleDiv.Length() == 0 {
		zap.L().Error("未找到 id 为 js_article 的 div", zap.String("file_path", filePath))
		return title, "未找到匹配的内容"
	}

	// 获取 div 内的 HTML 内容
	contentStr, err := articleDiv.Html()
	if err != nil {
		zap.L().Error("获取文章内容时出现错误", zap.Error(err))
		return title, "获取文章内容时出现错误"
	}

	//zap.L().Info("匹配到的内容：" + contentStr)
	zap.L().Info("匹配到的内容长度：" + fmt.Sprintf("%d", len(contentStr)))

	// 提取 section 和 span 标签的文本内容
	extractedContent := svc.ExtractArticleContent(contentStr)
	zap.L().Info("提取到的文本内容长度：" + fmt.Sprintf("%d", len(extractedContent)))

	content = fmt.Sprintf("第 %d 篇文章====> \r\n", num)
	content += "标题： " + title + "\r\n"
	content += "描述： " + desc + "\r\n"
	content += "正文内容 --------------- \r\n " + extractedContent + "\r\n ------------- \r\n"

	//zap.L().Info("文案内容：\n" + content)
	return title, content
}

// ExtractArticleContent 从HTML内容中提取section和span标签的文本内容
func (svc *CrawlerImgService) ExtractArticleContent(htmlContent string) string {
	// 使用 goquery 进一步解析文章内容，提取 section 和 span 标签文本
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		zap.L().Error("解析文章内容时出现错误", zap.Error(err))
		return "解析文章内容时出现错误"
	}

	// 提取所有 section 标签的文本内容
	var contentBuilder strings.Builder

	// 提取所有 section 标签的文本内容
	doc.Find("section").Each(func(i int, selection *goquery.Selection) {
		text := strings.TrimSpace(selection.Text())
		if text != "" {
			contentBuilder.WriteString(text)
			contentBuilder.WriteString("\n")
		}
	})

	// 提取所有 span 标签的文本内容
	doc.Find("span").Each(func(i int, selection *goquery.Selection) {
		text := strings.TrimSpace(selection.Text())
		if text != "" {
			contentBuilder.WriteString(text)
			contentBuilder.WriteString("\n")
		}
	})

	// 返回拼接后的文本内容
	return contentBuilder.String()
}

// WriteWenAnContent 写入文案内容到文件中
func (svc *CrawlerImgService) WriteWenAnContent(contents []types.CrawlResult) error {
	// 一、所有文案保存到一个文件中
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

	// 二、每个文章文案保存到一个文件中
	if err := utils.MkdirIfNotExist(svc.TextContentFileDir); err != nil {
		return errors.Wrap(err, "创建文本文件保存目录出现异常")
	}
	for _, content := range contents {
		filePath := filepath.Join(svc.TextContentFileDir, fmt.Sprintf("%d_%s.txt", content.Number, content.Title))
		if err := utils.CreateFileIfNotExist(filePath); err != nil {
			return errors.Wrapf(err, "写入 %s 文件时，发生错误：%+v", filePath, err)
		}

		if err := os.WriteFile(filePath, []byte(content.WriteContent), 0644); err != nil {
			return errors.Wrapf(err, "写入 %s 文件时，发生错误：%+v", filePath, err)
		}
	}

	return nil
}

// 辅助函数来清理文件名
func sanitizeFilename(filename string) string {
	// 移除非字母数字和常见符号的字符
	// 注意：在Go的正则表达式中，使用\p{Han}来匹配所有汉字
	// reg := regexp.MustCompile(`[^a-zA-Z0-9\-\_\.\u4e00-\u9fa5]+`)
	reg := regexp.MustCompile(`[^a-zA-Z0-9\-\_\.\p{Han}]+`)
	cleanName := reg.ReplaceAllString(filename, "_")
	// 限制文件名长度
	if len(cleanName) > 100 {
		cleanName = cleanName[:100]
	}
	return cleanName
}

// SaveFailedDownloadUrl 用于保存失败的下载地址到单独的文本文件
func (svc *CrawlerImgService) SaveFailedDownloadUrl(url string, errMsg string) {
	// 生成唯一的文件名，使用时间戳和URL的一部分
	urlHash := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.txt", timestamp, urlHash[:8])
	filepath := filepath.Join(svc.FailedDownloadDir, filename)

	// 创建文件并写入失败信息
	file, err := os.Create(filepath)
	if err != nil {
		zap.L().Error("创建失败下载记录文件失败", zap.String("filename", filename), zap.Error(err))
		return
	}
	defer file.Close()

	// 写入URL和错误信息，增加类型标识
	recordType := "资源文件"
	if strings.Contains(url, "mp.weixin.qq.com") || strings.Contains(errMsg, "抓取文章") {
		recordType = "微信文章"
	}

	_, err = file.WriteString(fmt.Sprintf("类型: %s\nURL: %s\nError: %s\nTimestamp: %s\n",
		recordType,
		url,
		errMsg,
		time.Now().Format("2006-01-02 15:04:05")))
	if err != nil {
		zap.L().Error("写入失败下载记录文件失败", zap.String("filename", filename), zap.Error(err))
	} else {
		zap.L().Info("成功记录失败下载地址", zap.String("filename", filename))
	}
}
