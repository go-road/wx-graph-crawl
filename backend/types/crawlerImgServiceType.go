package types

type CrawlResult struct {
	URL                string   // 需要被抓取的原始链接地址
	Number             int      // 当前子协程的编号
	Err                error    // 抓取过程中出现的错误
	Title              string   // 文章标题
	Html               string   // 链接地址对应的抓取内容
	ImgSavePathSuccess []string // 图片存储的硬盘路径地址
	WriteContent       string   // 需要被写入的文字内容
}
