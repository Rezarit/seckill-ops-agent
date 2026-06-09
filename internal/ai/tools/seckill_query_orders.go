// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"

	"SuperBizAgent/internal/seckill/domain"
	"SuperBizAgent/internal/seckill/service"
)

// SeckillQueryOrdersInput 查询秒杀订单的输入参数
type SeckillQueryOrdersInput struct {
	OrderID *int64  `json:"order_id,omitempty" jsonschema:"description=订单ID，可选，不传则查询所有订单统计"`
	UserID  *int64  `json:"user_id,omitempty" jsonschema:"description=用户ID，可选，查询指定用户的订单"`
	Status  *string `json:"status,omitempty" jsonschema:"description=订单状态，可选，如pending、processing、paid、shipping、completed、cancelled"`
	Limit   *int    `json:"limit,omitempty" jsonschema:"description=返回数量限制，默认100"`
	Offset  *int    `json:"offset,omitempty" jsonschema:"description=偏移量，默认0"`
}

// NewSeckillQueryOrdersTool 创建查询秒杀订单工具
// 查询单个订单详情、用户订单列表或所有订单统计
func NewSeckillQueryOrdersTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"seckill_query_orders",
		"查询秒杀系统的订单信息。可以查询单个订单详情、用户订单列表、所有订单统计。返回订单的基本信息、金额、状态等，以及统计数据。",
		func(ctx context.Context, input *SeckillQueryOrdersInput, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-seckill_query_orders] 工具调用开始")

			// 构造查询条件
			query := &domain.OrderQuery{
				OrderID: input.OrderID,
				UserID:  input.UserID,
				Status:  input.Status,
				Limit:   input.Limit,
				Offset:  input.Offset,
			}

			config.Info("[Tool-seckill_query_orders] 查询条件: OrderID=%v, UserID=%v, Status=%v, Limit=%v, Offset=%v",
				input.OrderID, input.UserID, input.Status, input.Limit, input.Offset)

			// 调用 Service 层查询订单
			orderService := service.NewOrderService()
			result := orderService.QueryOrders(ctx, query)

			// 转换为 JSON 格式返回
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")

			config.Info("[Tool-seckill_query_orders] 查询完成，返回 %d 字节数据", len(jsonBytes))
			return string(jsonBytes), nil
		})
	if err != nil {
		config.Error("[Tool-seckill_query_orders] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}
	return t
}
