package utils

import (
	"proxypanel/config"
	"proxypanel/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB 初始化数据库连接
func InitDB() (*gorm.DB, error) {
	// 配置GORM日志
	var gormLogger logger.Interface
	if config.AppConfig.Server.Mode == "debug" {
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		gormLogger = logger.Default.LogMode(logger.Silent)
	}

	// 连接数据库
	db, err := gorm.Open(sqlite.Open(config.AppConfig.Database.Path), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移数据库表结构
	err = db.AutoMigrate(&models.ProxyRule{})
	if err != nil {
		return nil, err
	}

	// 设置到models包
	models.SetDB(db)

	return db, nil
}

// CloseDB 关闭数据库连接
func CloseDB(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}