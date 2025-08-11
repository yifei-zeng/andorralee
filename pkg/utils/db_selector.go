package utils

import (
	"andorralee/internal/config"

	"gorm.io/gorm"
)

// SelectActiveDB 返回当前激活的数据库连接（根据 DB_MODE 环境变量）
// 优先：配置指定 -> 连接存在 -> 其它可用数据库回退
func SelectActiveDB() *gorm.DB {
	return config.GetActiveDB()
}
