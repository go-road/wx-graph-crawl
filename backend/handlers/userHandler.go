package handlers

import (
	"context"

	"github.com/pudongping/wx-graph-crawl/backend/types"
)

type User struct {
	ctx context.Context
}

func NewUser(ctx context.Context) *User {
	return &User{
		ctx: ctx,
	}
}

func (u *User) GetPreferenceInfo() types.PreferenceSet {
	return types.PreferenceSet{
		SaveImgPath:        "D:/wx-graph-crawl/img",
		DownloadTimeout:    30,
		CropImgBottomPixel: 100,
	}
}
