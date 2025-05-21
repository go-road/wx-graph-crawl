package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/pudongping/wx-graph-crawl/backend/constant"
	"github.com/pudongping/wx-graph-crawl/backend/global"
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
	now := time.Now().Unix()
	prefJson, err := json.Marshal(req)
	if err != nil {
		zap.L().Error("SetPreferenceInfo Marshal", zap.Error(err))
		return res, err
	}

	// 使用 UPSERT 语法更新或插入记录
	query := `
		INSERT INTO system_configs (key, content, version, created_at, updated_at)
		VALUES (?, ?, 1, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
		content = excluded.content,
		version = version + 1,
		updated_at = excluded.updated_at
	`
	_, err = global.DB.Exec(query, constant.SystemConfigKeyPreferenceInfo, string(prefJson), now, now)
	if err != nil {
		zap.L().Error("SetPreferenceInfo Exec", zap.Error(err))
		return res, err
	}
	res.UpdatedTime = now

	return
}

func (svc *UserService) GetPreferenceInfo() (*types.GetPreferenceInfoResponse, error) {
	var (
		err error
		res types.GetPreferenceInfoResponse
		sc  types.SystemConfig
	)
	err = global.DB.Get(&sc, "SELECT * FROM system_configs WHERE key = ?", constant.SystemConfigKeyPreferenceInfo)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			zap.L().Info("没有找到偏好信息")
			return nil, nil
		}
		zap.L().Error("GetPreferenceInfo", zap.Error(err))
		return nil, err
	}

	// 如果内容为空
	if sc.Content == "" {
		return nil, err
	}

	// 解析内容
	err = json.Unmarshal([]byte(sc.Content), &res)
	if err != nil {
		zap.L().Error("GetPreferenceInfo Unmarshal", zap.Error(err))
		return nil, err
	}

	res.UpdatedTime = sc.UpdatedAt

	return &res, nil
}
