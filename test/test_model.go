package main

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/models"
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func main() {
	fmt.Println("开始测试模型...")

	if err := config.InitConfig(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		return
	}
	fmt.Println("配置初始化成功")

	ctx := context.Background()

	fmt.Println("创建模型客户端...")
	model, err := models.NewDoubaoOpenAIClient(ctx)
	if err != nil {
		fmt.Printf("创建模型客户端失败: %v\n", err)
		return
	}
	fmt.Println("模型客户端创建成功")

	fmt.Println("发送消息...")
	msgs := []*schema.Message{
		schema.UserMessage("你好"),
	}
	out, err := model.Generate(ctx, msgs)
	if err != nil {
		fmt.Printf("调用模型失败: %v\n", err)
		return
	}

	fmt.Printf("模型回答: %s\n", out.Content)
}
