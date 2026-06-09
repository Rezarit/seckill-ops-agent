package models

import (
	"context"
	"fmt"
	"strings"

	"SuperBizAgent/config"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

// NewDoubaoOpenAIClient 使用 OpenAI 兼容协议连接火山 Ark，支持 Tool Calling 与流式输出。
func NewDoubaoOpenAIClient(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	cfg := config.AppConfig

	modelName := cfg.DoubaoChat.Model
	apiKey := cfg.DoubaoChat.APIKey
	baseURL := cfg.DoubaoChat.BaseURL

	if modelName == "" || apiKey == "" || baseURL == "" {
		config.Error("[Models] 豆包聊天模型配置不完整: Model=%s, APIKey=%s, BaseURL=%s", modelName, apiKey, baseURL)
		return nil, fmt.Errorf("豆包聊天模型配置不完整")
	}

	return newArkOpenAIChatModel(ctx, modelName, apiKey, normalizeArkBaseURL(baseURL))
}

func newArkOpenAIChatModel(ctx context.Context, modelName, apiKey, baseURL string) (model.ToolCallingChatModel, error) {
	innerModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:   modelName,
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
	if err != nil {
		return nil, err
	}

	config.Info("[Models] 豆包模型初始化成功: %s", modelName)
	return innerModel, nil
}

func normalizeArkBaseURL(base string) string {
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/api/v3") {
		return base
	}
	return base + "/api/v3"
}
