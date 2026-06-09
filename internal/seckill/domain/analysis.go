package domain

// AnalysisType 分析类型常量
const (
	AnalysisTypeHotProducts = "hot_products"
	AnalysisTypeStockAlert  = "stock_alert"
	AnalysisTypeAll         = "all"
)

// AnalysisQuery 数据分析查询条件
type AnalysisQuery struct {
	AnalysisType string // 分析类型
}

// HotProductItem 热点商品
type HotProductItem struct {
	ProductID   int64   `json:"product_id" jsonschema:"description=商品ID"`
	ProductName string  `json:"product_name" jsonschema:"description=商品名称"`
	OrderCount  int     `json:"order_count" jsonschema:"description=订单数量"`
	TotalSold   int     `json:"total_sold" jsonschema:"description=总销量"`
	TotalAmount float64 `json:"total_amount" jsonschema:"description=总销售额"`
	Stock       int     `json:"stock" jsonschema:"description=当前库存"`
}

// StockAlertItem 库存告警
type StockAlertItem struct {
	ProductID   int64  `json:"product_id" jsonschema:"description=商品ID"`
	ProductName string `json:"product_name" jsonschema:"description=商品名称"`
	Stock       int    `json:"stock" jsonschema:"description=当前库存"`
	AlertLevel  string `json:"alert_level" jsonschema:"description=告警级别：low（库存低）、critical（库存告急）"`
}

// OrderSummary 订单汇总
type OrderSummary struct {
	TotalOrders    int     `json:"total_orders" jsonschema:"description=总订单数"`
	TotalAmount    float64 `json:"total_amount" jsonschema:"description=总金额"`
	SuccessRate    float64 `json:"success_rate" jsonschema:"description=成功率（已完成/总数）"`
	TotalProducts  int     `json:"total_products" jsonschema:"description=商品总数"`
	LowStockCount  int     `json:"low_stock_count" jsonschema:"description=低库存商品数（库存<10）"`
	ZeroStockCount int     `json:"zero_stock_count" jsonschema:"description=零库存商品数"`
}

// AnalysisResult 数据分析结果
type AnalysisResult struct {
	Success      bool             `json:"success" jsonschema:"description=分析是否成功"`
	HotProducts  []HotProductItem `json:"hot_products,omitempty" jsonschema:"description=热点商品列表"`
	StockAlerts  []StockAlertItem `json:"stock_alerts,omitempty" jsonschema:"description=库存告警列表"`
	OrderSummary *OrderSummary    `json:"order_summary,omitempty" jsonschema:"description=订单汇总"`
	Message      string           `json:"message,omitempty" jsonschema:"description=状态消息"`
}
