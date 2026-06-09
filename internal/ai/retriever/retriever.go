// Package retriever 提供向量检索功能封装
// 实现基于 Milvus 的 RAG 检索，将用户查询转换为向量后进行相似度搜索
// 使用 sync.Once 保证全局单例
package retriever

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/embedder"
	"SuperBizAgent/internal/ai/pkg/milvus"
	"SuperBizAgent/utility/common"
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/components/retriever"
	"github.com/cloudwego/eino/schema"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

var (
	milvusRetriever     retriever.Retriever // 全局单例检索器
	milvusRetrieverOnce sync.Once           // 确保只初始化一次
	milvusRetrieverErr  error               // 初始化错误
)

// GetMilvusRetriever 获取全局单例 Milvus Retriever
// 使用 sync.Once 保证线程安全的懒加载
func GetMilvusRetriever(ctx context.Context) (retriever.Retriever, error) {
	milvusRetrieverOnce.Do(func() {
		config.Info("[Retriever] ========== 首次初始化 Milvus Retriever（全局单例）==========")
		milvusRetriever, milvusRetrieverErr = initMilvusRetriever(ctx)
		if milvusRetrieverErr != nil {
			config.Error("[Retriever] 全局 Retriever 初始化失败: %v", milvusRetrieverErr)
		} else {
			config.Info("[Retriever] 全局 Retriever 初始化成功")
		}
	})
	return milvusRetriever, milvusRetrieverErr
}

// NewMilvusRetriever 保留旧接口以兼容外部调用
// 内部委托给 GetMilvusRetriever
func NewMilvusRetriever(ctx context.Context) (retriever.Retriever, error) {
	return GetMilvusRetriever(ctx)
}

// EmptyRetriever 空检索器实现
// 当 Milvus 连接失败时使用，返回空结果集
type EmptyRetriever struct{}

// Retrieve 实现空检索，返回空结果
func (r *EmptyRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	config.Warn("[Retriever] 使用空检索器，返回空结果")
	return []*schema.Document{}, nil
}

// CustomMilvusRetriever 自定义 Milvus 检索器
// 直接使用 Milvus SDK 进行向量搜索，提供更好的控制和调试能力
type CustomMilvusRetriever struct {
	client       *milvusclient.Client // Milvus 客户端
	eb           embedding.Embedder   // 嵌入器
	collection   string               // Collection 名称
	vectorField  string               // 向量字段名
	outputFields []string             // 查询返回的字段
	topK         int                  // 返回结果数量
}

// Retrieve 执行向量检索
// 执行流程:
// 1. 将查询文本向量化
// 2. 转换向量类型为 FloatVector
// 3. 执行 Milvus 搜索
// 4. 解析搜索结果
func (r *CustomMilvusRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	config.Info("[Retriever] ========== 开始检索 ==========")
	config.Info("[Retriever] 查询文本: %s", truncateText(query, 100))
	config.Info("[Retriever] 查询长度: %d", len(query))

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// 步骤1: 使用嵌入模型将查询文本转换为向量
	config.Info("[Retriever] 步骤1/3: 向量化查询文本...")
	start := time.Now()
	vectors, err := r.eb.EmbedStrings(ctx, []string{query})
	elapsed := time.Since(start)
	config.Info("[Retriever] 向量化耗时: %v", elapsed)

	if err != nil {
		config.Error("[Retriever] 向量化失败: %v", err)
		config.Info("[Retriever] ========== 检索结束 (失败) ==========")
		return []*schema.Document{}, nil
	}

	if len(vectors) == 0 {
		config.Error("[Retriever] 向量化返回空向量")
		config.Info("[Retriever] ========== 检索结束 (失败) ==========")
		return []*schema.Document{}, nil
	}

	config.Info("[Retriever] 向量维度: %d", len(vectors[0]))

	// 步骤2: 将 float64 向量转换为 FloatVector
	config.Info("[Retriever] 步骤2/3: 转换向量类型为 FloatVector...")
	float32Vec := make(entity.FloatVector, len(vectors[0]))
	for i, val := range vectors[0] {
		float32Vec[i] = float32(val)
	}

	// 步骤3: 执行 Milvus 搜索
	config.Info("[Retriever] 步骤3/3: 执行 Milvus 搜索...")
	start = time.Now()

	// 构建搜索选项
	searchOption := milvusclient.NewSearchOption(r.collection, r.topK, []entity.Vector{float32Vec}).
		WithANNSField(r.vectorField).
		WithOutputFields(r.outputFields...)

	config.Info("[Retriever] 搜索参数:")
	config.Info("[Retriever]   - Collection: %s", r.collection)
	config.Info("[Retriever]   - TopK: %d", r.topK)
	config.Info("[Retriever]   - VectorField: %s", r.vectorField)
	config.Info("[Retriever]   - OutputFields: %v", r.outputFields)

	results, err := r.client.Search(ctx, searchOption)
	elapsed = time.Since(start)

	config.Info("[Retriever] 搜索耗时: %v", elapsed)

	if err != nil {
		config.Error("[Retriever] Milvus 搜索失败: %v", err)
		config.Info("[Retriever] ========== 检索结束 (失败) ==========")
		return []*schema.Document{}, nil
	}

	config.Info("[Retriever] 搜索返回结果数量: %d", len(results))

	if len(results) == 0 {
		config.Warn("[Retriever] 未检索到任何文档")
		config.Warn("[Retriever] 可能原因:")
		config.Warn("[Retriever]   1) 查询向量与数据库向量不匹配")
		config.Warn("[Retriever]   2) 索引未正确构建")
		config.Warn("[Retriever]   3) 数据库中没有数据")
		config.Info("[Retriever] ========== 检索结束 (空结果) ==========")
		return []*schema.Document{}, nil
	}

	// 解析搜索结果
	return parseSearchResults(results)
}

// parseSearchResults 解析 Milvus 搜索结果
// 注意：Milvus SDK v2.6.x 返回的是 ResultSet 类型
func parseSearchResults(results []milvusclient.ResultSet) ([]*schema.Document, error) {
	docs := make([]*schema.Document, 0)
	for _, result := range results {
		// 获取字段列
		idCol := result.GetColumn("id")
		contentCol := result.GetColumn("content")

		if idCol == nil {
			config.Warn("[Retriever] ID列为空，跳过")
			continue
		}

		// 获取行数
		rowCount := idCol.Len()
		config.Info("[Retriever] 搜索结果行数: %d", rowCount)

		// 逐行解析
		for i := 0; i < rowCount; i++ {
			doc := &schema.Document{}

			// 获取ID
			id, err := idCol.GetAsString(i)
			if err != nil {
				config.Warn("[Retriever] 获取第 %d 行ID失败: %v", i, err)
				continue
			}
			doc.ID = id

			// 获取Content
			if contentCol != nil {
				content, err := contentCol.GetAsString(i)
				if err == nil {
					doc.Content = content
				}
			}

			docs = append(docs, doc)
			config.Info("[Retriever] 添加文档 %d: ID=%s, ContentLen=%d", i+1, doc.ID, len(doc.Content))
		}
	}

	// 输出结果详情
	config.Info("[Retriever] 解析到 %d 个文档", len(docs))
	for i, doc := range docs {
		contentLen := len(doc.Content)
		config.Info("[Retriever] 文档 %d: ID=%s, ContentLength=%d", i+1, doc.ID, contentLen)
		if contentLen > 0 {
			preview := doc.Content
			if contentLen > 100 {
				preview = doc.Content[:100] + "..."
			}
			config.Info("[Retriever] 文档 %d 预览: %s", i+1, preview)
		} else {
			config.Warn("[Retriever] 文档 %d: 内容为空", i+1)
		}
	}

	config.Info("[Retriever] ========== 检索结束 (成功) ==========")
	return docs, nil
}

// truncateText 截断文本用于日志输出
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// initMilvusRetriever 初始化 Milvus 检索器
func initMilvusRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	config.Info("[Retriever] 初始化 Milvus retriever...")

	// 获取 Milvus 客户端
	milvusCli, err := milvus.GetMilvusClient(ctx)
	if err != nil {
		config.Error("[Retriever] 获取 Milvus 客户端失败: %v", err)
		config.Warn("[Retriever] 将使用空检索器")
		return &EmptyRetriever{}, nil
	}
	config.Info("[Retriever] Milvus 客户端获取成功")

	// 获取嵌入器
	eb, err := embedder.GetDoubaoEmbedding(ctx)
	if err != nil {
		config.Error("[Retriever] 获取嵌入器失败: %v", err)
		config.Warn("[Retriever] 将使用空检索器")
		return &EmptyRetriever{}, nil
	}
	config.Info("[Retriever] 嵌入器获取成功")

	// 创建自定义检索器
	config.Info("[Retriever] 创建自定义 retriever:")
	config.Info("[Retriever]   - Collection: %s", common.MilvusCollectionName)
	config.Info("[Retriever]   - TopK: %d", 3)
	config.Info("[Retriever]   - OutputFields: %v", []string{"id", "content", "metadata"})

	rtr = &CustomMilvusRetriever{
		client:       milvusCli,
		eb:           eb,
		collection:   common.MilvusCollectionName,
		vectorField:  "vector",
		outputFields: []string{"id", "content", "metadata"},
		topK:         3,
	}

	config.Info("[Retriever] 自定义 Milvus retriever 创建成功")
	return rtr, nil
}
