// Package config 提供全局配置管理和日志工具
// 负责从配置文件加载应用参数，包括服务器端口、数据库连接、AI模型配置等
// 同时提供统一的日志接口（Info/Error/Warn/Debug），支持 Zap 和标准库 log 双模式
package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 应用全局配置结构体
// 包含服务器、数据库、AI模型等所有配置项
type Config struct {
	Server      ServerConfig   // 服务器配置
	Database    DatabaseConfig // 数据库配置（秒杀系统）
	DoubaoChat  ModelConfig    // 豆包思考模型配置
	DoubaoQuick ModelConfig    // 豆包快速模型配置
	DoubaoEmb   ModelConfig    // 豆包嵌入模型配置
	DSTHINK     ModelConfig    // 大模型思考配置
	DSQUICK     ModelConfig    // 大模型快速配置
	MCPURL      string         // MCP服务地址
	FileDir     string         // 文件存储目录
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int // 服务端口，默认6872
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string // 数据库主机地址
	Port     int    // 数据库端口
	User     string // 数据库用户名
	Password string // 数据库密码
	DBName   string // 数据库名称
}

// ModelConfig AI模型配置
type ModelConfig struct {
	APIKey       string // API密钥
	BaseURL      string // API基础地址
	Model        string // 模型名称
	EmbeddingDim int    // 嵌入向量维度（仅嵌入模型使用）
}

// AppConfig 全局配置实例
var AppConfig *Config

// Logger 全局日志实例
var Logger *zap.Logger

// InitConfig 初始化配置和日志系统
// 执行流程:
// 1. 初始化 Zap 日志框架
// 2. 配置 Viper 读取配置文件路径
// 3. 加载配置项到 AppConfig
// 4. 设置默认值（如端口）
// 返回: 初始化错误
func InitConfig() error {
	Info("[Config] ========== 初始化配置系统 ==========")

	// 步骤1: 初始化 Zap 日志
	Info("[Config] 步骤1/3: 初始化 Zap 日志框架...")
	var err error
	Logger, err = zap.NewProduction()
	if err != nil {
		Error("[Config] Zap 日志初始化失败: %v", err)
		return fmt.Errorf("failed to initialize logger: %v", err)
	}
	Info("[Config] Zap 日志初始化成功")

	// 步骤2: 配置 Viper 读取配置文件
	Info("[Config] 步骤2/3: 配置 Viper 读取配置文件...")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./manifest/config")

	if err := viper.ReadInConfig(); err != nil {
		Warn("[Config] 读取配置文件失败，将使用默认值: %v", err)
	} else {
		Info("[Config] 配置文件读取成功: %s", viper.ConfigFileUsed())
	}

	// 步骤3: 加载配置到 AppConfig
	Info("[Config] 步骤3/3: 加载配置项到 AppConfig...")
	AppConfig = &Config{
		Server: ServerConfig{
			Port: viper.GetInt("server.port"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("seckill_database.host"),
			Port:     viper.GetInt("seckill_database.port"),
			User:     viper.GetString("seckill_database.user"),
			Password: viper.GetString("seckill_database.password"),
			DBName:   viper.GetString("seckill_database.dbname"),
		},
		DoubaoChat: ModelConfig{
			APIKey:  viper.GetString("doubao_think_chat_model.api_key"),
			BaseURL: viper.GetString("doubao_think_chat_model.base_url"),
			Model:   viper.GetString("doubao_think_chat_model.model"),
		},
		DoubaoQuick: ModelConfig{
			APIKey:  viper.GetString("doubao_quick_chat_model.api_key"),
			BaseURL: viper.GetString("doubao_quick_chat_model.base_url"),
			Model:   viper.GetString("doubao_quick_chat_model.model"),
		},
		DoubaoEmb: ModelConfig{
			APIKey:       viper.GetString("doubao_embedding_model.api_key"),
			BaseURL:      viper.GetString("doubao_embedding_model.base_url"),
			Model:        viper.GetString("doubao_embedding_model.model"),
			EmbeddingDim: viper.GetInt("doubao_embedding_model.embedding_dim"),
		},
		DSTHINK: ModelConfig{
			APIKey:  viper.GetString("ds_think_chat_model.api_key"),
			BaseURL: viper.GetString("ds_think_chat_model.base_url"),
			Model:   viper.GetString("ds_think_chat_model.model"),
		},
		DSQUICK: ModelConfig{
			APIKey:  viper.GetString("ds_quick_chat_model.api_key"),
			BaseURL: viper.GetString("ds_quick_chat_model.base_url"),
			Model:   viper.GetString("ds_quick_chat_model.model"),
		},
		MCPURL:  viper.GetString("mcp_url"),
		FileDir: viper.GetString("file_dir"),
	}

	// 设置默认端口
	if AppConfig.Server.Port == 0 {
		AppConfig.Server.Port = 6872
		Info("[Config] 使用默认端口: %d", AppConfig.Server.Port)
	}

	// 校验配置
	if err := AppConfig.Validate(); err != nil {
		Error("[Config] 配置校验失败: %v", err)
		return fmt.Errorf("配置校验失败: %v", err)
	}

	// 记录配置摘要（敏感信息脱敏）
	Info("[Config] ========== 配置加载完成 ==========")
	Info("[Config] 服务器端口: %d", AppConfig.Server.Port)
	Info("[Config] 文件目录: %s", AppConfig.FileDir)
	Info("[Config] 聊天模型: %s", AppConfig.DoubaoChat.Model)
	Info("[Config] 嵌入模型: %s (维度: %d)", AppConfig.DoubaoEmb.Model, AppConfig.DoubaoEmb.EmbeddingDim)

	return nil
}

func GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		AppConfig.Database.User,
		AppConfig.Database.Password,
		AppConfig.Database.Host,
		AppConfig.Database.Port,
		AppConfig.Database.DBName,
	)
}

func Info(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Info(fmt.Sprintf(format, args...))
	} else {
		log.Printf(format, args...)
	}
}

func Error(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Error(fmt.Sprintf(format, args...))
	} else {
		log.Printf(format, args...)
	}
}

func Warn(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Warn(fmt.Sprintf(format, args...))
	} else {
		log.Printf(format, args...)
	}
}

func Debug(format string, args ...interface{}) {
	if Logger != nil {
		Logger.Debug(fmt.Sprintf(format, args...))
	} else {
		log.Printf(format, args...)
	}
}

// Validate 校验配置的有效性，设置默认值
// 返回: 校验错误（如果有）
func (c *Config) Validate() error {
	// 校验服务器配置
	if c.Server.Port <= 0 {
		c.Server.Port = 6872
	}
	if c.Server.Port < 1024 || c.Server.Port > 65535 {
		return fmt.Errorf("服务器端口必须在 1024-65535 范围内，当前值: %d", c.Server.Port)
	}

	// 校验嵌入模型配置
	if c.DoubaoEmb.EmbeddingDim <= 0 {
		c.DoubaoEmb.EmbeddingDim = 384 // 默认维度
		Warn("[Config] 嵌入维度未配置，使用默认值: 384")
	}

	// 校验 API Key（警告级别，不阻止启动）
	if c.DoubaoChat.APIKey == "" {
		Warn("[Config] 豆包聊天模型 API Key 未配置")
	}
	if c.DoubaoEmb.APIKey == "" {
		Warn("[Config] 豆包嵌入模型 API Key 未配置")
	}

	// 校验数据库配置（如果配置了数据库地址）
	if c.Database.Host != "" && c.Database.User == "" {
		Warn("[Config] 数据库已配置但用户名为空")
	}

	return nil
}
