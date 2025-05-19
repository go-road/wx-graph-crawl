package types

type CrawlingRequest struct {
	ImgSavePath    string   `json:"img_save_path"`   // 图片保存路径
	ImgUrls        []string `json:"img_urls"`        // 图片链接地址
	TimeoutSeconds int64    `json:"timeout_seconds"` // 下载超时时间
}

type CrawlingResponse struct {
	TextContentSavePath string `json:"text_content_save_path"` // 文字内容保存路径
	CrawlUrlCount       int64  `json:"crawl_url_count"`        // 抓取的链接地址数量
	CrawlImgCount       int64  `json:"crawl_img_count"`        // 抓取成功并保存成功的图片数量
	ErrContent          string `json:"err_content"`            // 错误信息
	CastTimeStr         string `json:"cast_time_str"`          // 耗时字符串
}

type CroppingRequest struct {
	ImgSavePath string `json:"img_save_path"` // 图片保存路径
	BottomPixel int    `json:"bottom_pixel"`  // 裁剪图片底部像素
}

type CroppingResponse struct {
	CropImgPath  string `json:"crop_img_path"`  // 裁剪图片保存路径
	CropImgCount int64  `json:"crop_img_count"` // 裁剪成功的图片数量
	ErrContent   string `json:"err_content"`    // 错误信息
	CastTimeStr  string `json:"cast_time_str"`  // 耗时字符串
}

type ShufflingRequest struct {
	ImgSavePath string `json:"img_save_path"` // 图片保存路径
	MaxNumImage int    `json:"max_num_image"` // 当一个目录中的图片超过多少张时，开始拆分目录
}

type ShufflingResponse struct {
	ShuffleImgPath string `json:"shuffle_img_path"` // 打乱的图片路径
	CastTimeStr    string `json:"cast_time_str"`    // 耗时字符串
}
