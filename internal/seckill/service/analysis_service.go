// Package service 秒杀系统业务逻辑层
// 提供数据分析、热点商品、库存告警等业务功能
package service

import (
	"context"

	"SuperBizAgent/config"
	"SuperBizAgent/internal/seckill/dao"
	"SuperBizAgent/internal/seckill/domain"
)

// AnalysisService 数据分析服务
// 提供热点商品分析、库存告警、订单汇总等功能
type AnalysisService struct {
	analysisDAO *dao.AnalysisDAO // 数据分析访问对象
}

// NewAnalysisService 创建数据分析服务实例
func NewAnalysisService() *AnalysisService {
	return &AnalysisService{
		analysisDAO: dao.NewAnalysisDAO(),
	}
}

// AnalyzeData 执行数据分析
// 参数:
//   - ctx: 请求上下文
//   - query: 分析查询条件（包含分析类型）
// 返回: 分析结果（热点商品、库存告警、订单汇总）
func (s *AnalysisService) AnalyzeData(ctx context.Context, query *domain.AnalysisQuery) *domain.AnalysisResult {
	config.Info("[Seckill-AnalysisService] 执行数据分析 | Type: %s", query.AnalysisType)

	// 默认分析类型为全部
	analysisType := query.AnalysisType
	if analysisType == "" {
		analysisType = domain.AnalysisTypeAll
		config.Info("[Seckill-AnalysisService] 使用默认分析类型: %s", analysisType)
	}

	result := &domain.AnalysisResult{
		Success: true,
		Message: "分析成功",
	}

	// 获取热点商品
	if analysisType == domain.AnalysisTypeHotProducts || analysisType == domain.AnalysisTypeAll {
		config.Info("[Seckill-AnalysisService] 获取热点商品...")
		hotProducts, err := s.analysisDAO.GetHotProducts(ctx)
		if err != nil {
			config.Error("[Seckill-AnalysisService] 获取热点商品失败: %v", err)
		} else {
			result.HotProducts = hotProducts
			config.Info("[Seckill-AnalysisService] 获取热点商品成功 | 数量: %d", len(hotProducts))
		}
	}

	// 获取库存告警
	if analysisType == domain.AnalysisTypeStockAlert || analysisType == domain.AnalysisTypeAll {
		config.Info("[Seckill-AnalysisService] 获取库存告警...")
		stockAlerts, err := s.analysisDAO.GetStockAlerts(ctx)
		if err != nil {
			config.Error("[Seckill-AnalysisService] 获取库存告警失败: %v", err)
		} else {
			result.StockAlerts = stockAlerts
			config.Info("[Seckill-AnalysisService] 获取库存告警成功 | 数量: %d", len(stockAlerts))
		}
	}

	// 获取订单汇总（仅全部分析时）
	if analysisType == domain.AnalysisTypeAll {
		config.Info("[Seckill-AnalysisService] 获取订单汇总...")
		orderSummary, err := s.analysisDAO.GetOrderSummary(ctx)
		if err != nil {
			config.Error("[Seckill-AnalysisService] 获取订单汇总失败: %v", err)
		} else {
			result.OrderSummary = orderSummary
			config.Info("[Seckill-AnalysisService] 获取订单汇总成功")
		}
	}

	config.Info("[Seckill-AnalysisService] 数据分析完成")
	return result
}
