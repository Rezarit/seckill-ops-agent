package chat_pipeline

import (
	"SuperBizAgent/config"
	"context"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func BuildChatAgent(ctx context.Context) (r compose.Runnable[*UserMessage, *schema.Message], err error) {
	config.Info("[Orchestration] ========== 开始构建 Chat Agent ==========")

	const (
		ChatTemplate       = "ChatTemplate"
		ReactAgent         = "ReactAgent"
		CompleteRagProcess = "CompleteRagProcess"
	)
	g := compose.NewGraph[*UserMessage, *schema.Message]()

	// --- RAG 对话流程 ---
	config.Info("[Orchestration] 正在创建 MilvusRetriever...")
	milvusRetriever, err := newRetriever(ctx)
	if err != nil {
		config.Warn("[Orchestration] 创建 MilvusRetriever 失败: %v，将使用空 Retriever", err)
	} else {
		config.Info("[Orchestration] MilvusRetriever 创建完成")
	}

	config.Info("[Orchestration] 正在添加 CompleteRagProcess 节点...")
	completeRagLambda, err := createCompleteRagLambda(milvusRetriever)
	if err != nil {
		config.Error("[Orchestration] 创建 CompleteRagLambda 失败: %v", err)
		return nil, err
	}
	_ = g.AddLambdaNode(CompleteRagProcess, completeRagLambda, compose.WithNodeName("CompleteRagProcess"))

	config.Info("[Orchestration] 正在创建 ChatTemplate...")
	chatTemplateKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		config.Error("[Orchestration] 创建 ChatTemplate 失败: %v", err)
		return nil, err
	}
	_ = g.AddChatTemplateNode(ChatTemplate, chatTemplateKeyOfChatTemplate)

	config.Info("[Orchestration] 正在创建 ReActAgent...")
	reactAgentKeyOfLambda, err := newReactAgentLambda(ctx)
	if err != nil {
		config.Error("[Orchestration] 创建 ReActAgent 失败: %v", err)
		return nil, err
	}
	_ = g.AddLambdaNode(ReactAgent, reactAgentKeyOfLambda, compose.WithNodeName("ReActAgent"))

	// 构建边
	_ = g.AddEdge(compose.START, CompleteRagProcess)
	_ = g.AddEdge(CompleteRagProcess, ChatTemplate)
	_ = g.AddEdge(ChatTemplate, ReactAgent)
	_ = g.AddEdge(ReactAgent, compose.END)

	config.Info("[Orchestration] 正在编译工作流图...")
	r, err = g.Compile(ctx, compose.WithGraphName("ChatAgent"), compose.WithNodeTriggerMode(compose.AnyPredecessor))
	if err != nil {
		config.Error("[Orchestration] 编译图失败: %v", err)
		return nil, err
	}
	config.Info("[Orchestration] ========== Chat Agent 构建完成 ==========")
	return r, err
}
