// Package domain 秒杀系统领域模型
package domain

import "time"

// Product 商品模型
type Product struct {
	ProductID   int64     `json:"product_id" gorm:"column:product_id;primaryKey;autoIncrement"`
	MerchantID  int64     `json:"merchant_id" gorm:"column:merchant_id;not null"`
	ProductName string    `json:"product_name" gorm:"column:product_name;not null"`
	Description string    `json:"description" gorm:"column:description;not null"`
	CommentNum  int       `json:"comment_num" gorm:"column:comment_num;default:0"`
	Price       float64   `json:"price" gorm:"column:price;type:decimal(10,2);default:0"`
	Stock       int       `json:"stock" gorm:"column:stock;default:0"`
	Cover       string    `json:"cover" gorm:"column:cover;not null"`
	PublishTime time.Time `json:"publish_time" gorm:"column:publish_time;autoCreateTime"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// ProductQuery 商品查询条件
type ProductQuery struct {
	ProductID *int64  // 商品ID
	Keyword   *string // 搜索关键词
}

// ProductQueryResult 商品查询结果
type ProductQueryResult struct {
	Success bool       `json:"success" jsonschema:"description=查询是否成功"`
	Count   int        `json:"count" jsonschema:"description=商品数量"`
	Data    []*Product `json:"data,omitempty" jsonschema:"description=商品列表"`
	Message string     `json:"message,omitempty" jsonschema:"description=状态消息"`
}
