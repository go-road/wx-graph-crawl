package global

import (
	"github.com/jmoiron/sqlx"
)

var (
	RootPath string // 项目根目录
	DB       *sqlx.DB
)
