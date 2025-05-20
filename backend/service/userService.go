package service

import (
	"github.com/pudongping/wx-graph-crawl/backend/types"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (svc *UserService) GetPreferenceInfo() (res types.GetPreferenceInfoResponse, err error) {
	res.SaveImgPath = ""
	res.DownloadTimeout = 30
	res.CropImgBottomPixel = 100
	return
}
