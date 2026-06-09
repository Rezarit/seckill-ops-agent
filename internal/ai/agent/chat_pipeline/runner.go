package chat_pipeline

import (
	"SuperBizAgent/config"
	"context"
	"sync"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// runnerCache 缓存编译好的 Chat Agent Runner
var (
	runner     compose.Runnable[*UserMessage, *schema.Message]
	runnerOnce sync.Once
	runnerErr  error
	isInitDone bool
)

// InitChatRunner 主动初始化聊天代理（在服务启动时调用）
// 如果核心组件初始化失败，会返回错误
// 如果只是增强组件失败，会降级处理并继续
func InitChatRunner(ctx context.Context) error {
	config.Info("[Runner] ========== 开始初始化 Chat Agent ==========")

	var err error
	runner, err = BuildChatAgent(ctx)
	if err != nil {
		config.Error("[Runner] ========== Chat Agent 初始化失败 ==========")
		config.Error("[Runner] 错误原因: %v", err)
		return err
	}

	isInitDone = true
	config.Info("[Runner] ========== Chat Agent 初始化成功 ==========")
	return nil
}

// GetChatRunner 获取缓存的 Chat Agent Runner
// 如果已经通过 InitChatRunner 初始化，直接返回缓存的实例
// 如果未初始化，则尝试初始化（兼容旧代码）
func GetChatRunner(ctx context.Context) (compose.Runnable[*UserMessage, *schema.Message], error) {
	config.Info("[Runner] Getting chat runner...")

	if isInitDone {
		config.Info("[Runner] 使用已初始化的 runner")
		return runner, nil
	}

	// 兼容旧代码：如果未提前初始化，在第一次调用时初始化
	runnerOnce.Do(func() {
		config.Warn("[Runner] 未提前初始化，正在延迟初始化...")
		runner, runnerErr = BuildChatAgent(ctx)
		if runnerErr != nil {
			config.Error("[Runner] 延迟初始化失败: %v", runnerErr)
		} else {
			isInitDone = true
			config.Info("[Runner] 延迟初始化成功")
		}
	})

	if runnerErr != nil {
		return nil, runnerErr
	}

	return runner, nil
}
