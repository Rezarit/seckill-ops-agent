// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"context"

	e_mcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetLogMcpTool 获取日志查询 MCP 工具
// MCP (Model Context Protocol) 是用于扩展 AI Agent 能力的协议
// 该工具连接到 MCP 服务器获取日志查询能力
//
// 相关文档:
// - https://cloud.tencent.com/developer/mcp/server/11710
// - https://cloud.tencent.com/document/product/614/118699
// - https://www.cloudwego.io/zh/docs/eino/ecosystem_integration/tool/tool_mcp/
// - https://mcp-go.dev/clients
func GetLogMcpTool() ([]tool.BaseTool, error) {
	config.Info("[Tool-GetLogMcpTool] 初始化 MCP 日志工具...")

	mcpURL := config.AppConfig.MCPURL
	config.Info("[Tool-GetLogMcpTool] MCP URL: %s", mcpURL)

	ctx := context.Background()

	// 创建 MCP 客户端
	cli, err := client.NewSSEMCPClient(mcpURL)
	if err != nil {
		config.Error("[Tool-GetLogMcpTool] 创建 MCP 客户端失败: %v", err)
		return []tool.BaseTool{}, err
	}

	// 启动客户端
	err = cli.Start(ctx)
	if err != nil {
		config.Error("[Tool-GetLogMcpTool] 启动 MCP 客户端失败: %v", err)
		return []tool.BaseTool{}, err
	}

	// 初始化请求
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "SuperBizAgent",
		Version: "1.0.0",
	}

	// 发送初始化请求
	if _, err = cli.Initialize(ctx, initRequest); err != nil {
		config.Error("[Tool-GetLogMcpTool] MCP 初始化失败: %v", err)
		return []tool.BaseTool{}, err
	}

	// 获取 MCP 工具列表
	mcpTools, err := e_mcp.GetTools(ctx, &e_mcp.Config{Cli: cli})
	if err != nil {
		config.Error("[Tool-GetLogMcpTool] 获取 MCP 工具失败: %v", err)
		return []tool.BaseTool{}, err
	}

	config.Info("[Tool-GetLogMcpTool] MCP 工具初始化成功，获取到 %d 个工具", len(mcpTools))
	return mcpTools, nil
}
