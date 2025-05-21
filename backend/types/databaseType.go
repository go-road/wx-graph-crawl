package types

type SystemConfig struct {
	ID        int64  `db:"id"`         // 自增主键
	Key       string `db:"key"`        // 配置项的key
	Content   string `db:"content"`    // 配置项的内容
	Version   int    `db:"version"`    // 版本号
	CreatedAt int64  `db:"created_at"` // 创建时间
	UpdatedAt int64  `db:"updated_at"` // 更新时间
}
