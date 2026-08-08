# 配置工程化改造——环境变量与端口管理

**日期**: 2026-08-08
**类型**: 配置改造
**关联**: [2026-08-08-chat-pipeline-bugfix.md](./2026-08-08-chat-pipeline-bugfix.md)

---

## 一、背景

1. Windows 开发机上 Hyper-V 保留了端口段 8983~9182，导致 docker-compose 中 MinIO（9000/9001）和 Milvus 健康检查（9091）无法绑定，容器启动失败。
2. `docker-compose.yml` 端口硬编码，换环境（开发/测试/生产）需直接改文件，不符合 12-Factor App 规范。
3. 有 `config.example.yaml` 模板，但 Go 代码不支持环境变量覆盖。

---

## 二、改造内容

### 2.1 docker-compose.yml：端口环境变量化

**修改前**（硬编码）：
```yaml
ports:
  - "19530:19530"
  - "9091:9091"
  - "8000:3000"
```

**修改后**（通过 `.env` 文件配置）：
```yaml
ports:
  - "${MILVUS_PORT:-19530}:19530"
  - "${MILVUS_HEALTH_PORT:-9091}:9091"
  - "${ATTU_PORT:-8000}:3000"
```

语法 `${VAR:-default}` 表示：读取环境变量 `VAR`，未设置时使用 `default`。不提供 `.env` 文件也能用默认值启动。

### 2.2 新增 .env / .env.example

| 文件 | 用途 | Git |
|------|------|:---:|
| `.env.example` | 模板，含注释说明每个变量的含义 | ✅ 提交 |
| `.env` | 本机实际配置，按需修改端口 | ❌ `.gitignore` |

**本机 `.env` 当前值**：
```ini
MILVUS_PORT=19530
MILVUS_HEALTH_PORT=19091    # 避开 Windows Hyper-V 保留段 8983~9182
ATTU_PORT=8000
```

### 2.3 MinIO 控制台端口

移除 MinIO 的宿主机端口映射。MinIO 仅被 Milvus 容器内网访问（`minio:9000`），不需要暴露到宿主机，也避免了端口冲突。

### 2.4 config.yaml

MySQL 端口恢复为 3306（之前为适配 Docker MySQL 容器临时改为 3307）。

---

## 三、使用方式

```bash
# 首次使用：从模板创建本机配置
cd manifest/docker
cp .env.example .env

# 按需编辑 .env（本机端口冲突时改高位端口即可）
# MILVUS_HEALTH_PORT=19091

# 启动
docker compose --env-file .env up -d

# 切换环境：换一个 .env 文件
docker compose --env-file .env.prod up -d
```

---

## 四、验证

| 测试项 | 结果 |
|--------|:---:|
| `docker compose --env-file .env up -d` 正常启动 | ✅ |
| Milvus standalone 健康检查通过 | ✅ |
| 基础对话 / 工具调用 / RAG 检索 均正常 | ✅ |
| Attu 管理界面可访问 | ✅ |

---

## 五、后续建议

1. **Go 配置环境变量化**：`config.go` 中的 API Key、数据库密码等敏感信息，建议通过 Viper 的 `AutomaticEnv()` 或 `BindEnv()` 支持环境变量覆盖，避免敏感信息写入 `config.yaml`。
2. **生产环境 .env**：服务器部署时可创建 `.env.prod`，将 Attu 端口设为 0 或去掉映射（生产环境不需要管理界面）。
3. **Prometheus 工具**：当前 Agent 有时会误将普通问题路由到 Prometheus 工具（如"几点了"触发了 `query_prometheus_alerts`），可考虑优化 System Prompt 或工具描述来改善路由准确性。
