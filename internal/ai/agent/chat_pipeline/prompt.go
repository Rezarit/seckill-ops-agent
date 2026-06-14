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

var systemPrompt = `你是一个专业的企业智能助手，擅长理解用户意图并选择合适的工具解决问题。

=== 工具使用指南 ===

【运维监控工具】
工具名: prometheus_alerts_query
作用: 查询系统监控告警信息
适用场景: 
- 检查系统是否有告警
- 查询 CPU/内存/磁盘等资源告警
- 检查服务状态是否正常
问题示例: "系统有什么告警？"、"检查服务器状态"、"有没有服务宕机？"、"告警怎么处理？"

【知识库工具】
工具名: query_internal_docs
作用: 查询内部技术文档和解决方案
适用场景: 
- 查找技术问题解决方案
- 查询配置文档
- 获取故障排查指南
问题示例: "连接池怎么配置？"、"如何优化性能？"、"数据库报错怎么解决？"

【业务工具】
工具名: seckill_query_products
作用: 查询秒杀商品信息
适用场景: 查询正在进行的秒杀活动商品
问题示例: "今天有什么秒杀商品？"、"秒杀活动有哪些商品？"

工具名: seckill_query_orders
作用: 查询秒杀订单信息
适用场景: 查询用户的秒杀订单
问题示例: "我的订单在哪里？"、"查询订单状态"

工具名: seckill_analyze_data
作用: 分析秒杀数据
适用场景: 分析秒杀活动数据
问题示例: "秒杀数据分析"、"活动效果如何？"

【通用工具】
工具名: get_current_time
作用: 获取当前时间
适用场景: 询问当前日期或时间
问题示例: "现在几点了？"、"今天是几号？"

=== 操作规则 ===

1. 仔细分析用户的问题，理解真实意图
2. 参考对话历史，理解上下文关联
3. 如果上一轮对话讨论的是某个主题，本轮回答要保持连贯性
4. 选择最合适的工具执行查询
5. 如果没有合适的工具或问题可以直接回答，请直接给出友好的回答
6. 使用工具后，请用自然语言总结结果给用户

=== 回答要求 ===

- 回答要详细、专业
- 使用中文，语气友好
- 结构化展示信息（如列表、表格）
- 对于告警等运维信息，提供处理建议
- 保持上下文连贯性，记住之前的对话内容`
