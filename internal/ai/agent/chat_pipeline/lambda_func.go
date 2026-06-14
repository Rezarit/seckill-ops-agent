package chat_pipeline

import (
	"SuperBizAgent/config"
	"context"
	"time"

	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// newInputToRagLambda component initialization function of node 'InputToQuery' in graph 'EinoAgent'
func newInputToRagLambda(ctx context.Context, input *UserMessage, opts ...any) (output string, err error) {
	config.Info("[Lambda-InputToRag] 输入: ID=%s, Query=%s", input.ID, input.Query)
	config.Info("[Lambda-InputToRag] 输出: query_length=%d", len(input.Query))
	return input.Query, nil
}

// newInputToChatLambda component initialization function of node 'InputToHistory' in graph 'EinoAgent'
func newInputToChatLambda(ctx context.Context, input *UserMessage, opts ...any) (output map[string]any, err error) {
	return map[string]any{
		"documents": "",
		"content":   input.Query,
		"history":   input.History,
		"date":      time.Now().Format("2006-01-02 15:04:05"),
	}, nil
}

// newDocsToChatVariablesLambda converts retrieved documents to chat template variables
func newDocsToChatVariablesLambda(ctx context.Context, input map[string]any, opts ...any) (output map[string]any, err error) {
	// Extract documents from input map
	var docs []*schema.Document
	if docsInterface, ok := input["documents"]; ok {
		if docsSlice, ok := docsInterface.([]interface{}); ok {
			for _, docInterface := range docsSlice {
				if doc, ok := docInterface.(*schema.Document); ok {
					docs = append(docs, doc)
				}
			}
		}
	}

	config.Info("[Lambda] Extracted %d documents", len(docs))

	// Build context from retrieved documents
	var documentsContent string
	for _, doc := range docs {
		if doc.Content != "" {
			if documentsContent != "" {
				documentsContent += "\n\n"
			}
			documentsContent += doc.Content
		}
	}

	// Get query from input if available
	var queryContent string
	if query, ok := input["query"]; ok {
		if q, ok := query.(string); ok {
			queryContent = q
		}
	}

	config.Info("[Lambda] Built context: documents_length=%d, query_length=%d", len(documentsContent), len(queryContent))

	// 限制最终输出的 documents 长度
	outputDocs := documentsContent
	if len(outputDocs) > 500 {
		outputDocs = outputDocs[:500] + "...[truncated]"
	}

	return map[string]any{
		"documents": outputDocs,
		"content":   queryContent,
		"history":   []*schema.Message{},
		"date":      time.Now().Format("2006-01-02 15:04:05"),
	}, nil
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

		outputDocs := documentsContent
		if len(outputDocs) > 500 {
			outputDocs = outputDocs[:500] + "...[truncated]"
		}

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
