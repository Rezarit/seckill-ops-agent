package models

import (
	"SuperBizAgent/config"
	"context"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
)

func OpenAIForDeepSeekV31Think(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	config.Info("[Model] 开始初始化 DeepSeek Think 模型...")

	model := config.AppConfig.DSTHINK.Model
	apiKey := config.AppConfig.DSTHINK.APIKey
	baseURL := config.AppConfig.DSTHINK.BaseURL

	config.Info("[Model] DeepSeek Think 配置: model=%s, base_url=%s", model, baseURL)

	chatConfig := &openai.ChatModelConfig{
		Model:   model,
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
	cm, err = openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		config.Error("[Model] DeepSeek Think 初始化失败: %v", err)
		return nil, err
	}

	config.Info("[Model] DeepSeek Think 初始化成功")
	return cm, nil
}

func OpenAIForDeepSeekV3Quick(ctx context.Context) (cm model.ToolCallingChatModel, err error) {
	config.Info("[Model] 开始初始化 DeepSeek Quick 模型...")

	model := config.AppConfig.DSQUICK.Model
	apiKey := config.AppConfig.DSQUICK.APIKey
	baseURL := config.AppConfig.DSQUICK.BaseURL

	config.Info("[Model] DeepSeek Quick 配置: model=%s, base_url=%s", model, baseURL)

	chatConfig := &openai.ChatModelConfig{
		Model:   model,
		APIKey:  apiKey,
		BaseURL: baseURL,
	}
	cm, err = openai.NewChatModel(ctx, chatConfig)
	if err != nil {
		config.Error("[Model] DeepSeek Quick 初始化失败: %v", err)
		return nil, err
	}

	config.Info("[Model] DeepSeek Quick 初始化成功")
	return cm, nil
}
