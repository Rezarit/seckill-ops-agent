// Package dao 秒杀系统数据访问层
// 提供数据分析相关的数据库操作
package dao

import (
	"context"

	"SuperBizAgent/internal/seckill/config"
	"SuperBizAgent/internal/seckill/domain"
)

// AnalysisDAO 数据分析访问对象
// 提供热点商品、库存告警、订单汇总等数据分析功能
type AnalysisDAO struct{}

// NewAnalysisDAO 创建数据分析 DAO 实例
func NewAnalysisDAO() *AnalysisDAO {
	return &AnalysisDAO{}
}

// GetHotProducts 获取热点商品列表
// 根据订单销量排序，返回前10个商品
// 参数:
//   - ctx: 请求上下文
// 返回: 热点商品列表和错误信息
func (dao *AnalysisDAO) GetHotProducts(ctx context.Context) ([]domain.HotProductItem, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var hotProducts []domain.HotProductItem
	err = db.Raw(`
		SELECT 
			p.product_id,
			p.product_name,
			COUNT(DISTINCT oi.order_id) as order_count,
			SUM(oi.quantity) as total_sold,
			SUM(oi.quantity * oi.price) as total_amount,
			p.stock
		FROM order_items oi
		JOIN products p ON oi.product_id = p.product_id
		GROUP BY p.product_id
		ORDER BY total_sold DESC
		LIMIT 10
	`).Scan(&hotProducts).Error

	return hotProducts, err
}

// GetStockAlerts 获取库存告警列表
// 返回库存低于10的商品，标记告警级别
// 参数:
//   - ctx: 请求上下文
// 返回: 库存告警列表和错误信息
func (dao *AnalysisDAO) GetStockAlerts(ctx context.Context) ([]domain.StockAlertItem, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var stockAlerts []domain.StockAlertItem
	err = db.Raw(`
		SELECT 
			product_id,
			product_name,
			stock,
			CASE 
				WHEN stock = 0 THEN 'critical'
				WHEN stock < 10 THEN 'low'
				ELSE 'normal'
			END as alert_level
		FROM products
		WHERE stock < 10
		ORDER BY stock ASC
	`).Scan(&stockAlerts).Error

	return stockAlerts, err
}

// GetOrderSummary 获取订单汇总数据
// 包含订单总数、总金额、完成率、商品总数、库存告警等统计
// 参数:
//   - ctx: 请求上下文
// 返回: 订单汇总数据和错误信息
func (dao *AnalysisDAO) GetOrderSummary(ctx context.Context) (*domain.OrderSummary, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var totalOrders int64
	var totalAmount float64
	var totalProducts int64
	var completedOrders int64
	var lowStockCount int64
	var zeroStockCount int64

	// 查询订单总数
	db.Model(&domain.Order{}).Count(&totalOrders)
	// 查询订单总金额（COALESCE处理空表情况）
	db.Model(&domain.Order{}).Select("COALESCE(SUM(total), 0)").Scan(&totalAmount)
	// 查询已完成订单数
	db.Model(&domain.Order{}).Where("status = ?", "completed").Count(&completedOrders)
	// 查询商品总数
	db.Model(&domain.Product{}).Count(&totalProducts)
	// 查询低库存商品数（库存大于0小于10）
	db.Model(&domain.Product{}).Where("stock < 10 AND stock > 0").Count(&lowStockCount)
	// 查询零库存商品数
	db.Model(&domain.Product{}).Where("stock = 0").Count(&zeroStockCount)

	// 计算订单完成率
	successRate := 0.0
	if totalOrders > 0 {
		successRate = float64(completedOrders) / float64(totalOrders) * 100
	}

	return &domain.OrderSummary{
		TotalOrders:    int(totalOrders),
		TotalAmount:    totalAmount,
		SuccessRate:    successRate,
		TotalProducts:  int(totalProducts),
		LowStockCount:  int(lowStockCount),
		ZeroStockCount: int(zeroStockCount),
	}, nil
}
