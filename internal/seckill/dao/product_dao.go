// Package dao
package dao

import (
	"context"

	"SuperBizAgent/internal/seckill/config"
	"SuperBizAgent/internal/seckill/domain"
)

// ProductDAO 鍟嗗搧鏁版嵁璁块棶瀵硅薄
type ProductDAO struct{}

// NewProductDAO 鍒涘缓鍟嗗搧 DAO
func NewProductDAO() *ProductDAO {
	return &ProductDAO{}
}

// Query 鏌ヨ鍟嗗搧
func (dao *ProductDAO) Query(ctx context.Context, query *domain.ProductQuery) ([]*domain.Product, error) {
	db, err := config.GetDB()
	if err != nil {
		return nil, err
	}

	var products []*domain.Product
	dbQuery := db.Model(&domain.Product{})

	if query.ProductID != nil {
		dbQuery = dbQuery.Where("product_id = ?", *query.ProductID)
	}

	if query.Keyword != nil && *query.Keyword != "" {
		dbQuery = dbQuery.Where("product_name LIKE ?", "%"+*query.Keyword+"%")
	}

	err = dbQuery.Find(&products).Error
	return products, err
}
