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
