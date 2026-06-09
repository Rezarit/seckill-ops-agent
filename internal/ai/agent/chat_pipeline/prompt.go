package chat_pipeline

import (
	"SuperBizAgent/config"
	"context"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

type ChatTemplateConfig struct {
	FormatType schema.FormatType
	Templates  []schema.MessagesTemplate
}

// newChatTemplate component initialization function of node 'ChatTemplate' in graph 'EinoAgent'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	config.Info("[Prompt] Initializing chat template...")

	config.Info("[Prompt] System prompt (长度=%d):\n%s", len(systemPrompt), systemPrompt)

	config.Info("[Prompt] Chat template initialized with format: %v", schema.FString)

	config.Debug("[Prompt] User template: {documents}\\n\\n{content}")

	config.Info("[Prompt] Chat template ready. Variables: documents, content, history, date")

	templateConfig := &ChatTemplateConfig{
		FormatType: schema.FString,
		Templates: []schema.MessagesTemplate{
			schema.SystemMessage(systemPrompt),
			schema.MessagesPlaceholder("history", false),
			schema.UserMessage("{documents}\n\n{content}"),
		},
	}
	ctp = prompt.FromMessages(templateConfig.FormatType, templateConfig.Templates...)
	return ctp, nil
}

var systemPrompt = `你是一个智能助手，能够根据用户的问题选择合适的工具来获取信息。

工具列表：
- get_current_time：获取当前时间
- seckill_query_products：查询秒杀商品信息
- seckill_query_orders：查询订单信息
- query_internal_docs：查询内部文档和知识库
- prometheus_alerts_query：查询 Prometheus 告警信息

请根据用户的问题，判断是否需要调用工具：
- 如果用户询问时间相关问题，可以调用 get_current_time
- 如果用户询问秒杀商品相关问题，可以调用 seckill_query_products
- 如果用户询问订单相关问题，可以调用 seckill_query_orders
- 如果用户询问文档、日志、报错信息、内部知识等问题，可以调用 query_internal_docs
- 如果用户询问系统告警相关问题，可以调用 prometheus_alerts_query

如果没有合适的工具可用，或者问题可以直接回答，请直接给出答案。`
