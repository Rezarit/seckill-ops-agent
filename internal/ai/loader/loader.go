package loader

import (
	"SuperBizAgent/config"
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

func NewFileLoader(ctx context.Context) (ldr document.Loader, err error) {
	config.Info("[Loader] 寮€濮嬪垵濮嬪寲鏂囦欢鍔犺浇鍣?..")

	loaderConfig := &file.FileLoaderConfig{}
	ldr, err = file.NewFileLoader(ctx, loaderConfig)
	if err != nil {
		config.Error("[Loader] 鍒濆鍖栧け璐? %v", err)
		return nil, err
	}

	config.Info("[Loader] 初始化成功")
	return ldr, nil
}
