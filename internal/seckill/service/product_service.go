// Package service 绉掓潃绯荤粺涓氬姟閫昏緫灞?// 鍖呭惈鍟嗗搧鏈嶅姟銆佽鍗曟湇鍔°€佹暟鎹垎鏋愭湇鍔＄瓑鏍稿績涓氬姟
package service

import (
	"context"

	"SuperBizAgent/config"
	"SuperBizAgent/internal/seckill/dao"
	"SuperBizAgent/internal/seckill/domain"
)

// ProductService 鍟嗗搧鏈嶅姟
// 鎻愪緵鍟嗗搧鏌ヨ銆佸簱瀛樼鐞嗙瓑鍔熻兘
type ProductService struct {
	productDAO *dao.ProductDAO // 鍟嗗搧鏁版嵁璁块棶瀵硅薄
}

// NewProductService 鍒涘缓鍟嗗搧鏈嶅姟瀹炰緥
func NewProductService() *ProductService {
	return &ProductService{
		productDAO: dao.NewProductDAO(),
	}
}

// QueryProducts 鏌ヨ鍟嗗搧鍒楄〃
// 鍙傛暟:
//   - ctx: 璇锋眰涓婁笅鏂?//   - query: 鏌ヨ鏉′欢锛堟敮鎸佹寜鍟嗗搧ID銆佸叧閿瘝鎼滅储锛?// 杩斿洖: 鏌ヨ缁撴灉锛堝寘鍚晢鍝佸垪琛ㄥ拰鐘舵€佷俊鎭級
func (s *ProductService) QueryProducts(ctx context.Context, query *domain.ProductQuery) *domain.ProductQueryResult {
	config.Info("[Seckill-ProductService] 鏌ヨ鍟嗗搧 | ProductID: %v, Keyword: %v", query.ProductID, query.Keyword)

	// 璋冪敤 DAO 鏌ヨ鍟嗗搧
	products, err := s.productDAO.Query(ctx, query)
	if err != nil {
		config.Error("[Seckill-ProductService] 鏌ヨ鍟嗗搧澶辫触: %v", err)
		return &domain.ProductQueryResult{
			Success: false,
			Message: "鏌ヨ鍟嗗搧澶辫触",
		}
	}

	config.Info("[Seckill-ProductService] 鏌ヨ鍟嗗搧鎴愬姛 | 鏁伴噺: %d", len(products))
	return &domain.ProductQueryResult{
		Success: true,
		Count:   len(products),
		Data:    products,
		Message: "鏌ヨ鎴愬姛",
	}
}
