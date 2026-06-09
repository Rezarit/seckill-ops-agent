// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/retriever"
	"context"
	"encoding/json"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// QueryInternalDocsInput 内部文档查询工具的输入参数
type QueryInternalDocsInput struct {
	Query string `json:"query" jsonschema:"description=The query string to search in internal documentation for relevant information and processing steps"`
}

// NewQueryInternalDocsTool 创建内部文档查询工具
// 执行流程:
// 1. 获取 Milvus 检索器
// 2. 执行文档检索
// 3. 处理检索结果
// 4. 返回匹配的文档列表
func NewQueryInternalDocsTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"query_internal_docs",
		"Use this tool to search internal documentation and knowledge base for relevant information.",
		func(ctx context.Context, input *QueryInternalDocsInput, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-query_internal_docs] ========== 工具调用开始 ==========")
			config.Info("[Tool-query_internal_docs] 输入参数: Query=%s, QueryLength=%d", truncateText(input.Query, 50), len(input.Query))

			// 步骤1: 获取 Milvus 检索器（使用全局单例）
			config.Info("[Tool-query_internal_docs] 步骤1/4: 获取 Milvus 检索器...")
			rr, err := retriever.GetMilvusRetriever(ctx)
			if err != nil {
				config.Error("[Tool-query_internal_docs] 获取检索器失败: %v", err)
				config.Info("[Tool-query_internal_docs] ========== 工具调用结束 (失败) ==========")
				return `{"message":"文档检索服务暂不可用，请稍后重试"}`, nil
			}
			config.Info("[Tool-query_internal_docs] 检索器获取成功")

			// 步骤2: 执行文档检索
			config.Info("[Tool-query_internal_docs] 步骤2/4: 执行文档检索...")
			start := time.Now()
			docs, err := rr.Retrieve(ctx, input.Query)
			elapsed := time.Since(start)
			config.Info("[Tool-query_internal_docs] 检索耗时: %v", elapsed)

			if err != nil {
				config.Error("[Tool-query_internal_docs] 检索执行失败: %v", err)
				config.Info("[Tool-query_internal_docs] ========== 工具调用结束 (失败) ==========")
				return `{"message":"检索失败，请稍后重试"}`, nil
			}

			// 步骤3: 处理检索结果
			config.Info("[Tool-query_internal_docs] 步骤3/4: 处理检索结果...")
			config.Info("[Tool-query_internal_docs] 检索结果数量: %d", len(docs))

			// 处理空结果
			if len(docs) == 0 {
				config.Warn("[Tool-query_internal_docs] 未找到匹配的文档")
				config.Warn("[Tool-query_internal_docs] 可能原因:")
				config.Warn("[Tool-query_internal_docs]   1) 查询词与知识库不匹配")
				config.Warn("[Tool-query_internal_docs]   2) 向量索引问题")
				config.Warn("[Tool-query_internal_docs]   3) 数据未正确加载")
				config.Info("[Tool-query_internal_docs] ========== 工具调用结束 (空结果) ==========")
				return `{"message":"未找到相关文档"}`, nil
			}

			// 步骤4: 记录匹配文档详情
			config.Info("[Tool-query_internal_docs] 步骤4/4: 记录匹配文档详情...")
			for i, doc := range docs {
				contentLen := len(doc.Content)
				config.Info("[Tool-query_internal_docs] 文档 %d: ID=%s, ContentLength=%d", i+1, doc.ID, contentLen)
				if contentLen > 0 {
					preview := doc.Content
					if contentLen > 150 {
						preview = doc.Content[:150] + "..."
					}
					config.Info("[Tool-query_internal_docs] 文档 %d 预览: %s", i+1, preview)
				} else {
					config.Warn("[Tool-query_internal_docs] 文档 %d: 内容为空", i+1)
				}
			}

			// 返回结果
			respBytes, _ := json.Marshal(docs)
			config.Info("[Tool-query_internal_docs] 返回结果长度: %d", len(respBytes))
			config.Info("[Tool-query_internal_docs] ========== 工具调用结束 (成功) ==========")
			return string(respBytes), nil
		})
	if err != nil {
		config.Error("[Tool-query_internal_docs] 工具创建失败: %v", err)
		panic(err)
	}
	return t
}
