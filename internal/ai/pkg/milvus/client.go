// Package milvus 提供 Milvus 向量数据库的客户端封装
// 包含单例客户端管理、Collection 创建、索引管理等功统
// 使用 sync.Once 保证全局只初始化一次
package milvus

import (
	"SuperBizAgent/config"
	"SuperBizAgent/utility/common"
	"context"
	"fmt"
	"sync"

	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

var (
	milvusClient     *milvusclient.Client // 全局单例 Milvus 客户端
	milvusClientOnce sync.Once            // 确保只初始化一次
	milvusClientErr  error                // 初始化错误
	milvusClientMu   sync.RWMutex         // 客户端访问锁
)

// GetMilvusClient 获取全局单例 Milvus 客户端
// 如果首次初始化失败，支持重试连接
func GetMilvusClient(ctx context.Context) (*milvusclient.Client, error) {
	// 快速路径：检查是否已有有效客户端
	milvusClientMu.RLock()
	client := milvusClient
	err := milvusClientErr
	milvusClientMu.RUnlock()

	if client != nil && err == nil {
		return client, nil
	}

	// 需要初始化或重试
	milvusClientMu.Lock()
	defer milvusClientMu.Unlock()

	// 双重检查
	if milvusClient != nil && milvusClientErr == nil {
		return milvusClient, nil
	}

	// 允许重试：重置错误状态
	milvusClientOnce = sync.Once{}

	milvusClientOnce.Do(func() {
		config.Info("[Milvus Client] ========== 初始化 Milvus 客户端（支持重试）==========")
		milvusClient, milvusClientErr = initMilvusClient(ctx)
		if milvusClientErr != nil {
			config.Error("[Milvus Client] 客户端初始化失败: %v", milvusClientErr)
		} else {
			config.Info("[Milvus Client] 客户端初始化成功")
		}
	})
	return milvusClient, milvusClientErr
}

// NewMilvusClient 保留旧接口以兼容外部调用
// 内部委托给 GetMilvusClient
func NewMilvusClient(ctx context.Context) (*milvusclient.Client, error) {
	return GetMilvusClient(ctx)
}

// DropAndRecreateCollection 删除并重新创建 Collection
// 用于解决 schema 不匹配或数据迁移场景
func DropAndRecreateCollection(ctx context.Context) error {
	cli, err := GetMilvusClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Milvus client: %w", err)
	}

	// 删除旧的 collection
	config.Info("[Milvus Client] 删除旧的 Collection '%s'...", common.MilvusCollectionName)
	err = cli.DropCollection(ctx, milvusclient.NewDropCollectionOption(common.MilvusCollectionName))
	if err != nil {
		config.Warn("[Milvus Client] 删除 Collection 返回: %v", err)
	} else {
		config.Info("[Milvus Client] Collection 删除成功")
	}

	// 创建新的 collection
	config.Info("[Milvus Client] 创建新的 Collection '%s'...", common.MilvusCollectionName)
	schema := buildCollectionSchema()

	err = cli.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(common.MilvusCollectionName, schema))
	if err != nil {
		config.Error("[Milvus Client] 创建 Collection 失败: %v", err)
		return fmt.Errorf("failed to create collection: %w", err)
	}
	config.Info("[Milvus Client] Collection 创建成功")

	// 创建向量索引
	if err := createVectorIndex(ctx, cli); err != nil {
		return err
	}

	// 加载 Collection
	if err := loadCollection(ctx, cli); err != nil {
		return err
	}

	return nil
}

// buildCollectionSchema 构建 Collection 的 schema 定义
func buildCollectionSchema() *entity.Schema {
	return entity.NewSchema().
		WithName(common.MilvusCollectionName).
		WithDescription("Business knowledge collection").
		WithField(entity.NewField().
			WithName("id").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(256).
			WithIsPrimaryKey(true)).
		WithField(entity.NewField().
			WithName("vector").
			WithDataType(entity.FieldTypeFloatVector).
			WithDim(int64(config.AppConfig.DoubaoEmb.EmbeddingDim))).
		WithField(entity.NewField().
			WithName("content").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(16384)).
		WithField(entity.NewField().
			WithName("metadata").
			WithDataType(entity.FieldTypeVarChar).
			WithMaxLength(2048))
}

// createVectorIndex 创建向量索引
func createVectorIndex(ctx context.Context, cli *milvusclient.Client) error {
	config.Info("[Milvus Client] 为 Collection '%s' 创建向量索引...", common.MilvusCollectionName)
	vectorIndex := index.NewAutoIndex(entity.L2)
	_, err := cli.CreateIndex(ctx, milvusclient.NewCreateIndexOption(common.MilvusCollectionName, "vector", vectorIndex))
	if err != nil {
		config.Error("[Milvus Client] 创建向量索引失败: %v", err)
		return fmt.Errorf("failed to create vector index: %w", err)
	}
	config.Info("[Milvus Client] 向量索引创建成功")
	return nil
}

// loadCollection 加载 Collection 到内存
func loadCollection(ctx context.Context, cli *milvusclient.Client) error {
	config.Info("[Milvus Client] 加载 Collection '%s'...", common.MilvusCollectionName)
	_, err := cli.LoadCollection(ctx, milvusclient.NewLoadCollectionOption(common.MilvusCollectionName))
	if err != nil {
		config.Error("[Milvus Client] 加载 Collection 失败: %v", err)
		return fmt.Errorf("failed to load collection: %w", err)
	}
	config.Info("[Milvus Client] Collection 加载成功")
	return nil
}

// initMilvusClient 初始化 Milvus 客户端
// 执行流程:
// 1. 连接 Milvus 服务器
// 2. 检查创建数据库
// 3. 连接数据库
// 4. 检查创建 Collection
// 5. 加载 Collection
func initMilvusClient(ctx context.Context) (*milvusclient.Client, error) {
	// 步骤1: 连接 Milvus 服务器
	config.Info("[Milvus Client] 步骤1/5: 连接 Milvus 服务器 (localhost:19530)...")
	defaultClient, err := milvusclient.New(ctx, &milvusclient.ClientConfig{
		Address: "localhost:19530",
	})
	if err != nil {
		config.Error("[Milvus Client] 连接 Milvus 失败: %v", err)
		return nil, fmt.Errorf("failed to connect to default database: %w", err)
	}
	config.Info("[Milvus Client] 连接到 Milvus 服务器成功")

	// 步骤2: 检查数据库
	if err := ensureDatabaseExists(ctx, defaultClient); err != nil {
		return nil, err
	}

	// 步骤3: 连接数据库
	if err := connectToDatabase(ctx, defaultClient); err != nil {
		return nil, err
	}

	// 步骤4: 检查创建 Collection
	if err := ensureCollectionExists(ctx, defaultClient); err != nil {
		return nil, err
	}

	// 步骤5: 加载 Collection
	if err := loadCollection(ctx, defaultClient); err != nil {
		return nil, err
	}

	config.Info("[Milvus Client] ========== Milvus 客户端初始化完成 ==========")
	config.Info("[Milvus Client] 当前状态: Database=%s, Collection=%s, Dim=%d",
		common.MilvusDBName, common.MilvusCollectionName, config.AppConfig.DoubaoEmb.EmbeddingDim)

	return defaultClient, nil
}

// ensureDatabaseExists 确保数据库存在，不存在则创建
func ensureDatabaseExists(ctx context.Context, cli *milvusclient.Client) error {
	config.Info("[Milvus Client] 步骤2/5: 检查数据库 '%s'...", common.MilvusDBName)
	databases, err := cli.ListDatabase(ctx, milvusclient.NewListDatabaseOption())
	if err != nil {
		config.Error("[Milvus Client] 列出数据库失败: %v", err)
		return fmt.Errorf("failed to list databases: %w", err)
	}

	agentDBExists := false
	for _, dbName := range databases {
		if dbName == common.MilvusDBName {
			agentDBExists = true
			break
		}
	}

	if !agentDBExists {
		config.Info("[Milvus Client] 数据库 '%s' 不存在，正在创建...", common.MilvusDBName)
		err = cli.CreateDatabase(ctx, milvusclient.NewCreateDatabaseOption(common.MilvusDBName))
		if err != nil {
			config.Error("[Milvus Client] 创建数据库失败: %v", err)
			return fmt.Errorf("failed to create agent database: %w", err)
		}
		config.Info("[Milvus Client] 数据库 '%s' 创建成功", common.MilvusDBName)
	} else {
		config.Info("[Milvus Client] 数据库 '%s' 已存在", common.MilvusDBName)
	}
	return nil
}

// connectToDatabase 连接到指定数据库
func connectToDatabase(ctx context.Context, cli *milvusclient.Client) error {
	config.Info("[Milvus Client] 步骤3/5: 连接数据库 '%s'...", common.MilvusDBName)
	err := cli.UseDatabase(ctx, milvusclient.NewUseDatabaseOption(common.MilvusDBName))
	if err != nil {
		config.Error("[Milvus Client] 连接数据库 '%s' 失败: %v", common.MilvusDBName, err)
		return fmt.Errorf("failed to connect to agent database: %w", err)
	}
	config.Info("[Milvus Client] 连接数据库 '%s' 成功", common.MilvusDBName)
	return nil
}

// ensureCollectionExists 确保 Collection 存在，不存在则创建
func ensureCollectionExists(ctx context.Context, cli *milvusclient.Client) error {
	config.Info("[Milvus Client] 步骤4/5: 检查 Collection '%s'...", common.MilvusCollectionName)
	collections, err := cli.ListCollections(ctx, milvusclient.NewListCollectionOption())
	if err != nil {
		config.Error("[Milvus Client] 列出 Collections 失败: %v", err)
		return fmt.Errorf("failed to list collections: %w", err)
	}

	collectionExists := false
	for _, collName := range collections {
		if collName == common.MilvusCollectionName {
			collectionExists = true
			break
		}
	}

	if !collectionExists {
		// 创建新的 Collection
		config.Info("[Milvus Client] Collection '%s' 不存在，正在创建...", common.MilvusCollectionName)
		schema := buildCollectionSchema()

		err = cli.CreateCollection(ctx, milvusclient.NewCreateCollectionOption(common.MilvusCollectionName, schema))
		if err != nil {
			config.Error("[Milvus Client] 创建 Collection 失败: %v", err)
			return fmt.Errorf("failed to create biz collection: %w", err)
		}
		config.Info("[Milvus Client] Collection '%s' 创建成功", common.MilvusCollectionName)

		// 创建向量索引
		if err := createVectorIndex(ctx, cli); err != nil {
			return err
		}
	} else {
		// Collection 已存在，确保索引存在
		config.Info("[Milvus Client] Collection '%s' 已存在，检查向量索引...", common.MilvusCollectionName)
		config.Info("[Milvus Client] 确保向量索引存在...")
		vectorIndex := index.NewAutoIndex(entity.L2)
		_, err = cli.CreateIndex(ctx, milvusclient.NewCreateIndexOption(common.MilvusCollectionName, "vector", vectorIndex))
		if err != nil {
			config.Info("[Milvus Client] 索引创建返回: %v (可能已存在)", err)
		} else {
			config.Info("[Milvus Client] 向量索引创建成功")
		}
	}
	return nil
}
