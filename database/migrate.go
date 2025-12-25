package database

import (
	"os"

	"github.com/sirupsen/logrus"
)

// MigrateFromEnv 从环境变量迁移管理员密钥（仅首次运行）
func MigrateFromEnv() error {
	adminKey := os.Getenv("ADMIN_KEY")
	if adminKey == "" {
		return nil
	}
	
	// 检查密钥是否已存在
	_, err := GetAPIKey(adminKey)
	if err == nil {
		// 密钥已存在，无需迁移
		return nil
	}
	
	if err != ErrKeyNotFound {
		return err
	}
	
	// 添加管理员密钥
	if err := AddAPIKey(adminKey, nil); err != nil {
		return err
	}
	
	logrus.Info("Admin key migrated from environment variable")
	return nil
}
