// Package chat_pipeline 鎻愪緵宸ュ叿鑺傜偣鐨勫垵濮嬪寲灏佽
package chat_pipeline

import (
	"SuperBizAgent/config"
	"context"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
)

// newSearchTool 鍒濆鍖?DuckDuckGo 鎼滅储宸ュ叿
// 鐢ㄤ簬缃戠粶鎼滅储鏌ヨ
func newSearchTool(ctx context.Context) (bt tool.BaseTool, err error) {
	config.Info("[Tool] 鍒濆鍖?DuckDuckGo 鎼滅储宸ュ叿...")

	searchConfig := &duckduckgo.Config{}
	bt, err = duckduckgo.NewTextSearchTool(ctx, searchConfig)
	if err != nil {
		config.Error("[Tool] DuckDuckGo 鎼滅储宸ュ叿鍒濆鍖栧け璐? %v", err)
		return nil, err
	}

	config.Info("[Tool] DuckDuckGo 搜索工具初始化成功")
	return bt, nil
}