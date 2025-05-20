package service

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/types"
	"go.uber.org/zap"
)

type UserService struct {
}

func NewUserService() *UserService {
	return &UserService{}
}

func (svc *UserService) SetPreferenceInfo(ctx context.Context, req types.SetPreferenceInfoRequest) (res types.SetPreferenceInfoResponse, err error) {
	zap.L().Info("SetPreferenceInfo", zap.Any("req", req))
	return
}

func (svc *UserService) GetPreferenceInfo() (res types.GetPreferenceInfoResponse, err error) {
	res.SaveImgPath = ""
	res.DownloadTimeout = 30
	res.CropImgBottomPixel = 100
	return
}
