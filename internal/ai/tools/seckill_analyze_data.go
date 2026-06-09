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

// SeckillAnalyzeDataInput 秒杀数据分析的输入参数
type SeckillAnalyzeDataInput struct {
	AnalysisType string `json:"analysis_type" jsonschema:"description=分析类型，可选：hot_products（热点商品）、stock_alert（库存告警）、all（全部），默认all"`
}

// NewSeckillAnalyzeDataTool 创建秒杀数据分析工具
// 分析热点商品、库存告警、订单汇总等数据
func NewSeckillAnalyzeDataTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"seckill_analyze_data",
		"分析秒杀系统的数据。可以分析热点商品（按销量排序）、库存告警（库存低的商品）、订单汇总统计。帮助运营人员快速了解秒杀活动的情况。",
		func(ctx context.Context, input *SeckillAnalyzeDataInput, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-seckill_analyze_data] 工具调用开始")

			// 设置默认分析类型
			analysisType := input.AnalysisType
			if analysisType == "" {
				analysisType = "all"
			}

			// 构造查询条件
			query := &domain.AnalysisQuery{
				AnalysisType: analysisType,
			}

			config.Info("[Tool-seckill_analyze_data] 分析类型: %s", analysisType)

			// 调用 Service 层进行数据分析
			analysisService := service.NewAnalysisService()
			result := analysisService.AnalyzeData(ctx, query)

			// 转换为 JSON 格式返回
			jsonBytes, _ := json.MarshalIndent(result, "", "  ")

			config.Info("[Tool-seckill_analyze_data] 分析完成，返回 %d 字节数据", len(jsonBytes))
			return string(jsonBytes), nil
		})
	if err != nil {
		config.Error("[Tool-seckill_analyze_data] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}
	return t
}
