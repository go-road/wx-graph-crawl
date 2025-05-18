package types

type CropResult struct {
	ImgPath string // 图片存储的硬盘路径地址
	Err     error  // 裁剪过程中出现的错误
}
