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
	dbErr      error
	dbMu       sync.RWMutex
)

// GetConfig 获取秒杀系统配置
func GetConfig() *config.DatabaseConfig {
	return &config.AppConfig.Database
}

// GetDB 获取秒杀系统数据库连接（支持失败重试的单例模式）
// 首次初始化失败后，后续调用会自动重试
func GetDB() (*gorm.DB, error) {
	// 快速路径：已有有效连接
	dbMu.RLock()
	db := dbInstance
	err := dbErr
	dbMu.RUnlock()

	if db != nil && err == nil {
		return db, nil
	}

	// 慢路径：初始化或重试
	dbMu.Lock()
	defer dbMu.Unlock()

	// 双重检查
	if dbInstance != nil && dbErr == nil {
		return dbInstance, nil
	}

	// 尝试初始化（支持失败后重试）
	cfg := config.AppConfig.Database
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	dbInstance, dbErr = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if dbErr != nil {
		return nil, dbErr
	}

	// 配置连接池
	sqlDB, err := dbInstance.DB()
	if err != nil {
		dbErr = err
		return nil, dbErr
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(0)

	return dbInstance, nil
}
