package chat_pipeline

import (
	"SuperBizAgent/config"
	"context"
	"time"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// truncateByRune 按字符（rune）截断字符串，避免在多字节 UTF-8 字符中间切断
// maxLen 为 rune 数量上限
func truncateByRune(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "...[truncated]"
}

// createCompleteRagLambda 创建一个带有 retriever 的 Lambda，处理完整的 RAG 流程
func createCompleteRagLambda(rtr retriever.Retriever) (*compose.Lambda, error) {
	handler := func(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
		config.Info("[Lambda-CompleteRag] ========== 开始处理 ==========")
		config.Info("[Lambda-CompleteRag] 会话ID: %s", input.ID)
		config.Info("[Lambda-CompleteRag] 用户问题: %s", input.Query)
		config.Info("[Lambda-CompleteRag] 历史记录数量: %d", len(input.History))
		
		// 记录历史消息详情
		if len(input.History) > 0 {
			for i, msg := range input.History {
				content := msg.Content
				if len(content) > 100 {
					content = content[:100] + "..."
				}
				config.Info("[Lambda-CompleteRag] 历史[%d]: Role=%s, Content=%s", 
					i, msg.Role, content)
			}
		}

		config.Info("[Lambda-CompleteRag] 开始检索文档...")
		docs, err := rtr.Retrieve(ctx, input.Query)
		if err != nil {
			config.Error("[Lambda-CompleteRag] 检索失败: %v", err)
			config.Error("[Lambda-CompleteRag] 会话ID: %s, 查询: %s", input.ID, input.Query)
			config.Error("[Lambda-CompleteRag] 历史记录数量: %d", len(input.History))
			docs = []*schema.Document{}
		} else {
			config.Info("[Lambda-CompleteRag] 检索到 %d 个文档", len(docs))
		}

		var documentsContent string
		for _, doc := range docs {
			if doc.Content != "" {
				if documentsContent != "" {
					documentsContent += "\n\n"
				}
				documentsContent += doc.Content
			}
		}

		outputDocs := truncateByRune(documentsContent, 500)

		result := map[string]any{
			"documents": outputDocs,
			"content":   input.Query,
			"history":   input.History,
			"date":      time.Now().Format("2006-01-02 15:04:05"),
		}

		config.Info("[Lambda-CompleteRag] 检索文档数量: %d", len(docs))
		config.Info("[Lambda-CompleteRag] 文档内容长度: %d", len(outputDocs))
		config.Info("[Lambda-CompleteRag] 历史记录传递: %d 条", len(input.History))
		config.Info("[Lambda-CompleteRag] ========== 处理完成 ==========")
		return result, nil
	}

	return compose.AnyLambda(handler, nil, nil, nil)
}
