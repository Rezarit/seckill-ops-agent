// Package config 秒杀系统配置管理
package config

import (
	"fmt"
	"sync"

	"SuperBizAgent/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	dbInstance *gorm.DB
	dbOnce     sync.Once
)

// GetConfig 获取秒杀系统配置
func GetConfig() *config.DatabaseConfig {
	return &config.AppConfig.Database
}

// GetDB 获取秒杀系统数据库连接（单例）
func GetDB() (*gorm.DB, error) {
	var err error
	dbOnce.Do(func() {
		cfg := config.AppConfig.Database
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

		dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			return
		}

		// 配置连接池
		sqlDB, err := dbInstance.DB()
		if err != nil {
			return
		}
		sqlDB.SetMaxIdleConns(10)   // 最大空闲连接数
		sqlDB.SetMaxOpenConns(100)  // 最大打开连接数
		sqlDB.SetConnMaxLifetime(0) // 连接最大生命周期
	})
	return dbInstance, err
}
