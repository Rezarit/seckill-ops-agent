package main

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/agent/chat_pipeline"
	"SuperBizAgent/utility/mem"
	"context"
	"fmt"

	"github.com/cloudwego/eino/schema"
)

func main() {
	fmt.Println("=== 开始聊天测试 ===")

	// 初始化配置
	fmt.Println("1. 初始化配置...")
	if err := config.InitConfig(); err != nil {
		fmt.Printf("配置初始化失败: %v\n", err)
		return
	}
	fmt.Println("配置初始化完成")

	ctx := context.Background()
	id := "111"
	userMessage := &chat_pipeline.UserMessage{
		ID:      id,
		Query:   "你好",
		History: mem.GetSimpleMemory(id).GetMessages(),
	}
	fmt.Println("2. 构建聊天代理...")
	runner, err := chat_pipeline.BuildChatAgent(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("聊天代理构建完成")

	// 第一次对话
	fmt.Println("3. 第一次对话: 你好")
	out, err := runner.Invoke(ctx, userMessage)
	if err != nil {
		panic(err)
	}
	answer := out.Content
	fmt.Println("Q: 你好")
	fmt.Println("A:", answer)
	mem.GetSimpleMemory(id).SetMessages(schema.UserMessage("你好"))
	mem.GetSimpleMemory(id).SetMessages(schema.SystemMessage(out.Content))

	// 第二次对话
	fmt.Println("4. 第二次对话: 现在是几点")
	userMessage = &chat_pipeline.UserMessage{
		ID:      id,
		Query:   "现在是几点",
		History: mem.GetSimpleMemory(id).GetMessages(),
	}
	out, err = runner.Invoke(ctx, userMessage)
	if err != nil {
		panic(err)
	}
	answer = out.Content
	fmt.Println("----------------")
	fmt.Println("Q: 现在是几点")
	fmt.Println("A:", answer)

	fmt.Println("=== 聊天测试结束 ===")
}
