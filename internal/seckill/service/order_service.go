// Package service 秒杀系统业务逻辑层
// 提供订单查询、统计分析等业务功能
package service

import (
	"context"

	"SuperBizAgent/config"
	"SuperBizAgent/internal/seckill/dao"
	"SuperBizAgent/internal/seckill/domain"
)

// OrderService 订单服务
// 提供订单查询、统计等功能
type OrderService struct {
	orderDAO *dao.OrderDAO // 订单数据访问对象
}

// NewOrderService 创建订单服务实例
func NewOrderService() *OrderService {
	return &OrderService{
		orderDAO: dao.NewOrderDAO(),
	}
}

// QueryOrders 查询订单列表
// 参数:
//   - ctx: 请求上下文
//   - query: 查询条件（支持按订单ID、用户ID、状态筛选）
// 返回: 查询结果（包含订单列表和统计信息）
func (s *OrderService) QueryOrders(ctx context.Context, query *domain.OrderQuery) *domain.OrderQueryResult {
	config.Info("[Seckill-OrderService] 查询订单 | OrderID: %v, UserID: %v, Status: %v", query.OrderID, query.UserID, query.Status)

	// 调用 DAO 查询订单
	orders, err := s.orderDAO.Query(ctx, query)
	if err != nil {
		config.Error("[Seckill-OrderService] 查询订单失败: %v", err)
		return &domain.OrderQueryResult{
			Success: false,
			Message: "查询订单失败",
		}
	}

	// 如果不是按订单ID查询，获取统计信息
	var stats *domain.OrderStats
	if query.OrderID == nil {
		stats, err = s.orderDAO.GetStats(ctx)
		if err != nil {
			config.Warn("[Seckill-OrderService] 获取订单统计失败: %v", err)
		}
	}

	config.Info("[Seckill-OrderService] 查询订单成功 | 数量: %d", len(orders))
	return &domain.OrderQueryResult{
		Success: true,
		Count:   len(orders),
		Data:    orders,
		Stats:   stats,
		Message: "查询成功",
	}
}
