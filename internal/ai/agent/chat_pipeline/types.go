// Package chat_pipeline
package chat_pipeline

import "github.com/cloudwego/eino/schema"

// UserMessage 鐢ㄦ埛娑堟伅缁撴瀯浣?// 鍖呭惈鐢ㄦ埛鏌ヨ鐨勫熀鏈俊鎭拰瀵硅瘽鍘嗗彶
type UserMessage struct {
	ID      string            `json:"id"`      // 鐢ㄦ埛浼氳瘽鍞竴鏍囪瘑
	Query   string            `json:"query"`   // 鐢ㄦ埛褰撳墠鎻愰棶鍐呭
	History []*schema.Message `json:"history"` // 瀵硅瘽鍘嗗彶璁板綍
}