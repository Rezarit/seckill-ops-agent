package domain

import "time"

// Order 订单模型
type Order struct {
	OrderID   int64     `json:"order_id" gorm:"column:order_id;primaryKey;autoIncrement"`
	UserID    int64     `json:"user_id" gorm:"column:user_id;index"`
	Address   string    `json:"address" gorm:"column:address;not null"`
	Total     float64   `json:"total" gorm:"column:total;type:decimal(10,2);default:0"`
	Status    string    `json:"status" gorm:"column:status;default:'pending'"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// OrderItem 订单项模型
type OrderItem struct {
	OrderItemID int64   `json:"order_item_id" gorm:"column:order_item_id;primaryKey;autoIncrement"`
	OrderID     int64   `json:"order_id" gorm:"column:order_id;index"`
	ProductID   int64   `json:"product_id" gorm:"column:product_id"`
	ProductName string  `json:"product_name" gorm:"column:product_name;not null"`
	Quantity    int     `json:"quantity" gorm:"column:quantity;not null"`
	Price       float64 `json:"price" gorm:"column:price;type:decimal(10,2);not null"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// OrderQuery 订单查询条件
type OrderQuery struct {
	OrderID *int64  // 订单ID
	UserID  *int64  // 用户ID
	Status  *string // 订单状态
	Limit   *int    // 数量限制
	Offset  *int    // 偏移量
}

// OrderStats 订单统计
type OrderStats struct {
	TotalOrders     int     `json:"total_orders"`
	TotalAmount     float64 `json:"total_amount"`
	PendingCount    int     `json:"pending_count"`
	ProcessingCount int     `json:"processing_count"`
	PaidCount       int     `json:"paid_count"`
	CompletedCount  int     `json:"completed_count"`
	CancelledCount  int     `json:"cancelled_count"`
}

// OrderQueryResult 订单查询结果
type OrderQueryResult struct {
	Success bool        `json:"success" jsonschema:"description=查询是否成功"`
	Count   int         `json:"count" jsonschema:"description=订单数量"`
	Data    []*Order    `json:"data,omitempty" jsonschema:"description=订单列表"`
	Stats   *OrderStats `json:"stats,omitempty" jsonschema:"description=订单统计信息"`
	Message string      `json:"message,omitempty" jsonschema:"description=状态消息"`
}
