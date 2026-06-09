// Package chat_pipeline 鎻愪緵鑱婂ぉ绠￠亾鐨勬绱㈠櫒灏佽
package chat_pipeline

import (
	"context"

	retriever2 "SuperBizAgent/internal/ai/retriever"

	"github.com/cloudwego/eino/components/retriever"
)

// newRetriever 鍒涘缓妫€绱㈠櫒瀹炰緥
// 濮旀墭缁?internal/ai/retriever 鍖呯殑 GetMilvusRetriever
func newRetriever(ctx context.Context) (rtr retriever.Retriever, err error) {
	return retriever2.GetMilvusRetriever(ctx)
}