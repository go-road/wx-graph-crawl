package service

import (
	"encoding/json"
	"fmt"
	"github.com/pudongping/wx-graph-crawl/backend/types"
	"github.com/pudongping/wx-graph-crawl/backend/utils"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// AlbumArticle 表示专辑中的单篇文章结构
type AlbumArticle struct {
	Title      string `json:"title"`
	URL        string `json:"url"`
	MsgID      string `json:"msgid"`
	CreateTime string `json:"create_time"`
}

// CgiData 表示专辑首页接口返回的HTML中的window.cgiData结构
type CgiData struct {
	Ret          string         `json:"ret"`
	ArticleList  []AlbumArticle `json:"articleList"`
	ContinueFlag int            `json:"continue_flag"`
	Title        string         `json:"title"`
	Desc         string         `json:"desc"`
	NickName     string         `json:"nick_name"`
	ArticleCount int            `json:"article_count"`
}

// AlbumResponse 表示后续循环专辑接口返回的完整的JSON响应结构
type AlbumResponse struct {
	BaseResp struct {
		Ret int `json:"ret"`
	} `json:"base_resp"`
	GetAlbumResp struct {
		ArticleList         []AlbumArticle `json:"article_list"`
		ContinueFlag        string         `json:"continue_flag"`
		ReverseContinueFlag string         `json:"reverse_continue_flag"`
	} `json:"getalbum_resp"`
}

// GetWechatAlbumAllArticleURLs 获取微信公众号专辑中所有文章的URL列表
// albumHomeURL: 专辑首页地址，如 https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&__biz=Mzg5MzgxMTIyOQ==&scene=1&album_id=2544487917101039623&count=3#wechat_redirect
// 返回值: 所有文章的URL列表、文章详细信息列表和可能的错误
func GetWechatAlbumAllArticleURLs(albumHomeURL string) ([]string, []types.AlbumArticleInfo, error) {
	// 存储所有文章URL
	var allArticleURLs []string
	var uniqueUrls = make(map[string]struct{})   // 已去重的文章 URL
	var allArticleInfos []types.AlbumArticleInfo // 存储所有文章的详细信息
	var lastMsgID string                         // 最后一个msgid，用于分页
	continueFlag := "0"                          // 0表示没有更多数据
	requestCount := 0                            // 请求计数
	lastArticleCount := 0                        // 最后一次统计的文章数量
	const maxRequests = 100                      // 最大请求次数，防止死循环
	const requestInterval = 2 * time.Second      // 请求间隔，避免频率限制

	// 创建HTTP客户端
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	// 1. 首先获取专辑首页HTML，解析window.cgiData对象
	zap.L().Info("开始获取专辑首页HTML", zap.String("url", albumHomeURL))
	htmlContent, err := utils.HttpGetBody(httpClient, albumHomeURL)
	if err != nil {
		return nil, nil, fmt.Errorf("获取专辑首页失败: %v", err)
	}

	// 解析HTML中的window.cgiData对象
	zap.L().Info("开始解析HTML中的window.cgiData对象")
	// 提取<script>标签中的JSON数据
	dataRegex := regexp.MustCompile(`window\.cgiData\s*=\s*({[\s\S]*?});`)
	matches := dataRegex.FindStringSubmatch(htmlContent)
	if len(matches) < 2 {
		return nil, nil, fmt.Errorf("未能从HTML中提取window.cgiData对象")
	}
	cgiStr := matches[1]
	fmt.Println("window.cgiData对象内容:", cgiStr)

	// 清理JavaScript对象字符串，转换为有效的JSON字符串
	jsonStr, err := cleanJavaScriptObjectV4(cgiStr)
	if err != nil {
		zap.L().Error("清理JavaScript对象失败", zap.Error(err))
		return nil, nil, fmt.Errorf("清理JavaScript对象失败: %v", err)
	}
	zap.L().Info("清理后的JSON字符串", zap.String("content", jsonStr))

	// 解析首页打开请求的JSON数据
	var indexResp CgiData
	if err := json.Unmarshal([]byte(jsonStr), &indexResp); err != nil {
		fmt.Println("解析window.cgiData对象失败:", err)
		return nil, nil, fmt.Errorf("解析window.cgiData对象失败: %v", err)
	}
	zap.L().Info("解析到专辑信息",
		zap.String("title", indexResp.Title),
		zap.String("desc", indexResp.Desc),
		zap.String("nick_name", indexResp.NickName),
		//zap.Int("article_count", len(indexResp.ArticleList)),
		// 文章总数量
		zap.Int("article_count", indexResp.ArticleCount),
	)

	// 处理初始文章列表
	for _, article := range indexResp.ArticleList {
		// 修复URL中的特殊字符
		fixedURL := strings.ReplaceAll(article.URL, "&amp;", "&")
		if _, exists := uniqueUrls[fixedURL]; !exists {
			uniqueUrls[fixedURL] = struct{}{}
			allArticleURLs = append(allArticleURLs, fixedURL)
			// 文章详细信息
			articleInfo := types.AlbumArticleInfo{
				Index:  len(allArticleInfos) + 1,
				Title:  article.Title,
				URL:    fixedURL,
				Status: "未下载", // 初始状态
			}
			allArticleInfos = append(allArticleInfos, articleInfo)
		}
		// 更新最后一个msgid
		lastMsgID = article.MsgID
	}

	// 2. 解析专辑首页URL获取必要参数
	params, err := parseAlbumHomeURL(albumHomeURL) // map结构中包含__biz, album_id
	if err != nil {
		//return allArticleURLs, errors.Wrap(err, "解析首页URL参数失败")
		return allArticleURLs, allArticleInfos, fmt.Errorf("解析首页URL参数失败: %v", err)
	}

	// 检查是否需要继续请求（continue_flag为1表示还有更多文章）
	if indexResp.ContinueFlag == 1 {
		continueFlag = "1"
	}

	// 3. 循环请求获取剩余文章，直到没有更多数据或达到最大请求次数
	for continueFlag == "1" && requestCount < maxRequests {
		requestCount++
		fmt.Printf("正在进行第%d次请求\n", requestCount)
		zap.L().Info(fmt.Sprintf("正在进行第%d次请求", requestCount))
		// 构建API请求URL
		apiURL := buildAlbumAPIURL(params, lastMsgID)
		zap.L().Info("开始请求专辑文章列表API", zap.String("url", apiURL))

		// 发送请求并解析响应
		fmt.Println("正在发送GET请求：", apiURL)
		response, err := utils.HttpGetBody(httpClient, apiURL)
		if err != nil {
			zap.L().Error("请求专辑API失败，停止获取", zap.Error(err))
			return allArticleURLs, allArticleInfos, errors.Wrap(err, "请求专辑文章列表失败")
		}

		// 解析后续请求的JSON响应
		var albumResp AlbumResponse
		if err := json.Unmarshal([]byte(response), &albumResp); err != nil {
			zap.L().Error("解析专辑API响应失败", zap.Error(err))
			return allArticleURLs, allArticleInfos, errors.Wrap(err, "解析专辑文章列表响应失败")
		}

		// 发送请求并解析响应（另一种方式）
		/*
			resp, err := utils.HttpGet(httpClient, apiURL)
			if err != nil {
				zap.L().Error("请求失败，停止获取", zap.Error(err))
				return allArticleURLs, errors.Wrap(err, "请求专辑文章列表失败")
			}
			decoder := json.NewDecoder(resp.Body)
			if err := decoder.Decode(&albumResp); err != nil {
				zap.L().Error("解析失败，停止获取", zap.Error(err))
				return allArticleURLs, errors.Wrap(err, "解析专辑文章列表响应失败")
			}*/

		// 检查响应是否成功
		if albumResp.BaseResp.Ret != 0 {
			zap.L().Error("专辑API返回错误", zap.Int("ret", albumResp.BaseResp.Ret))
			return allArticleURLs, allArticleInfos, errors.Errorf("请求专辑文章列表返回错误码: %d", albumResp.BaseResp.Ret)
		}

		// 提取文章URL
		articleList := albumResp.GetAlbumResp.ArticleList
		// 如果没有新文章，停止请求
		if len(articleList) == 0 {
			zap.L().Info("本次请求未获取到新文章，停止循环")
			break
		}

		// 处理文章列表
		newArticleFound := false
		for _, article := range articleList {
			// 修复URL中的特殊字符
			//fixedURL := strings.Replace(article.URL, "&amp;", "&", -1)
			fixedURL := strings.ReplaceAll(article.URL, "&amp;", "&")
			if _, exists := uniqueUrls[fixedURL]; !exists {
				uniqueUrls[fixedURL] = struct{}{}
				allArticleURLs = append(allArticleURLs, fixedURL)
				// 文章详细信息
				articleInfo := types.AlbumArticleInfo{
					Index:  len(allArticleInfos) + 1,
					Title:  article.Title,
					URL:    fixedURL,
					Status: "未下载", // 初始状态
				}
				allArticleInfos = append(allArticleInfos, articleInfo)

				zap.L().Info("获取到文章", zap.String("title", article.Title), zap.String("url", fixedURL))
				newArticleFound = true
			}
			// 更新最后一个msgid
			lastMsgID = article.MsgID
		}

		// 检查是否获取到了新文章
		if len(allArticleURLs) == lastArticleCount {
			zap.L().Info("没有获取到新文章，可能进入重复循环，停止获取")
			break
		}
		lastArticleCount = len(allArticleURLs)

		// 更新分页参数
		if len(articleList) > 0 {
			// 设置下一次请求的begin_msgid为当前列表最后一篇文章的msgid
			lastMsgID = articleList[len(articleList)-1].MsgID
		}
		// 更新继续标志
		continueFlag = albumResp.GetAlbumResp.ContinueFlag
		zap.L().Info("更新分页参数", zap.String("begin_msgid", lastMsgID), zap.String("continue_flag", continueFlag))

		// 如果没有找到新文章，可能是因为API返回了重复数据，停止请求
		if !newArticleFound {
			zap.L().Info("没有找到新文章（可能是重复数据），停止请求")
			break
		}

		// 设置请求间隔，避免请求过于频繁触发频率限制
		time.Sleep(requestInterval)
	}

	if requestCount >= maxRequests {
		zap.L().Warn("达到最大请求次数，可能未获取到全部文章")
	}

	zap.L().Info("成功获取所有专辑文章URL", zap.Int("urls_count", len(allArticleURLs)), zap.Int("request_count", requestCount))

	// 4. 导出文章列表到CSV文件
	if len(allArticleInfos) > 0 {
		// 文章总数量
		articleTotal := len(allArticleInfos)
		zap.L().Info("文章总数", zap.Int("total", articleTotal))

		// 生成CSV文件名：NickName_Title.csv
		csvFileName := fmt.Sprintf("%s_%s.csv", indexResp.NickName, sanitizeFilename(indexResp.Title))

		// 创建CSV内容
		var csvContent strings.Builder
		// 写入CSV标题行
		csvContent.WriteString("序号,标题,地址\n")

		// 写入每篇文章的信息
		for _, article := range allArticleInfos {
			// 处理标题中的逗号和引号（CSV格式要求）
			index := articleTotal - article.Index + 1 // 倒序编号
			title := strings.ReplaceAll(article.Title, "\"", "\"\"")
			url := strings.ReplaceAll(article.URL, "\"", "\"\"")
			csvContent.WriteString(fmt.Sprintf("%d,\"%s\",\"%s\"\n", index, title, url))
		}

		// 保存CSV文件到downloads目录
		downloadsDir := utils.GetDefaultDownloadsDir()
		csvFilePath := filepath.Join(downloadsDir, csvFileName)

		// 使用utils.SaveFile函数保存文件
		if err := utils.SaveFile(csvContent.String(), csvFilePath); err != nil {
			fmt.Printf("保存CSV文件失败: %s,%v\n", csvFilePath, err)
			zap.L().Error("保存CSV文件失败", zap.String("filePath", csvFilePath), zap.Error(err))
		} else {
			fmt.Printf("成功保存文章列表到CSV文件: %s\n", csvFilePath)
			zap.L().Info("成功保存文章列表到CSV文件", zap.String("filePath", csvFilePath))
		}
	}

	return allArticleURLs, allArticleInfos, nil
}

// 完全重构的cleanJavaScriptObject函数，更可靠地处理JavaScript对象
func cleanJavaScriptObjectV4(jsObj string) (string, error) {
	// 去除多余的空白字符
	cleaned := strings.TrimSpace(jsObj)

	// 移除JavaScript注释
	cleaned = removeJavaScriptComments(cleaned)

	// 核心改进：使用更强大的JavaScript表达式处理器
	processed := processJavaScriptExpressionsFull(cleaned)

	// 验证生成的字符串是否为有效的JSON
	if !isValidJSON(processed) {
		// 尝试使用终极后备方案：使用状态机逐个字符处理
		fallbackResult := processWithStateMachine(cleaned)
		if isValidJSON(fallbackResult) {
			return fallbackResult, nil
		}
		return "", fmt.Errorf("转换后的字符串不是有效的JSON: %s", processed)
	}

	return processed, nil
}

// 更强大的JavaScript表达式处理器，能处理各种复杂情况
func processJavaScriptExpressionsFull(jsObj string) string {
	result := jsObj

	// 1. 处理URL字符串中的多余引号问题
	// 修复形如: ""http://..." 的情况
	doubleQuoteURLRegex := regexp.MustCompile(`""(https?://[^"\s]*)"`)
	result = doubleQuoteURLRegex.ReplaceAllString(result, `"$1"`)

	// 修复形如: "http://..."" 的情况
	reverseDoubleQuoteURLRegex := regexp.MustCompile(`"(https?://[^"\s]*)""`)
	result = reverseDoubleQuoteURLRegex.ReplaceAllString(result, `"$1"`)

	// 重点改进：修复形如: ""http"://..." 或 ""https"://..." 的特殊情况
	// 使用更精确的正则表达式，确保能匹配所有变体
	specialDoubleQuoteURLRegex := regexp.MustCompile(`""http"://([^"\s]*)"`)
	result = specialDoubleQuoteURLRegex.ReplaceAllString(result, `"http://$1"`)
	specialDoubleQuoteHttpsRegex := regexp.MustCompile(`""https"://([^"\s]*)"`)
	result = specialDoubleQuoteHttpsRegex.ReplaceAllString(result, `"https://$1"`)

	// 2. 使用重复处理直到没有更多变化（确保处理嵌套表达式）
	prevResult := ""
	for i := 0; i < 10; i++ { // 限制循环次数，防止无限循环
		prevResult = result

		// 重点改进：增强处理形如 "nick_name": "一安未来" || "" 的表达式
		// 使用更宽松的正则表达式，确保能匹配所有包含中文字符的字符串
		enhancedOrExprRegex := regexp.MustCompile(`"([^"\\]*(?:\\.[^"\\]*)*)"\s*\|\|\s*"[^"]*"`)
		result = enhancedOrExprRegex.ReplaceAllString(result, `"$1"`)

		// 在循环内部再次应用URL修复，确保处理嵌套或复杂情况
		result = specialDoubleQuoteURLRegex.ReplaceAllString(result, `"http://$1"`)
		result = specialDoubleQuoteHttpsRegex.ReplaceAllString(result, `"https://$1"`)

		// 3. 处理字符串和数字的乘法转换: "123" * 1 -> 123
		numExprRegex := regexp.MustCompile(`"(\d+)"\s*\*\s*1`)
		result = numExprRegex.ReplaceAllString(result, `$1`)

		// 4. 处理空字符串乘法: "" * 1 -> 0
		emptyNumExprRegex := regexp.MustCompile(`""\s*\*\s*1`)
		result = emptyNumExprRegex.ReplaceAllString(result, `0`)

		// 5. 处理逻辑或表达式：非空字符串 || 默认值 -> 保留非空字符串
		// 匹配形如: "value" || "default"
		orExprRegex := regexp.MustCompile(`"([^"]+)"\s*\|\|\s*"[^"]+"`)
		result = orExprRegex.ReplaceAllString(result, `"$1"`)

		// 6. 处理逻辑或表达式：空字符串 || 非空字符串 -> 保留非空字符串
		emptyOrExprRegex := regexp.MustCompile(`""\s*\|\|\s*"([^"]+)"`)
		result = emptyOrExprRegex.ReplaceAllString(result, `"$1"`)

		// 7. 处理逻辑或表达式：空字符串 || 数字 -> 保留数字
		emptyOrNumExprRegex := regexp.MustCompile(`""\s*\|\|\s*(-?\d+)`)
		result = emptyOrNumExprRegex.ReplaceAllString(result, `$1`)

		// 8. 处理逻辑或表达式：字符串 || 数字 -> 保留字符串
		strOrNumExprRegex := regexp.MustCompile(`"([^"]+)"\s*\|\|\s*(-?\d+)`)
		result = strOrNumExprRegex.ReplaceAllString(result, `"$1"`)

		// 9. 处理三元运算符：空字符串条件 -> 使用默认值
		// 匹配形如: "" ? value1 : value2
		ternaryRegex1 := regexp.MustCompile(`""\s*\?\s*"[^"]*"\s*:\s*("[^"]*"|-?\d+)`)
		result = ternaryRegex1.ReplaceAllString(result, `$1`)
		// 匹配形如: "" ? 0 : -1
		ternaryRegex2 := regexp.MustCompile(`""\s*\?\s*(-?\d+)\s*:\s*(-?\d+)`)
		result = ternaryRegex2.ReplaceAllString(result, `$2`)

		// 10. 处理三元运算符：非空字符串条件 -> 使用第一个值
		ternaryNonEmptyRegex1 := regexp.MustCompile(`"([^"]+)"\s*\?\s*("[^"]*"|-?\d+)\s*:\s*("[^"]*"|-?\d+)`)
		result = ternaryNonEmptyRegex1.ReplaceAllString(result, `$2`)

		// 11. 修复多余的逗号
		result = regexp.MustCompile(`,\s*}`).ReplaceAllString(result, `}`)
		result = regexp.MustCompile(`,\s*]`).ReplaceAllString(result, `]`)

		// 12. 修复连续的逗号
		result = regexp.MustCompile(`\,\s*\,`).ReplaceAllString(result, `,`)

		// 13. 修复单引号为双引号
		result = strings.ReplaceAll(result, `'`, `"`)

		// 14. 修复键名缺失引号的情况
		result = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*:`).ReplaceAllString(result, `"$1":`)

		// 如果没有变化，退出循环
		if prevResult == result {
			break
		}
	}

	// 最终清理：再次应用所有URL修复，确保彻底解决问题
	result = specialDoubleQuoteURLRegex.ReplaceAllString(result, `"http://$1"`)
	result = specialDoubleQuoteHttpsRegex.ReplaceAllString(result, `"https://$1"`)
	result = doubleQuoteURLRegex.ReplaceAllString(result, `"$1"`)
	result = reverseDoubleQuoteURLRegex.ReplaceAllString(result, `"$1"`)

	return result
}

func processJavaScriptExpressionsFull_Backup(jsObj string) string {
	result := jsObj

	// 重要：先处理字符串中的特殊情况，特别是URL字符串中的多余引号
	result = fixURLStrings(result)

	// 使用重复处理直到没有更多变化（确保处理嵌套表达式）
	prevResult := ""
	for prevResult != result && len(result) > 0 {
		prevResult = result

		// 1. 处理字符串和数字的乘法转换: "123" * 1 -> 123
		numExprRegex := regexp.MustCompile(`"(\d+)"\s*\*\s*1`)
		result = numExprRegex.ReplaceAllString(result, `$1`)

		// 2. 处理空字符串乘法: "" * 1 -> 0
		emptyNumExprRegex := regexp.MustCompile(`""\s*\*\s*1`)
		result = emptyNumExprRegex.ReplaceAllString(result, `0`)

		// 3. 处理逻辑或表达式：非空字符串 || 默认值 -> 保留非空字符串
		// 匹配形如: "value" || "default"
		orExprRegex := regexp.MustCompile(`"([^"]+)"\s*\|\|\s*"[^"]+"`)
		result = orExprRegex.ReplaceAllString(result, `"$1"`)

		// 4. 处理逻辑或表达式：空字符串 || 非空字符串 -> 保留非空字符串
		emptyOrExprRegex := regexp.MustCompile(`""\s*\|\|\s*"([^"]+)"`)
		result = emptyOrExprRegex.ReplaceAllString(result, `"$1"`)

		// 5. 处理逻辑或表达式：空字符串 || 数字 -> 保留数字
		emptyOrNumExprRegex := regexp.MustCompile(`""\s*\|\|\s*(-?\d+)`)
		result = emptyOrNumExprRegex.ReplaceAllString(result, `$1`)

		// 6. 处理逻辑或表达式：字符串 || 数字 -> 保留字符串
		strOrNumExprRegex := regexp.MustCompile(`"([^"]+)"\s*\|\|\s*(-?\d+)`)
		result = strOrNumExprRegex.ReplaceAllString(result, `"$1"`)

		// 7. 处理三元运算符：空字符串条件 -> 使用默认值
		ternaryRegex := regexp.MustCompile(`""\s*\?\s*"[^"]*"\s*:\s*("[^"]*"|-?\d+)`)
		result = ternaryRegex.ReplaceAllString(result, `$1`)

		// 8. 处理三元运算符：非空字符串条件 -> 使用第一个值
		ternaryNonEmptyRegex := regexp.MustCompile(`"([^"]+)"\s*\?\s*("[^"]*"|-?\d+)\s*:\s*("[^"]*"|-?\d+)`)
		result = ternaryNonEmptyRegex.ReplaceAllString(result, `$2`)

		// 修复可能引入的格式问题
		result = fixJSONFormat(result)
	}

	return result
}

// 修复URL字符串中的多余引号问题
func fixURLStrings(jsObj string) string {
	result := jsObj

	// 修复形如: ""http://..." 的情况
	doubleQuoteURLRegex := regexp.MustCompile(`""(https?://[^"\s]+)"`)
	result = doubleQuoteURLRegex.ReplaceAllString(result, `"$1"`)

	// 修复形如: "http://..."" 的情况
	reverseDoubleQuoteURLRegex := regexp.MustCompile(`"(https?://[^"\s]+)""`)
	result = reverseDoubleQuoteURLRegex.ReplaceAllString(result, `"$1"`)

	return result
}

// 使用状态机处理，作为最后的后备方案
func processWithStateMachine(jsObj string) string {
	var result strings.Builder
	var inString bool
	var inKey bool
	var expectingValue bool
	var escapeNext bool
	var braceCount int
	var bracketCount int

	for i := 0; i < len(jsObj); i++ {
		currentChar := jsObj[i]

		// 处理转义字符
		if escapeNext {
			result.WriteByte(currentChar)
			escapeNext = false
			continue
		}

		// 处理转义符
		if currentChar == '\\' && inString {
			escapeNext = true
			result.WriteByte(currentChar)
			continue
		}

		// 处理字符串边界
		if (currentChar == '"' || currentChar == '\'') && !escapeNext {
			if !inString {
				// 开始字符串
				inString = true
				result.WriteByte('"') // 统一使用双引号
			} else if (currentChar == '"' && jsObj[i-1] != '\\') ||
				(currentChar == '\'' && jsObj[i-1] != '\\') {
				// 结束字符串
				inString = false
				result.WriteByte('"') // 统一使用双引号
			} else {
				// 字符串内容中的引号，转义它
				result.WriteByte('\\')
				result.WriteByte(currentChar)
			}
			continue
		}

		// 不在字符串中时的特殊处理
		if !inString {
			// 计算括号嵌套级别
			if currentChar == '{' {
				braceCount++
				inKey = true
				expectingValue = false
				result.WriteByte(currentChar)
			} else if currentChar == '}' {
				braceCount--
				result.WriteByte(currentChar)
			} else if currentChar == '[' {
				bracketCount++
				result.WriteByte(currentChar)
			} else if currentChar == ']' {
				bracketCount--
				result.WriteByte(currentChar)
			} else if currentChar == ':' && braceCount > 0 {
				expectingValue = true
				inKey = false
				result.WriteByte(currentChar)
			} else if currentChar == ',' && braceCount > 0 {
				expectingValue = false
				inKey = true
				result.WriteByte(currentChar)
			} else if isIdentifierChar(currentChar) && braceCount > 0 {
				// 处理键名（如果没有引号）
				if inKey && !expectingValue {
					// 检查是否是键名（直到遇到冒号）
					j := i
					for j < len(jsObj) && isIdentifierChar(jsObj[j]) {
						j++
					}
					if j < len(jsObj) && jsObj[j] == ':' {
						// 是键名，需要用双引号包裹
						result.WriteByte('"')
						result.WriteString(jsObj[i:j])
						result.WriteByte('"')
						i = j - 1 // 跳过已经处理的字符
					}
				} else {
					result.WriteByte(currentChar)
				}
			} else if currentChar == ' ' || currentChar == '\t' || currentChar == '\n' || currentChar == '\r' {
				// 保留必要的空白字符
				result.WriteByte(currentChar)
			} else if (currentChar == '|' || currentChar == '*' || currentChar == '?' || currentChar == ':') && braceCount > 0 {
				// 简单处理JavaScript表达式：直接跳过
				// 注意：这是一个简化的处理方式，仅作为最后的后备方案
				j := i
				for j < len(jsObj) && jsObj[j] != ',' && jsObj[j] != '}' && jsObj[j] != ']' {
					j++
				}
				// 根据上下文提供一个默认值
				if expectingValue {
					if currentChar == '*' && i > 0 && jsObj[i-1] == '"' && j < len(jsObj) && jsObj[j-1] == '1' {
						// 可能是 "数字" * 1 的模式，设置为0
						result.WriteString("0")
					} else if currentChar == '|' {
						// 可能是 || 默认值的模式，设置为空字符串
						result.WriteString(`""`)
					} else {
						result.WriteString(`""`)
					}
				}
				i = j - 1 // 跳过已经处理的字符
			} else {
				// 其他字符直接写入
				result.WriteByte(currentChar)
			}
		} else {
			// 字符串内容直接写入
			result.WriteByte(currentChar)
		}
	}

	// 最终修复JSON格式
	finalResult := result.String()
	finalResult = fixJSONFormat(finalResult)

	return finalResult
}

// 完全重构的cleanJavaScriptObject函数，能够处理复杂的JavaScript表达式
func cleanJavaScriptObjectV3(jsObj string) (string, error) {
	// 去除多余的空白字符
	cleaned := strings.TrimSpace(jsObj)

	// 移除JavaScript注释
	cleaned = removeJavaScriptComments(cleaned)

	// 核心改进：使用增强版JavaScript表达式求值器预处理
	processed := evaluateJavaScriptExpressionsEnhanced(cleaned)

	// 尝试修复未加引号的键名
	processed = fixUnquotedKeys(processed)

	// 修复JSON格式问题
	processed = fixJSONFormat(processed)

	// 验证生成的字符串是否为有效的JSON
	if !isValidJSON(processed) {
		// 尝试使用宽松模式再次处理
		looseProcessed := tryLooseProcessing(jsObj)
		if isValidJSON(looseProcessed) {
			return looseProcessed, nil
		}
		return "", fmt.Errorf("转换后的字符串不是有效的JSON: %s", processed)
	}

	return processed, nil
}

// 增强版JavaScript表达式求值器
func evaluateJavaScriptExpressionsEnhanced(jsObj string) string {
	result := jsObj

	// 1. 处理字符串和数字的乘法转换: "123" * 1 -> 123
	numExprRegex := regexp.MustCompile(`("[^"]+")\s*:\s*"(\d+)"\s*\*\s*1`)
	result = numExprRegex.ReplaceAllString(result, `$1: $2`)

	// 2. 处理空字符串乘法: "" * 1 -> 0
	emptyNumExprRegex := regexp.MustCompile(`("[^"]+")\s*:\s*""\s*\*\s*1`)
	result = emptyNumExprRegex.ReplaceAllString(result, `$1: 0`)

	// 3. 处理逻辑或表达式：非空字符串 || 默认值 -> 保留非空字符串
	// 匹配形如: "key": "value" || "default"
	orExprRegex := regexp.MustCompile(`("[^"]+")\s*:\s*"([^"]+)"\s*\|\|\s*"[^"]+"`)
	result = orExprRegex.ReplaceAllString(result, `$1: "$2"`)

	// 4. 处理逻辑或表达式：空字符串 || 非空字符串 -> 保留非空字符串
	emptyOrExprRegex := regexp.MustCompile(`("[^"]+")\s*:\s*""\s*\|\|\s*"([^"]+)"`)
	result = emptyOrExprRegex.ReplaceAllString(result, `$1: "$2"`)

	// 5. 处理逻辑或表达式：空字符串 || 数字 -> 保留数字
	emptyOrNumExprRegex := regexp.MustCompile(`("[^"]+")\s*:\s*""\s*\|\|\s*(-?\d+)`)
	result = emptyOrNumExprRegex.ReplaceAllString(result, `$1: $2`)

	// 6. 处理逻辑或表达式：字符串 || 数字 -> 保留字符串
	strOrNumExprRegex := regexp.MustCompile(`("[^"]+")\s*:\s*"([^"]+)"\s*\|\|\s*(-?\d+)`)
	result = strOrNumExprRegex.ReplaceAllString(result, `$1: "$2"`)

	// 7. 处理三元运算符：空字符串条件 -> 使用默认值
	ternaryRegex := regexp.MustCompile(`("[^"]+")\s*:\s*""\s*\?\s*"[^"]*"\s*:\s*("[^"]*"|-?\d+)`)
	result = ternaryRegex.ReplaceAllString(result, `$1: $2`)

	// 8. 处理三元运算符：非空字符串条件 -> 使用第一个值
	ternaryNonEmptyRegex := regexp.MustCompile(`("[^"]+")\s*:\s*"([^"]+)"\s*\?\s*("[^"]*"|-?\d+)\s*:\s*("[^"]*"|-?\d+)`)
	result = ternaryNonEmptyRegex.ReplaceAllString(result, `$1: $3`)

	// 9. 处理嵌套对象中的表达式
	nestedObjRegex := regexp.MustCompile(`("[^"]+")\s*:\s*\{\s*("[^"]+")\s*:\s*"([^"]+)"\s*\*\s*1`)
	result = nestedObjRegex.ReplaceAllString(result, `$1: { $2: $3`)

	return result
}

// 修复未加引号的键名
func fixUnquotedKeys(jsObj string) string {
	// 简单处理：将未加引号的键名加上双引号
	// 这个正则表达式会匹配标识符后跟冒号的模式
	keyRegex := regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*:`)
	return keyRegex.ReplaceAllString(jsObj, `"$1":`)
}

// 修复JSON格式问题
func fixJSONFormat(jsObj string) string {
	result := jsObj

	// 1. 修复多余的逗号
	result = regexp.MustCompile(`,\s*}`).ReplaceAllString(result, `}`)
	result = regexp.MustCompile(`,\s*]`).ReplaceAllString(result, `]`)

	// 2. 修复连续的逗号
	result = regexp.MustCompile(`\,\s*\,`).ReplaceAllString(result, `,`)

	// 3. 修复单引号为双引号
	result = strings.ReplaceAll(result, `'`, `"`)

	// 4. 修复字符串中的转义字符（如果有必要）
	// 注意：这是一个简化的实现，可能需要根据具体情况调整
	result = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*:`).ReplaceAllString(result, `"$1":`)

	return result
}

// 宽松模式处理作为最后的后备方案
func tryLooseProcessing(jsObj string) string {
	// 1. 移除所有JavaScript表达式（非常粗暴的方法，但作为最后的手段）
	result := jsObj

	// 处理各种表达式模式
	patterns := []string{
		`\*\s*1`,              // 移除 "数字" * 1
		`\|\|\s*"[^"]+"`,      // 移除 || "default"
		`\|\|\s*(-?\d+)`,      // 移除 || 数字
		`\?\s*"[^"]*"\s*:\s*`, // 移除 ? "value" :
	}

	for _, pattern := range patterns {
		result = regexp.MustCompile(pattern).ReplaceAllString(result, ``)
	}

	// 2. 修复未加引号的键名
	result = fixUnquotedKeys(result)

	// 3. 修复JSON格式
	result = fixJSONFormat(result)

	return result
}

// 完全重构的cleanJavaScriptObject函数，能够评估JavaScript表达式
func cleanJavaScriptObjectV2(jsObj string) (string, error) {
	// 去除多余的空白字符
	cleaned := strings.TrimSpace(jsObj)

	// 移除JavaScript注释
	cleaned = removeJavaScriptComments(cleaned)

	// 核心改进：使用JavaScript表达式求值器预处理
	processed := evaluateJavaScriptExpressions(cleaned)

	// 验证生成的字符串是否为有效的JSON
	if !isValidJSON(processed) {
		return "", fmt.Errorf("转换后的字符串不是有效的JSON: %s", processed)
	}

	return processed, nil
}

// 评估并替换JavaScript表达式
func evaluateJavaScriptExpressions(jsObj string) string {
	result := jsObj

	// 1. 处理逻辑或表达式："value" || "default" -> "value" (如果value非空)
	// 匹配形如: "key": "value" || "default"
	orExprRegex := regexp.MustCompile(`("[^\"]+")\s*:\s*("[^"]*")\s*\|\|\s*("[^"]*")`)
	for {
		match := orExprRegex.FindStringSubmatch(result)
		if len(match) == 0 {
			break
		}

		key := match[1]
		value1 := match[2]
		value2 := match[3]

		// 如果第一个值非空字符串，使用它；否则使用第二个值
		if value1 != `""` {
			result = strings.Replace(result, match[0], fmt.Sprintf("%s: %s", key, value1), 1)
		} else {
			result = strings.Replace(result, match[0], fmt.Sprintf("%s: %s", key, value2), 1)
		}
	}

	// 2. 处理字符串和数字的逻辑或："value" || 123 -> "value" 或 123
	mixedOrExprRegex := regexp.MustCompile(`("[^\"]+")\s*:\s*("[^"]*")\s*\|\|\s*(-?\d+)`)
	for {
		match := mixedOrExprRegex.FindStringSubmatch(result)
		if len(match) == 0 {
			break
		}

		key := match[1]
		value1 := match[2]
		value2 := match[3]

		if value1 != `""` {
			result = strings.Replace(result, match[0], fmt.Sprintf("%s: %s", key, value1), 1)
		} else {
			result = strings.Replace(result, match[0], fmt.Sprintf("%s: %s", key, value2), 1)
		}
	}

	// 3. 处理数字转换表达式："123" * 1 -> 123
	numExprRegex := regexp.MustCompile(`("[^\"]+")\s*:\s*"(\d+)"\s*\*\s*1`)
	result = numExprRegex.ReplaceAllString(result, `$1: $2`)

	// 4. 处理空字符串乘法："" * 1 -> 0
	emptyNumExprRegex := regexp.MustCompile(`("[^\"]+")\s*:\s*""\s*\*\s*1`)
	result = emptyNumExprRegex.ReplaceAllString(result, `$1: 0`)

	// 5. 处理三元运算符："" ? "value1" : "value2" -> "value2"
	ternaryRegex := regexp.MustCompile(`("[^\"]+")\s*:\s*""\s*\?\s*"[^"]*"\s*:\s*("[^"]*"|-?\d+)`)
	result = ternaryRegex.ReplaceAllString(result, `$1: $2`)

	// 6. 处理空对象中的多余逗号
	result = regexp.MustCompile(`,\s*}`).ReplaceAllString(result, "}")
	result = regexp.MustCompile(`,\s*]`).ReplaceAllString(result, "]")

	// 7. 处理连续的逗号
	result = regexp.MustCompile(`\,\s*\,`).ReplaceAllString(result, ",")

	return result
}

// 完全重构的cleanJavaScriptObject函数，使用更强大的方法处理JavaScript表达式
func cleanJavaScriptObjectV1(jsObj string) (string, error) {
	// 去除多余的空白字符
	cleaned := strings.TrimSpace(jsObj)

	// 移除JavaScript注释
	cleaned = removeJavaScriptComments(cleaned)

	// 使用自定义解析器来安全地转换JavaScript对象为JSON字符串
	//result, err := parseJavaScriptObject(cleaned)
	// 使用增强版解析器来处理JavaScript对象
	result, err := enhancedParseJavaScriptObject(cleaned)
	if err != nil {
		return "", fmt.Errorf("解析JavaScript对象失败: %v\n原始对象: %s", err, jsObj)
	}

	// 验证生成的字符串是否为有效的JSON
	if !isValidJSON(result) {
		// 尝试使用更宽松的解析方式作为后备方案
		fallbackResult := tryLooseJSONParsing(cleaned)
		if isValidJSON(fallbackResult) {
			return fallbackResult, nil
		}
		return "", fmt.Errorf("转换后的字符串不是有效的JSON: %s\n尝试宽松解析后的结果: %s", result, fallbackResult)
	}

	return result, nil
}

// 增强版JavaScript对象解析器
func enhancedParseJavaScriptObject(jsObj string) (string, error) {
	// 预处理：处理JavaScript表达式
	processed := processJavaScriptExpressionsEnhanced(jsObj)

	// 现在使用标准的JSON解析器来验证
	var result strings.Builder
	var inString bool
	var escapeNext bool
	var braceCount int
	var bracketCount int
	var lastChar byte

	// 跟踪键和值的状态
	var inKey bool
	var expectingValue bool

	for i := 0; i < len(processed); i++ {
		currentChar := processed[i]

		// 处理转义字符
		if escapeNext {
			result.WriteByte(currentChar)
			escapeNext = false
			continue
		}

		// 处理转义符
		if currentChar == '\\' && inString {
			escapeNext = true
			result.WriteByte(currentChar)
			continue
		}

		// 处理字符串边界
		if (currentChar == '"' || currentChar == '\'') && !escapeNext {
			if !inString {
				// 开始字符串
				inString = true
				result.WriteByte('"') // 统一使用双引号
			} else if (currentChar == '"' && lastChar != '\\') ||
				(currentChar == '\'' && lastChar != '\\') {
				// 结束字符串
				inString = false
				result.WriteByte('"') // 统一使用双引号
			} else {
				// 字符串内容中的引号，转义它
				result.WriteByte('\\')
				result.WriteByte(currentChar)
			}
			lastChar = currentChar
			continue
		}

		// 不在字符串中时的特殊处理
		if !inString {
			// 计算括号嵌套级别
			if currentChar == '{' {
				braceCount++
				inKey = true
				expectingValue = false
			} else if currentChar == '}' {
				braceCount--
			} else if currentChar == '[' {
				bracketCount++
			} else if currentChar == ']' {
				bracketCount--
			} else if currentChar == ':' && braceCount > 0 {
				expectingValue = true
				inKey = false
			} else if currentChar == ',' && braceCount > 0 {
				expectingValue = false
				inKey = true
			} else if !isWhitespace(currentChar) && braceCount > 0 {
				// 处理键名（如果没有引号）
				if inKey && !expectingValue && !isSpecialChar(currentChar) {
					// 检查是否是键名（直到遇到冒号）
					j := i
					for j < len(processed) && !isSpecialChar(processed[j]) && !isWhitespace(processed[j]) {
						j++
					}
					if j < len(processed) && processed[j] == ':' {
						// 是键名，需要用双引号包裹
						result.WriteByte('"')
						result.WriteString(processed[i:j])
						result.WriteByte('"')
						i = j - 1 // 跳过已经处理的字符
						continue
					}
				}
			}
		}

		// 其他字符直接写入
		result.WriteByte(currentChar)
		lastChar = currentChar
	}

	return result.String(), nil
}

// 增强版JavaScript表达式处理器
func processJavaScriptExpressionsEnhanced(jsObj string) string {
	str := jsObj

	// 1. 处理数字转换："123" * 1 -> 123
	multRegex := regexp.MustCompile(`"(\d+)"\s*\*\s*1`)
	str = multRegex.ReplaceAllString(str, "$1")

	// 2. 处理空字符串乘法："" * 1 -> 0
	emptyMultRegex := regexp.MustCompile(`""\s*\*\s*1`)
	str = emptyMultRegex.ReplaceAllString(str, "0")

	// 3. 处理逻辑或默认值："string" || "default" -> "stringdefault"
	orStringRegex := regexp.MustCompile(`"([^"]*)"\s*\|\|\s*"([^"]*)"`)
	str = orStringRegex.ReplaceAllString(str, `"$1$2"`)

	// 4. 处理字符串或数字默认值："string" || 123 -> "string123"
	orMixedRegex := regexp.MustCompile(`"([^"]*)"\s*\|\|\s*(-?\d+)`)
	str = orMixedRegex.ReplaceAllString(str, `"$1$2"`)

	// 5. 处理空字符串或数字默认值："" || 123 -> 123
	emptyOrNumberRegex := regexp.MustCompile(`""\s*\|\|\s*(-?\d+)`)
	str = emptyOrNumberRegex.ReplaceAllString(str, "$1")

	// 6. 处理三元运算符：condition ? value1 : value2
	// 简单处理：空字符串条件 -> 使用默认值
	ternaryEmptyRegex := regexp.MustCompile(`""\s*\?\s*"[^"]*"\s*:\s*(-?\d+)`)
	str = ternaryEmptyRegex.ReplaceAllString(str, "$1")

	// 7. 处理多余的逗号
	str = regexp.MustCompile(`,\s*}`).ReplaceAllString(str, "}")
	str = regexp.MustCompile(`,\s*]`).ReplaceAllString(str, "]")

	// 8. 处理多个连续的逗号
	str = regexp.MustCompile(`\,\s*\,`).ReplaceAllString(str, ",")

	return str
}

// 宽松的JSON解析作为后备方案
func tryLooseJSONParsing(jsObj string) string {
	// 这是一个更宽松的解析方法，尝试尽可能修复JSON格式问题
	str := jsObj

	// 处理未加引号的键
	str = regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*:`).ReplaceAllString(str, `"$1":`)

	// 处理字符串中的单引号
	str = strings.ReplaceAll(str, `'`, `"`)

	// 处理多余的逗号
	str = regexp.MustCompile(`,\s*}`).ReplaceAllString(str, "}")
	str = regexp.MustCompile(`,\s*]`).ReplaceAllString(str, "]")

	return str
}

// 自定义解析器，安全地将JavaScript对象转换为JSON字符串，支持常见的JavaScript表达式
func parseJavaScriptObject(jsObj string) (string, error) {
	// 预处理：处理JavaScript表达式
	processed := processJavaScriptExpressions(jsObj)

	// 使用缓冲区来构建结果
	var result strings.Builder
	// 当前解析状态
	var inString bool
	var inNumber bool
	var inComment bool
	var escapeNext bool
	var quoteChar byte

	for i := 0; i < len(processed); i++ {
		currentChar := processed[i]

		// 处理转义字符
		if escapeNext {
			result.WriteByte(currentChar)
			escapeNext = false
			continue
		}

		// 处理转义符
		if currentChar == '\\' {
			escapeNext = true
			result.WriteByte(currentChar)
			continue
		}

		// 处理字符串
		if (currentChar == '"' || currentChar == '\'') && !inComment {
			if !inString {
				// 开始新字符串
				inString = true
				quoteChar = currentChar
				// 总是使用双引号
				result.WriteByte('"')
			} else if currentChar == quoteChar {
				// 结束字符串
				inString = false
				// 总是使用双引号
				result.WriteByte('"')
			} else {
				// 字符串内容
				result.WriteByte(currentChar)
			}
			continue
		}

		// 处理对象键名（未加引号的）
		if !inString && !inNumber && !inComment && isIdentifierChar(currentChar) {
			// 检查是否是键名（后面跟着冒号）
			isKey := false
			j := i
			for j < len(processed) && isIdentifierChar(processed[j]) {
				j++
			}
			if j < len(processed) && processed[j] == ':' {
				isKey = true
			}

			if isKey {
				// 是键名，用双引号包裹
				result.WriteByte('"')
				result.WriteString(processed[i:j])
				result.WriteByte('"')
				i = j - 1 // 跳过已经处理的字符
			} else {
				// 不是键名，正常写入
				result.WriteByte(currentChar)
			}
			continue
		}

		// 处理数字、布尔值、null等
		if !inString && !inComment {
			// 处理数字
			if isDigit(currentChar) && !inNumber {
				inNumber = true
			} else if !isDigit(currentChar) && inNumber {
				inNumber = false
			}

			// 处理布尔值和null（确保小写）
			if i < len(processed)-3 && strings.ToLower(processed[i:i+4]) == "true" {
				result.WriteString("true")
				i += 3 // 跳过剩余字符
				continue
			} else if i < len(processed)-4 && strings.ToLower(processed[i:i+5]) == "false" {
				result.WriteString("false")
				i += 4 // 跳过剩余字符
				continue
			} else if i < len(processed)-3 && strings.ToLower(processed[i:i+4]) == "null" {
				result.WriteString("null")
				i += 3 // 跳过剩余字符
				continue
			}
		}

		// 其他字符直接写入
		result.WriteByte(currentChar)
	}

	return result.String(), nil
}

// 处理常见的JavaScript表达式
/*
预处理JavaScript表达式：
处理字符串拼接 ("string" + "another")
处理数字转换 ("123" * 1 转换为 123)
处理逻辑或默认值 (value || defaultValue)
处理简单的三元运算符

JavaScript中的各种表达式，尤其是"string" || defaultValue、"number" * 1和三元运算符等复杂情况。
需处理JavaScript中的各种表达式（乘法、逻辑或、三元运算符等）如下：
"is_pay_subscribe": "0" * 1,
"isupdating": "1" * 1,
"nick_name": "一安未来" || "",
"user_name": "gh_0b542121801e" || "",
"total_onread": "" ? "" * 1 : -1,
*/
func processJavaScriptExpressions(jsObj string) string {
	// 处理字符串拼接："string" + "another string" -> "stringanother string"
	// 这个简单实现可能不足以处理所有情况，但对于常见的情况应该有效
	str := jsObj

	// 处理乘法转换："123" * 1 -> 123
	multRegex := regexp.MustCompile(`"(\d+)"\s*\*\s*1`)
	str = multRegex.ReplaceAllString(str, "$1")

	// 处理逻辑或默认值：value || defaultValue
	// 对于字符串或空值的情况
	orStrRegex := regexp.MustCompile(`"(.*?)"\s*\|\|\s*"(.*?)"`)
	str = orStrRegex.ReplaceAllString(str, "$1$2")

	// 处理空字符串或默认值："" || defaultValue
	emptyOrRegex := regexp.MustCompile(`""\s*\|\|\s*(-?\d+)"`)
	str = emptyOrRegex.ReplaceAllString(str, "$1")

	// 处理三元运算符的简单情况：condition ? value1 : value2
	ternaryRegex := regexp.MustCompile(`""\s*\?\s*".*?"\s*:\s*(-?\d+)`)
	str = ternaryRegex.ReplaceAllString(str, "$1")

	// 处理空对象中的多余逗号
	str = regexp.MustCompile(`,\s*}`).ReplaceAllString(str, "}")
	str = regexp.MustCompile(`,\s*]`).ReplaceAllString(str, "]")

	return str
}

// 判断字符是否为空白字符
func isWhitespace(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}

// 判断字符是否为特殊字符
func isSpecialChar(c byte) bool {
	return c == '{' || c == '}' || c == '[' || c == ']' || c == ':' || c == ',' || c == '"'
}

// 判断字符是否是标识符字符（字母、数字、下划线）
func isIdentifierChar(c byte) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_'
}

// 判断字符是否是数字
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// 移除JavaScript注释
func removeJavaScriptComments(jsCode string) string {
	// 移除单行注释 //
	singleLineCommentRegex := regexp.MustCompile(`//.*$`)
	cleaned := singleLineCommentRegex.ReplaceAllString(jsCode, "")

	// 移除多行注释 /* */
	multiLineCommentRegex := regexp.MustCompile(`/\*[\s\S]*?\*/`)
	cleaned = multiLineCommentRegex.ReplaceAllString(cleaned, "")

	return cleaned
}

// 验证字符串是否为有效的JSON
func isValidJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// parseAlbumHomeURL 解析专辑首页URL，提取必要参数
func parseAlbumHomeURL(homeURL string) (map[string]string, error) {
	// 解析URL
	parsedURL, err := url.Parse(homeURL)
	if err != nil {
		return nil, fmt.Errorf("解析URL失败: %v", err)
	}

	// 提取查询参数
	queryParams := parsedURL.Query()
	params := make(map[string]string)

	// 提取必要参数
	if val := queryParams.Get("__biz"); val != "" {
		params["__biz"] = val
	}
	if val := queryParams.Get("album_id"); val != "" {
		params["album_id"] = val
	}

	// 检查必要参数是否存在
	if params["__biz"] == "" || params["album_id"] == "" {
		return nil, fmt.Errorf("URL中缺少必要参数: __biz 或 album_id")
	}

	return params, nil
}

// 解析专辑首页URL，提取__biz和album_id参数
func parseAlbumHomeURLRegex(url string) (string, string, error) {
	// 提取__biz参数
	bizRegex := regexp.MustCompile(`__biz=([^&]+)`)
	bizMatches := bizRegex.FindStringSubmatch(url)
	if len(bizMatches) < 2 {
		return "", "", errors.New("未找到__biz参数")
	}
	biz := bizMatches[1]

	// 提取album_id参数
	albumIDRegex := regexp.MustCompile(`album_id=([^&]+)`)
	albumIDMatches := albumIDRegex.FindStringSubmatch(url)
	if len(albumIDMatches) < 2 {
		return "", "", errors.New("未找到album_id参数")
	}
	albumID := albumIDMatches[1]

	return biz, albumID, nil
}

// buildAlbumAPIURL 构建专辑API请求URL
func buildAlbumAPIURL(params map[string]string, beginMsgID string) string {
	// 基础API URL
	baseURL := "https://mp.weixin.qq.com/mp/appmsgalbum"
	count := 10 // 每次请求获取10篇文章

	// 构建查询参数
	queryParams := url.Values{}
	queryParams.Set("action", "getalbum")
	queryParams.Set("__biz", params["__biz"])
	queryParams.Set("album_id", params["album_id"])
	queryParams.Set("count", fmt.Sprintf("%d", count))
	queryParams.Set("f", "json")

	// 添加begin_msgid参数（如果有）
	if beginMsgID != "" {
		queryParams.Set("begin_msgid", beginMsgID)
		queryParams.Set("begin_itemidx", "1")
	}

	// 构建完整URL
	return fmt.Sprintf("%s?%s", baseURL, queryParams.Encode())
}

// 构建专辑API请求URL
func buildAlbumAPIURLParam(biz, albumID string, count int, beginMsgID string) string {
	// 基本URL模板
	baseURL := "https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&__biz=%s&album_id=%s&count=%d&begin_itemidx=1&uin=&key=&pass_ticket=&wxtoken=&devicetype=&clientversion=&__biz=%s&appmsg_token=&x5=0&f=json"

	// 如果有begin_msgid，则添加到URL中
	if beginMsgID != "" {
		baseURL = "https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&__biz=%s&album_id=%s&count=%d&begin_msgid=%s&begin_itemidx=1&uin=&key=&pass_ticket=&wxtoken=&devicetype=&clientversion=&__biz=%s&appmsg_token=&x5=0&f=json"
		return fmt.Sprintf(baseURL, biz, albumID, count, beginMsgID, biz)
	}

	return fmt.Sprintf(baseURL, biz, albumID, count, biz)
}
