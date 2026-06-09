// Package chat_pipeline 提供 ReAct 推理代理的初始化封装
package chat_pipeline

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/tools"
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
)

// newReactAgentLambda 创建 ReAct 推理代理
// 执行流程:
// 1. 配置代理参数（最大步骤数等）
// 2. 初始化聊天模型
// 3. 注册各种工具（MCP工具、Prometheus告警、时间查询、内部文档查询、秒杀系统工具等）
// 4. 创建 ReAct 代理实例
// 5. 包装为 Lambda
func newReactAgentLambda(ctx context.Context) (lba *compose.Lambda, err error) {
	config.Info("[Agent] 初始化 ReAct 代理...")

	// 配置代理参数
	agentConfig := &react.AgentConfig{
		MaxStep:            25,
		ToolReturnDirectly: map[string]struct{}{}}

	// 步骤1: 初始化聊天模型
	config.Info("[Agent] 步骤1/2: 初始化聊天模型...")
	chatModelIns11, err := newChatModel(ctx)
	if err != nil {
		config.Error("[Agent] 初始化聊天模型失败: %v", err)
		return nil, err
	}
	config.Info("[Agent] 聊天模型初始化成功")
	agentConfig.ToolCallingModel = chatModelIns11

	// 步骤2: 注册工具（使用降级方案，工具创建失败不影响服务启动）
	config.Info("[Agent] 步骤2/2: 注册工具...")

	// 工具注册辅助函数：安全注册工具，忽略失败的工具
	registerTool := func(name string, t tool.InvokableTool) {
		if t != nil {
			agentConfig.ToolsConfig.Tools = append(agentConfig.ToolsConfig.Tools, t)
			config.Info("[Agent] 注册工具: %s", name)
		} else {
			config.Warn("[Agent] 跳过工具: %s (创建失败)", name)
		}
	}

	// MCP 工具（日志查询等）
	mcpTool, err := tools.GetLogMcpTool()
	if err != nil {
		config.Warn("[Agent] MCP 服务未配置，跳过")
	} else if len(mcpTool) > 0 {
		agentConfig.ToolsConfig.Tools = append(agentConfig.ToolsConfig.Tools, mcpTool...)
		config.Info("[Agent] 注册 MCP 工具: %d 个", len(mcpTool))
	}

	// Prometheus 告警查询工具
	registerTool("prometheus_alerts_query", tools.NewPrometheusAlertsQueryTool())

	// 当前时间查询工具
	registerTool("get_current_time", tools.NewGetCurrentTimeTool())

	// 内部文档 RAG 查询工具
	registerTool("query_internal_docs", tools.NewQueryInternalDocsTool())

	// 秒杀系统工具集
	registerTool("seckill_query_products", tools.NewSeckillQueryProductsTool())
	registerTool("seckill_query_orders", tools.NewSeckillQueryOrdersTool())
	registerTool("seckill_analyze_data", tools.NewSeckillAnalyzeDataTool())

	config.Info("[Agent] 工具注册完成，共 %d 个工具", len(agentConfig.ToolsConfig.Tools))

	// 创建 ReAct 代理实例
	config.Info("[Agent] 创建 ReAct 代理实例...")
	ins, err := react.NewAgent(ctx, agentConfig)
	if err != nil {
		config.Error("[Agent] 创建 ReAct 代理失败: %v", err)
		return nil, err
	}

	// 包装为 Lambda
	config.Info("[Agent] 包装为 Lambda...")
	lba, err = compose.AnyLambda(ins.Generate, ins.Stream, nil, nil)
	if err != nil {
		config.Error("[Agent] 创建 Lambda 包装器失败: %v", err)
		return nil, err
	}

	config.Info("[Agent] ReAct 代理初始化成功，最大步骤数: %d", agentConfig.MaxStep)
	return lba, nil
}
