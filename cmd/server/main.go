// Package main 是 SuperBizAgent 服务的主入口
// 提供 AI 聊天服务、RAG 检索、秒杀业务查询等功能
// 服务端口: 6872
package main

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/agent/chat_pipeline"
	"SuperBizAgent/internal/api"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// main 函数是服务的入口点
// 执行流程:
// 1. 初始化配置和日志
// 2. 初始化 Chat Agent（包含 Milvus 客户端、嵌入器、检索器等）
// 3. 创建 Gin HTTP 引擎并配置路由
// 4. 启动 HTTP 服务
func main() {
	// 步骤1: 初始化配置和日志系统
	config.Info("[Main] 步骤1/3: 初始化配置和日志系统...")
	if err := config.InitConfig(); err != nil {
		config.Error("[Main] 配置初始化失败: %v", err)
		panic(fmt.Sprintf("配置初始化失败: %v", err))
	}
	config.Info("[Main] 配置初始化成功")

	fmt.Println("Starting SuperBizAgent service...")

	// 步骤2: 初始化聊天代理（启动时就初始化，后续复用）
	config.Info("[Main] 步骤2/3: 初始化 Chat Agent...")
	ctx := context.Background()
	if err := chat_pipeline.InitChatRunner(ctx); err != nil {
		config.Error("[Main] Chat Agent 初始化失败: %v", err)
		panic(fmt.Sprintf("Chat Agent initialization failed: %v", err))
	}
	config.Info("[Main] Chat Agent 初始化成功")

	// 步骤3: 创建 HTTP 服务
	config.Info("[Main] 步骤3/3: 创建 HTTP 服务...")
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())

	// 静态文件服务 - 前端页面（必须放在 API 路由之前）
	config.Info("[Main] 注册静态文件路由...")
	r.StaticFile("/", "./SuperBizAgentFrontend/index.html")
	r.StaticFile("/index.html", "./SuperBizAgentFrontend/index.html")
	r.StaticFile("/app.js", "./SuperBizAgentFrontend/app.js")
	r.StaticFile("/styles.css", "./SuperBizAgentFrontend/styles.css")

	// API 路由
	config.Info("[Main] 注册 API 路由...")
	r.POST("/api/chat", func(c *gin.Context) {
		chatCtrl := api.NewChatController()
		chatCtrl.Chat(c)
	})

	r.GET("/api/info", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "SuperBizAgent API",
			"version": "1.0.0",
		})
	})

	fmt.Println("Server starting on :6872")
	config.Info("[Main] ========== HTTP Server Started Successfully on :6872 ==========")
	err := http.ListenAndServe(":6872", r)
	if err != nil {
		config.Error("[Main] HTTP 服务器启动失败: %v", err)
	}
}
