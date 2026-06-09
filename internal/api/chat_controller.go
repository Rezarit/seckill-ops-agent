// Package api 提供 HTTP API 控制器实现
// 包含聊天接口、健康检查等 API 端点
package api

import (
	"net/http"

	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/agent/chat_pipeline"
	"SuperBizAgent/utility/mem"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

// ChatController 聊天控制器
// 处理用户的聊天请求，调用 Chat Agent 进行响应
type ChatController struct{}

// NewChatController 创建聊天控制器实例
func NewChatController() *ChatController {
	return &ChatController{}
}

// ChatReq 聊天请求结构体
type ChatReq struct {
	Id       string `json:"Id" binding:"required"`       // 用户会话ID
	Question string `json:"Question" binding:"required"` // 用户提问内容
}

// ChatRes 聊天响应结构体
type ChatRes struct {
	Answer string `json:"answer"` // AI 回复内容
}

// Chat 处理聊天请求
// 执行流程:
// 1. 解析请求参数
// 2. 获取聊天 Runner
// 3. 构建用户消息（包含历史记录）
// 4. 调用 Runner 进行推理
// 5. 更新对话历史
// 6. 返回响应
func (cc *ChatController) Chat(c *gin.Context) {
	config.Info("[ChatController] ========== 收到聊天请求 ==========")

	// 步骤1: 解析请求参数
	var req ChatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		config.Error("[ChatController] 请求参数解析失败: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	config.Info("[ChatController] 请求参数:")
	config.Info("[ChatController]   - Id: %s", req.Id)
	config.Info("[ChatController]   - Question: %s", truncateText(req.Question, 100))

	// 步骤2: 获取聊天 Runner
	config.Info("[ChatController] 获取聊天 Runner...")
	runner, err := chat_pipeline.GetChatRunner(c.Request.Context())
	if err != nil {
		config.Error("[ChatController] 获取聊天 Runner 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get chat runner",
			"error":   err.Error(),
		})
		return
	}
	config.Info("[ChatController] 聊天 Runner 获取成功")

	// 步骤3: 构建用户消息（包含历史记录）
	config.Info("[ChatController] 构建用户消息...")
	userMessage := &chat_pipeline.UserMessage{
		ID:      req.Id,
		Query:   req.Question,
		History: mem.GetSimpleMemory(req.Id).GetMessages(),
	}
	config.Info("[ChatController] 用户消息构建完成")

	// 步骤4: 调用 Runner 进行推理
	config.Info("[ChatController] 调用 Chat Agent...")
	out, err := runner.Invoke(c.Request.Context(), userMessage, compose.WithCallbacks())
	if err != nil {
		config.Error("[ChatController] 调用 Chat Agent 失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Chat failed",
			"error":   err.Error(),
		})
		return
	}

	// 步骤5: 更新对话历史
	config.Info("[ChatController] 更新对话历史...")
	mem.GetSimpleMemory(req.Id).SetMessages(schema.UserMessage(req.Question))
	mem.GetSimpleMemory(req.Id).SetMessages(schema.AssistantMessage(out.Content, nil))
	config.Info("[ChatController] 对话历史更新完成")

	// 步骤6: 返回响应
	config.Info("[ChatController] ========== 响应准备完成 ==========")
	config.Info("[ChatController] 响应内容长度: %d", len(out.Content))

	c.JSON(http.StatusOK, ChatRes{
		Answer: out.Content,
	})
}

// truncateText 截断文本用于日志输出
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}
