// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// GetCurrentTimeInput 获取当前时间的输入参数（无需输入）
type GetCurrentTimeInput struct {
	// 无需输入参数
}

// GetCurrentTimeOutput 获取当前时间的输出结果
type GetCurrentTimeOutput struct {
	Success      bool   `json:"success" jsonschema:"description=Indicates whether the time retrieval was successful"`
	Seconds      int64  `json:"seconds" jsonschema:"description=Current Unix timestamp in seconds since epoch (1970-01-01 00:00:00 UTC)"`
	Milliseconds int64  `json:"milliseconds" jsonschema:"description=Current Unix timestamp in milliseconds since epoch (1970-01-01 00:00:00 UTC)"`
	Microseconds int64  `json:"microseconds" jsonschema:"description=Current Unix timestamp in microseconds since epoch (1970-01-01 00:00:00 UTC)"`
	Timestamp    string `json:"timestamp" jsonschema:"description=Human-readable timestamp in format 'YYYY-MM-DD HH:MM:SS.microseconds'"`
	Message      string `json:"message" jsonschema:"description=Status message describing the operation result"`
}

// NewGetCurrentTimeTool 创建获取当前时间的工具
// 返回多种格式的当前时间：秒级时间戳、毫秒级时间戳、微秒级时间戳和可读格式
func NewGetCurrentTimeTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"get_current_time",
		"Get current system time in multiple formats. Returns the current time in seconds (Unix timestamp), milliseconds, and microseconds. Use this tool when you need to retrieve current system time for logging, timing operations, or timestamping events.",
		func(ctx context.Context, input *GetCurrentTimeInput, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-get_current_time] 工具调用开始")

			// 获取当前时间
			now := time.Now()

			// 计算各种时间格式
			seconds := now.Unix()                                 // 秒级时间戳
			milliseconds := now.UnixMilli()                       // 毫秒级时间戳
			microseconds := now.UnixMicro()                       // 微秒级时间戳
			timestamp := now.Format("2006-01-02 15:04:05.000000") // 可读格式

			config.Info("[Tool-get_current_time] 当前时间: %s", timestamp)

			// 构建输出结果
			timeOutput := GetCurrentTimeOutput{
				Success:      true,
				Seconds:      seconds,
				Milliseconds: milliseconds,
				Microseconds: microseconds,
				Timestamp:    timestamp,
				Message:      "Current time retrieved successfully",
			}

			// 转换为 JSON 格式
			jsonBytes, err := json.MarshalIndent(timeOutput, "", "  ")
			if err != nil {
				config.Error("[Tool-get_current_time] JSON 序列化失败: %v", err)
				return "", err
			}

			config.Info("[Tool-get_current_time] 工具调用完成")
			return string(jsonBytes), nil
		})

	if err != nil {
		config.Error("[Tool-get_current_time] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}

	return t
}
