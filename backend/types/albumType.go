package types

// AlbumArticleInfo 表示专辑文章的完整信息，用于导出CSV和跟踪下载状态
type AlbumArticleInfo struct {
	Index  int    `json:"index"`  // 序号
	Title  string `json:"title"`  // 文章标题
	URL    string `json:"url"`    // 文章地址
	Status string `json:"status"` // 下载状态（可后续扩展使用）
}
