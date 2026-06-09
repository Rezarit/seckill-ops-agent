package main

import (
	"SuperBizAgent/config"
	"SuperBizAgent/internal/ai/embedder"
	"SuperBizAgent/internal/ai/splitter"
	"SuperBizAgent/internal/ai/pkg/milvus"
	"SuperBizAgent/utility/common"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func main() {
	fmt.Println("=== 开始加载文档到向量库 ===")

	// 自动切换到项目根目录
	cwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("获取当前工作目录失败: %v", err))
	}

	testPath := cwd
	for {
		if _, err := os.Stat(filepath.Join(testPath, "manifest", "config", "config.yaml")); err == nil {
			break
		}
		parent := filepath.Dir(testPath)
		if parent == testPath {
			panic("无法找到项目根目录（找不到 manifest/config/config.yaml）")
		}
		testPath = parent
	}
	projectRoot := testPath

	if err := os.Chdir(projectRoot); err != nil {
		panic(fmt.Sprintf("切换到项目根目录失败: %v", err))
	}
	fmt.Printf("[Info] 项目根目录: %s\n", projectRoot)

	// 初始化配置
	fmt.Println("[Info] 初始化配置...")
	if err := config.InitConfig(); err != nil {
		panic(fmt.Sprintf("配置初始化失败: %v", err))
	}
	fmt.Println("[Info] 配置初始化完成")

	ctx := context.Background()

	// 重新创建 Collection
	fmt.Println("[Info] 重新创建 Collection...")
	err = milvus.DropAndRecreateCollection(ctx)
	if err != nil {
		panic(fmt.Sprintf("重新创建 Collection 失败: %v", err))
	}
	fmt.Println("[Info] Collection 重新创建成功")

	// 初始化 embedder
	fmt.Println("[Info] 初始化 Embedder...")
	eb, err := embedder.DoubaoEmbedding(ctx)
	if err != nil {
		panic(fmt.Sprintf("Embedder 初始化失败: %v", err))
	}
	fmt.Println("[Info] Embedder 初始化成功")

	// 初始化 Milvus 客户端
	fmt.Println("[Info] 初始化 Milvus 客户端...")
	milvusCli, err := milvus.NewMilvusClient(ctx)
	if err != nil {
		panic(fmt.Sprintf("Milvus 客户端初始化失败: %v", err))
	}
	fmt.Println("[Info] Milvus 客户端初始化成功")

	// 遍历文档目录
	docsDir := filepath.Join(projectRoot, "internal", "ai", "cmd", "knowledge_cmd", "docs")
	fmt.Printf("[Info] 扫描文档目录: %s\n", docsDir)

	// 检查目录是否存在
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		fmt.Printf("[Error] 文档目录不存在: %s\n", docsDir)
		return
	}

	totalIndexed := 0
	totalFailed := 0

	err = filepath.WalkDir(docsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walk dir failed: %w", err)
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			fmt.Printf("[Skip] 非 Markdown 文件: %s\n", path)
			return nil
		}

		// 统一路径分隔符
		normalizedPath := strings.ReplaceAll(path, "\\", "/")
		fmt.Printf("\n[Start] 处理文件: %s\n", normalizedPath)

		// 1. 直接读取文件内容
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("[Error] 读取文件失败: %v\n", err)
			return err
		}
		text := string(content)
		if strings.TrimSpace(text) == "" {
			fmt.Printf("[Warn] 文件为空: %s\n", normalizedPath)
			return nil
		}

		fmt.Printf("[Info] 文件读取成功，内容长度: %d\n", len(text))

		// 2. 分块
		mdSplitter := splitter.NewMarkdownSplitter()
		chunks, err := mdSplitter.SplitText(ctx, text, map[string]any{"source": normalizedPath})
		if err != nil {
			fmt.Printf("[Error] 分块失败: %v\n", err)
			return err
		}
		fmt.Printf("[Info] 分块完成，块数: %d\n", len(chunks))

		// 3. 向量化并插入
		ids := make([]string, 0, len(chunks))
		vectors := make([][]float32, 0, len(chunks))
		contents := make([]string, 0, len(chunks))
		metadatas := make([]string, 0, len(chunks))

		for i, chunk := range chunks {
			vec, err := eb.EmbedStrings(ctx, []string{chunk.Content})
			if err != nil {
				fmt.Printf("[Error] 向量化失败 (块 %d): %v\n", i, err)
				totalFailed++
				continue
			}

			if len(vec) == 0 || len(vec[0]) == 0 {
				fmt.Printf("[Warn] 向量为空 (块 %d)\n", i)
				continue
			}

			// 转换向量类型：[]float64 → []float32
			float32Vec := make([]float32, len(vec[0]))
			for j, v := range vec[0] {
				float32Vec[j] = float32(v)
			}

			id := uuid.New().String()
			metadataJSON, _ := json.Marshal(map[string]interface{}{
				"_source":    normalizedPath,
				"_file_name": d.Name(),
				"_extension": ".md",
			})

			ids = append(ids, id)
			vectors = append(vectors, float32Vec)
			contents = append(contents, chunk.Content)
			metadatas = append(metadatas, string(metadataJSON))
		}

		// 逐条插入
		successCount := 0
		for i := range ids {
			insertData := map[string]any{
				"id":       ids[i],
				"vector":   vectors[i],
				"content":  contents[i],
				"metadata": metadatas[i],
			}

			insertOption := milvusclient.NewRowBasedInsertOption(common.MilvusCollectionName, insertData)
			_, err = milvusCli.Insert(ctx, insertOption)
			if err != nil {
				fmt.Printf("[Error] 插入失败 (第 %d 条): %v\n", i+1, err)
				totalFailed++
			} else {
				successCount++
			}
		}

		if successCount > 0 {
			fmt.Printf("[Success] 成功插入 %d 条记录\n", successCount)
			totalIndexed += successCount
		}

		fmt.Printf("[Done] 文件处理完成: %s\n", normalizedPath)

		return nil
	})

	// Flush 数据
	fmt.Println("\n[Info] 正在 Flush 数据到磁盘...")
	flushOption := milvusclient.NewFlushOption(common.MilvusCollectionName)
	_, err = milvusCli.Flush(ctx, flushOption)
	if err != nil {
		fmt.Printf("[Warn] Flush 失败: %v\n", err)
	} else {
		fmt.Println("[Info] Flush 成功")
	}

	// 创建向量索引（确保检索可用）
	fmt.Println("[Info] 正在创建向量索引...")
	vectorIndex := index.NewAutoIndex(entity.L2)
	createIndexOption := milvusclient.NewCreateIndexOption(common.MilvusCollectionName, "vector", vectorIndex)
	_, err = milvusCli.CreateIndex(ctx, createIndexOption)
	if err != nil {
		fmt.Printf("[Warn] 创建索引失败（可能已存在）: %v\n", err)
	} else {
		fmt.Println("[Info] 向量索引创建成功")
	}

	// 重新加载 Collection
	fmt.Println("[Info] 重新加载 Collection...")
	loadOption := milvusclient.NewLoadCollectionOption(common.MilvusCollectionName)
	_, err = milvusCli.LoadCollection(ctx, loadOption)
	if err != nil {
		fmt.Printf("[Warn] 加载失败: %v\n", err)
	} else {
		fmt.Println("[Info] Collection 加载成功")
	}

	if err != nil {
		fmt.Printf("\n[Error] 文档加载过程中出现错误: %v\n", err)
	} else {
		fmt.Printf("\n=== 所有文档加载完成 ===\n")
		fmt.Printf("[Summary] 总插入: %d 块，失败: %d 块\n", totalIndexed, totalFailed)
	}
}
