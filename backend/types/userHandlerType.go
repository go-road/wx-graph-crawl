package types

type SetPreferenceInfoRequest struct {
	SaveImgPath        string `json:"save_img_path"`         // 图片保存路径
	DownloadTimeout    int    `json:"download_timeout"`      // 下载超时时间
	CropImgBottomPixel int    `json:"crop_img_bottom_pixel"` // 裁剪图片底部像素
}

type SetPreferenceInfoResponse struct {
	UpdatedTime int64 `json:"updated_time"` // 更新时间
}

type GetPreferenceInfoResponse struct {
	SaveImgPath        string `json:"save_img_path"`         // 图片保存路径
	DownloadTimeout    int    `json:"download_timeout"`      // 下载超时时间
	CropImgBottomPixel int    `json:"crop_img_bottom_pixel"` // 裁剪图片底部像素
	UpdatedTime        int64  `json:"updated_time"`          // 更新时间
}
