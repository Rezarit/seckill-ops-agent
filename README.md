# SuperBizAgent

基于 Go 语言 + Eino 框架构建的企业级 AI Agent 服务，专注于业务智能对话和自动化任务执行。

## ✨ 功能特性

- **智能对话**: 集成豆包大语言模型，支持多轮对话
- **RAG 文档检索**: 基于 Milvus 向量数据库的内部文档查询
- **工具调用**: 支持多种工具扩展（数据库操作、秒杀系统、指标查询等）
- **优雅降级**: 工具创建失败不影响服务启动
- **结构化日志**: 使用 Zap 实现高性能日志系统

## 🛠️ 技术栈

| 分类 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.25+ |
| Web 框架 | Gin | 1.12.0 |
| AI Agent | Eino | 0.9.2 |
| 向量数据库 | Milvus | 2.6.5 |
| 日志系统 | Zap | 1.28.0 |
| 配置管理 | Viper | 1.21.0 |
| ORM | GORM | 1.31.1 |

## 📁 项目结构

```
SuperBizAgent/
├── cmd/server/main.go      # 服务入口
├── config/config.go        # 配置管理
├── internal/
│   ├── ai/                 # AI 核心模块
│   │   ├── agent/          # Agent 工作流
│   │   ├── tools/          # 工具集（7个工具）
│   │   ├── models/         # LLM 模型适配
│   │   ├── embedder/       # 嵌入服务
│   │   └── retriever/      # 向量检索
│   ├── api/                # HTTP API 控制器
│   └── seckill/            # 秒杀业务系统
├── manifest/docker/        # Docker 部署配置
└── SuperBizAgentFrontend/  # 前端项目
```

## 🚀 快速开始

### 环境要求

- Go 1.25+
- Milvus 2.6+
- MySQL 5.7+

### 1. 启动依赖服务

```bash
cd manifest/docker
docker-compose up -d
```

### 2. 配置

```bash
cp config.example.yaml config.yaml
# 编辑 config.yaml，填写 API Key 和数据库配置
```

### 3. 启动服务

```bash
go run ./cmd/server/main.go
```

### 4. 访问服务

```bash
# 健康检查
curl http://localhost:6872/api/info

# 聊天接口
curl -X POST http://localhost:6872/api/chat \
  -H "Content-Type: application/json" \
  -d '{"message": "你好"}'
```

## 📡 API 接口

### POST /api/chat

智能对话接口

**请求体**:
```json
{
  "message": "查询内部文档",
  "history": []
}
```

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "reply": "这是查询结果...",
    "tool_used": "query_internal_docs"
  }
}
```

### GET /api/info

服务信息接口

**响应**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "version": "1.0.0",
    "status": "running"
  }
}
```

## 🧰 工具列表

| 工具 | 功能 | 状态 |
|------|------|------|
| `query_internal_docs` | 内部文档 RAG 查询 | ✅ |
| `get_current_time` | 当前时间查询 | ✅ |
| `mysql_crud` | MySQL 数据库操作 | ✅ |
| `query_metrics_alerts` | 告警指标查询 | ✅ |
| `seckill_query_products` | 秒杀商品查询 | ✅ |
| `seckill_query_orders` | 秒杀订单查询 | ✅ |
| `seckill_analyze_data` | 秒杀数据分析 | ✅ |

## 📝 License

MIT License - 详见 [LICENSE](LICENSE)
