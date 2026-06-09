package embedder

import (
	"SuperBizAgent/config"
	"context"
	"sync"

	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
)

var (
	embedderInstance embedding.Embedder
	embedderOnce     sync.Once
	embedderInitErr  error
)

// GetDoubaoEmbedding 获取全局单例嵌入器，服务启动时初始化一次，后续复用
func GetDoubaoEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	embedderOnce.Do(func() {
		config.Info("[Embedder] ========== 首次初始化嵌入器（全局单例）==========")
		embedderInstance, embedderInitErr = initDoubaoEmbedding(ctx)
		if embedderInitErr != nil {
			config.Error("[Embedder] 全局嵌入器初始化失败: %v", embedderInitErr)
		} else {
			config.Info("[Embedder] 全局嵌入器初始化成功")
		}
	})
	return embedderInstance, embedderInitErr
}

// DoubaoEmbedding 保留旧接口以兼容外部调用，内部委托给 GetDoubaoEmbedding
func DoubaoEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	return GetDoubaoEmbedding(ctx)
}

func initDoubaoEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	config.Info("[Embedder] 开始初始化阿里百炼嵌入器...")

	model := config.AppConfig.DoubaoEmb.Model
	apiKey := config.AppConfig.DoubaoEmb.APIKey

	config.Info("[Embedder] 配置: model=%s", model)

	eb, err = dashscope.NewEmbedder(ctx, &dashscope.EmbeddingConfig{
		Model:  model,
		APIKey: apiKey,
	})
	if err != nil {
		config.Error("[Embedder] 初始化失败: %v", err)
		return nil, err
	}

	config.Info("[Embedder] 初始化成功")
	return eb, nil
}
