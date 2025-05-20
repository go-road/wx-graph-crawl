package types

type SelectFileResponse struct {
	FilePath  string   `json:"file_path"`  // 选择的文件路径
	ValidURLs []string `json:"valid_urls"` // 从文件中读取的有效URL列表
}
