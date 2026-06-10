package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var errorTypes = []struct {
	module   string
	message  string
	cause    string
	solution string
}{
	{"订单服务", "创建订单失败: 数据库连接池耗尽", "数据库连接池配置不足，最大连接数仅为20", "将数据库连接池最大连接数调整为100"},
	{"支付服务", "支付回调超时，订单状态不一致", "支付网关响应超时，未正确处理幂等性", "增加超时时间配置，实现幂等性校验"},
	{"Redis", "Redis连接超时，重试次数: 3", "Redis主从切换导致短暂不可用", "优化Redis连接重试策略，增加熔断机制"},
	{"库存服务", "库存扣减失败，商品ID: xxx", "库存数据不一致，并发扣减冲突", "增加分布式锁机制，确保库存一致性"},
	{"限流组件", "请求被限流，IP: xxx", "限流阈值设置过低，无法应对峰值流量", "动态调整限流阈值，支持热点IP豁免"},
	{"消息队列", "消息发送失败，topic: xxx", "MQ集群故障，消息堆积", "切换备用MQ集群，增加消息重试机制"},
	{"用户服务", "用户认证失败，token过期", "token有效期过短，用户频繁重新登录", "延长token有效期至2小时，支持refresh token"},
	{"缓存服务", "缓存击穿，查询DB超时", "热点key未缓存，大量请求直接访问DB", "增加热点key预缓存机制，设置合理过期时间"},
	{"配置中心", "配置拉取失败", "配置中心不可用，服务无法启动", "增加本地配置缓存，支持配置热更新"},
	{"监控服务", "指标上报失败", "网络波动导致指标丢失", "增加本地缓存和批量上报机制"},
}

func main() {
	dir := "internal/ai/cmd/knowledge_cmd/docs/error_logs"
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	for i := 101; i <= 1000; i++ {
		idx := (i - 1) % len(errorTypes)
		et := errorTypes[idx]

		content := fmt.Sprintf(`# 【%s】错误报告

## 问题描述
第%d次模拟：%s

## 错误日志
ERROR [2026-06-11 14:%02d:%02d] %s.go:%d - %s


## 根因分析
%s

## 解决方案
%s
`, et.module, i, et.message, i/60, i%60, et.module, 100+i, et.message, et.cause, et.solution)

		filename := filepath.Join(dir, fmt.Sprintf("error_log_%03d.md", i))
		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			fmt.Printf("写入文件 %s 失败: %v\n", filename, err)
			return
		}

		if i%10 == 0 {
			fmt.Printf("已生成 %d 个文档...\n", i)
		}
	}

	fmt.Println("✅ 100个错误日志文档生成完成！")
}
