// Package dao 秒杀系统数据访问层
package dao

import (
	"context"

	"SuperBizAgent/internal/seckill/config"
	"SuperBizAgent/internal/seckill/domain"
)

// OrderDAO 订单数据访问对象
// 提供订单查询和统计功能
type OrderDAO struct{}

// NewOrderDAO 创建订单 DAO 实例
func NewOrderDAO() *OrderDAO {
	return &OrderDAO{}
}

// Query 查询订单列表
// 参数:
//   - ctx: 请求上下文
//   - query: 查询条件（支持订单ID、用户ID、状态筛选）
// 返回: 订单列表和错误信息
func (dao *OrderDAO) Query(ctx context.Context, query *domain.OrderQuery) ([]*domain.Order, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var orders []*domain.Order
	dbQuery := db.Model(&domain.Order{})

	// 按订单ID筛选
	if query.OrderID != nil {
		dbQuery = dbQuery.Where("order_id = ?", *query.OrderID)
	}

	// 按用户ID筛选
	if query.UserID != nil {
		dbQuery = dbQuery.Where("user_id = ?", *query.UserID)
	}

	// 按状态筛选
	if query.Status != nil && *query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", *query.Status)
	}

	// 设置分页参数
	limit := 100
	if query.Limit != nil {
		limit = *query.Limit
	}
	offset := 0
	if query.Offset != nil {
		offset = *query.Offset
	}

	// 执行查询，按创建时间倒序排列
	err = dbQuery.Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders).Error
	return orders, err
}

// GetStats 获取订单统计信息
// 参数:
//   - ctx: 请求上下文
// 返回: 订单统计数据和错误信息
func (dao *OrderDAO) GetStats(ctx context.Context) (*domain.OrderStats, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var totalOrders int64
	var totalAmount float64
	var pendingCount int64
	var processingCount int64
	var paidCount int64
	var completedCount int64
	var cancelledCount int64

	// 查询订单总数
	db.Model(&domain.Order{}).Count(&totalOrders)
	// 查询订单总金额
	db.Model(&domain.Order{}).Select("COALESCE(SUM(total), 0)").Scan(&totalAmount)
	// 按状态统计订单数
	db.Model(&domain.Order{}).Where("status = ?", "pending").Count(&pendingCount)
	db.Model(&domain.Order{}).Where("status = ?", "processing").Count(&processingCount)
	db.Model(&domain.Order{}).Where("status = ?", "paid").Count(&paidCount)
	db.Model(&domain.Order{}).Where("status = ?", "completed").Count(&completedCount)
	db.Model(&domain.Order{}).Where("status = ?", "cancelled").Count(&cancelledCount)

	return &domain.OrderStats{
		TotalOrders:     int(totalOrders),
		TotalAmount:     totalAmount,
		PendingCount:    int(pendingCount),
		ProcessingCount: int(processingCount),
		PaidCount:       int(paidCount),
		CompletedCount:  int(completedCount),
		CancelledCount:  int(cancelledCount),
	}, nil
}
