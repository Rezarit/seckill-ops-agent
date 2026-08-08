# Chat Pipeline Bug 修复验证测试

**日期**: 2026-08-08
**类型**: 验证测试
**关联**: [2026-08-08-chat-pipeline-bugfix.md](./2026-08-08-chat-pipeline-bugfix.md)

---

## 一、测试环境

| 组件 | 版本/来源 | 端口 | 备注 |
|------|----------|:---:|------|
| SuperBizAgent | go build `cmd/server/main.go` | 6872 | 当日最新代码 |
| Milvus Standalone | `milvusdb/milvus:2.6-20260521-4771ac8c-amd64` | 19530 | docker-compose 启动 |
| Milvus Etcd | `quay.io/coreos/etcd:v3.5.18` | 2379-2380 | 容器内网 |
| Milvus MinIO | `minio/minio:RELEASE.2023-03-20T20-16-18Z` | 容器内网 | 宿主机端口映射因 Windows 端口保留已移除 |
| MySQL | `mysql:8.0.45-bookworm` | **3307**→3306 | 容器映射 3307，config.yaml 已改为 3307 |
| LLM | 豆包 `doubao-seed-2-0-mini-260428` | - | 火山 Ark API |
| Embedder | 阿里百炼 `text-embedding-v4` | - | DashScope API |

**配置变更**：
- `config.yaml`：MySQL 端口 3306 → 3307（匹配 Docker 映射）
- `docker-compose.yml`：移除 MinIO 控制台和 Attu 的宿主机端口映射（Windows 端口保留冲突）

---

## 二、测试用例

### Case 1：Bug 1 — UTF-8 字符截断

**方法**：单元测试 `TestTruncateByRune`

| 子用例 | 输入 | 预期 | 实际 |
|--------|------|------|:---:|
| 纯英文短于限制 | `"hello world"`, maxLen=500 | 原样返回 | ✅ |
| 纯英文长于限制 | 600 字节英文字符串, maxLen=500 | 截断至 500 rune + 后缀 | ✅ |
| 纯中文短于限制 | `"你好世界"`, maxLen=500 | 原样返回 | ✅ |
| 纯中文长于限制 | 600 个中文字符, maxLen=500 | 截断至 500 rune + 后缀，无乱码 | ✅ |

---

### Case 2：Bug 2 — DB 断连恢复

| 步骤 | 操作 | 预期 | 实际 |
|------|------|------|:---:|
| 1 | 正常查询秒杀商品 | 返回查询结果 | ✅ 返回 `count=0`（库空但连接正常） |
| 2 | `docker stop mysql-server` | - | ✅ |
| 3 | 查询秒杀商品 | 友好报错，服务不崩溃 | ✅ 返回 `"查询商品失败"`，HTTP 200 |
| 4 | `docker start mysql-server` + 等待就绪 | `mysqld is alive` | ✅ |
| 5 | 查询秒杀商品 | 恢复查询，不再报错 | ✅ 返回正常结果 |

---

### Case 3：Bug 3 — Milvus 断连恢复

| 步骤 | 操作 | 预期 | 实际 |
|------|------|------|:---:|
| 1 | 正常对话 | RAG 检索正常工作 | ✅ |
| 2 | `docker stop milvus-standalone` | - | ✅ |
| 3 | 发送简单对话（不触发工具） | RAG 返回空文档，服务不崩溃 | ✅ HTTP 200，对话正常 |
| 4 | `docker start milvus-standalone` + 等待 healthy | - | ✅ |
| 5 | 发送简单对话 | 恢复检索，服务正常 | ✅ HTTP 200 |

**注**：步骤 3 中，如果用户问题触发其他不可用工具（如 Prometheus），会报 500。这是预期行为——RAG 层的 Milvus 检索本身不会导致崩溃，但 Agent 后续可能因缺少上下文而调用其他不可用的工具。

---

### Case 4：Bug 4 — 请求超时保护

**方法**：代码级验证。`chat_controller.go` 中 `runner.Invoke` 已包裹 `context.WithTimeout(5*time.Minute)`。正常请求在超时内完成即可间接验证。

---

### Case 5：Bug 5 — 并发 session 竞态

**方法**：代码级验证。`SimpleMemory.Lock()` 保护整个"读历史→推理→写历史"路径，`defer Unlock()` 保证异常路径释放。互斥范围覆盖 `ChatController.Chat()` 的步骤 3~6。

---

### Case 6：Bug 6 — 死代码清理

**方法**：`go build ./...` 编译通过，`go vet ./...` 无新增告警。已删除的 3 个函数无任何引用。

---

## 三、结果汇总

| Bug | 测试方式 | 结果 |
|-----|---------|:---:|
| Bug 1 — UTF-8 截断 | 单元测试（4 cases） | ✅ |
| Bug 2 — DB 断连恢复 | 手动：stop → 请求 → start → 请求 | ✅ |
| Bug 3 — Milvus 断连恢复 | 手动：stop → 请求 → start → 请求 | ✅ |
| Bug 4 — 请求超时 | 代码级 review | ✅ |
| Bug 5 — 并发 session | 代码级 review | ✅ |
| Bug 6 — 死代码 | 编译 + vet | ✅ |

---

## 四、备注

1. **Prometheus 不可用**：`localhost:9090` 上无 Prometheus 服务，Agent 调用 `query_prometheus_alerts` 工具时会报 500。这不是本次 Bug 的范围，后续可考虑为工具添加更完善的超时和降级处理。
2. **MCP 日志工具**：config.yaml 中 `mcp_url` 为空，MCP 工具在启动时静默跳过，不影响其他功能。
3. **Windows 端口保留**：宿主机的 9000、9001、9091 等端口被 Hyper-V 保留，修改了 docker-compose.yml 的端口映射以绕过限制。如需恢复原始端口映射，需要管理员权限执行 `netsh int ipv4 add excludedportrange`。
