package types

type PreferenceSet struct {
	SaveImgPath        string `json:"save_img_path"`         // 图片保存路径
	DownloadTimeout    int    `json:"download_timeout"`      // 下载超时时间
	CropImgBottomPixel int    `json:"crop_img_bottom_pixel"` // 裁剪图片底部像素
}
