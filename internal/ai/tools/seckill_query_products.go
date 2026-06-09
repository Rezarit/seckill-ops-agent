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

// SeckillQueryProductsInput 查询秒杀商品的输入参数
type SeckillQueryProductsInput struct {
	ProductID *int64  `json:"product_id,omitempty" jsonschema:"description=商品ID，可选，不传则查询所有商品"`
	Keyword   *string `json:"keyword,omitempty" jsonschema:"description=搜索关键词，可选，用于搜索商品名称"`
}

// NewSeckillQueryProductsTool 创建查询秒杀商品工具
// 根据商品ID查询单个商品、根据关键词搜索商品或查询所有商品列表
func NewSeckillQueryProductsTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"seckill_query_products",
		"查询秒杀系统的商品信息。可以根据商品ID查询单个商品，或根据关键词搜索商品，或查询所有商品列表。返回商品的基本信息、库存、价格等。",
		func(ctx context.Context, input *SeckillQueryProductsInput, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-seckill_query_products] 工具调用开始")

			// 构造查询条件
			query := &domain.ProductQuery{
				ProductID: input.ProductID,
				Keyword:   input.Keyword,
			}

			config.Info("[Tool-seckill_query_products] 查询条件: ProductID=%v, Keyword=%v", input.ProductID, input.Keyword)

			// 调用 Service 层查询商品
			productService := service.NewProductService()
			result := productService.QueryProducts(ctx, query)

			// 转换为 JSON 格式返回
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")

			config.Info("[Tool-seckill_query_products] 查询完成，返回 %d 字节数据", len(jsonBytes))
			return string(jsonBytes), nil
		})
	if err != nil {
		config.Error("[Tool-seckill_query_products] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}
	return t
}
