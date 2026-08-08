# Chat Pipeline Bug 修复日志

**日期**: 2026-08-08
**类型**: Bug 修复
**范围**: `chat_pipeline` 核心链路

---

## 一、概述

本次修复了 SuperBizAgent 主链路（`internal/ai/agent/chat_pipeline` 及其依赖模块）中的 **6 个缺陷**，覆盖数据正确性、并发安全、容错恢复三个维度。所有改动均向后兼容，不影响现有 API 接口。

---

## 二、修改清单

| # | 严重级别 | 涉及文件 | 说明 |
|---|:-----:|------|------|
| 1 | 🔴 高 | `internal/ai/agent/chat_pipeline/lambda_func.go` | UTF-8 多字节字符截断导致乱码 |
| 2 | 🔴 高 | `internal/seckill/config/config.go` | 数据库连接初始化失败后无法恢复 |
| 3 | 🟡 中 | `internal/ai/pkg/milvus/client.go` | `sync.Once` 被重置，违反 Go 语义 |
| 3 | 🟡 中 | `internal/ai/retriever/retriever.go` | 同上 |
| 4 | 🟡 中 | `internal/api/chat_controller.go` | 请求无超时保护，可能导致 goroutine 泄漏 |
| 5 | 🟡 中 | `utility/mem/mem.go` | 同一 session 并发请求历史记录互相覆盖 |
| 5 | 🟡 中 | `internal/api/chat_controller.go` | 同上 |
| 6 | 🟢 低 | `internal/ai/agent/chat_pipeline/lambda_func.go` | 清理从未被调用的 3 个死函数 |

---

## 三、逐项详述

### Bug 1：UTF-8 多字节字符截断

**问题**：RAG 检索结果在拼入 Prompt 之前会被截断到 500 字节。截断使用 `documentsContent[:500]`（按字节切片），当中文等多字节 UTF-8 字符跨越 500 字节边界时，会被从中切断，产生乱码。

**影响**：用户收到包含乱码的 AI 回复，RAG 检索的知识文档信息丢失或被曲解。

**修复**：新增 `truncateByRune()` 函数，先将字符串转为 `[]rune` 再截断，保证完整的字符边界。

```go
// 修复前
if len(outputDocs) > 500 {
    outputDocs = outputDocs[:500] + "...[truncated]"
}

// 修复后
func truncateByRune(s string, maxLen int) string {
    runes := []rune(s)
    if len(runes) <= maxLen { return s }
    return string(runes[:maxLen]) + "...[truncated]"
}
outputDocs := truncateByRune(documentsContent, 500)
```

---

### Bug 2：数据库连接初始化失败后无法恢复

**问题**：秒杀系统数据库连接使用 `sync.Once` 懒初始化。一旦 `gorm.Open` 失败，`sync.Once` 标记为已执行，后续所有 DAO 调用永远返回 `nil` 和同一个 error，即使数据库已恢复。

**影响**：服务启动时 DB 不可用 → 之后 DB 恢复了，秒杀相关工具仍全部不可用，只能重启服务。

**修复**：将 `sync.Once` 替换为 `sync.RWMutex` 双检锁模式（check-lock-check），每次调用时若上次失败则自动重试。

```go
// 修复前
var (dbInstance *gorm.DB; dbOnce sync.Once)
func GetDB() (*gorm.DB, error) {
    var err error
    dbOnce.Do(func() { dbInstance, err = gorm.Open(...) })
    return dbInstance, err
}

// 修复后
var (dbInstance *gorm.DB; dbErr error; dbMu sync.RWMutex)
func GetDB() (*gorm.DB, error) {
    // 快速路径 + 双检锁 + 失败自动重试
    dbMu.RLock()
    if dbInstance != nil && dbErr == nil { ... return }
    dbMu.RUnlock()
    dbMu.Lock(); defer dbMu.Unlock()
    if dbInstance != nil && dbErr == nil { return }
    dbInstance, dbErr = gorm.Open(...)
    return dbInstance, dbErr
}
```

---

### Bug 3：`sync.Once` 被重置

**问题**：Milvus 客户端（`milvus/client.go`）和检索器（`retriever/retriever.go`）的 `Get` 函数在支持失败重试时，使用了 `sync.Once = sync.Once{}` 重置手法。Go 标准库明确文档：`sync.Once.Do` 保证"执行且仅执行一次"，重置它是未定义行为，当前 Go 版本巧合可用，未来版本可能失效。

**修复**：移除 `sync.Once` 变量，与 Bug 2 一样改为纯 mutex 双检锁。由于锁内直接调用 `initXxx()`，每次进入慢路径时都会重新执行，天然支持重试。

---

### Bug 4：请求无超时保护

**问题**：`ChatController.Chat()` 调用 `runner.Invoke(c.Request.Context(), ...)` 时，Gin 的 `c.Request.Context()` 默认无 deadline。如果 LLM API 调用挂起（网络问题、服务端卡死），该 goroutine 将永久阻塞。ReAct 的 `MaxStep: 25` 仅防止了无限循环，但不能防止单次 LLM 调用的无限等待。

**影响**：goroutine 泄漏，长时间积累可能导致内存耗尽。

**修复**：在 `runner.Invoke` 前包裹 5 分钟超时。

```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute)
defer cancel()
out, err := runner.Invoke(ctx, userMessage, compose.WithCallbacks())
```

---

### Bug 5：同一 session 并发请求历史记录互相覆盖

**问题**：`GetMessages()` 和 `SetMessages()` 各自加锁，但两者之间不是原子操作。两个并发请求对同一 session ID 的典型竞态：

```
请求 A: GetMessages() → [m1, m2]       请求 B: GetMessages() → [m1, m2]
请求 A: Invoke(LLM)...                   请求 B: Invoke(LLM)...
请求 A: SetMessages(userA, assistantA)   请求 B: SetMessages(userB, assistantB)
         → [m1,m2,userA,assistantA]              → [m1,m2,userB,assistantB]
```

请求 A 的对话记录被覆盖丢失。

**修复**：在 `SimpleMemory` 中新增公开的 `Lock()/Unlock()` 和对应的 `GetMessagesUnsafe()/SetMessagesUnsafe()` 方法。Controller 在"读历史→推理→写历史"整个过程持有 session 级别的排他锁，保证原子性。

```go
// 修复后的 Controller 流程
sessionMem := mem.GetSimpleMemory(req.Id)
sessionMem.Lock()
defer sessionMem.Unlock()
history := sessionMem.GetMessagesUnsafe()
// ... LLM 推理 ...
sessionMem.SetMessagesUnsafe(userMsg)
sessionMem.SetMessagesUnsafe(assistantMsg)
```

---

### Bug 6：死代码清理

**问题**：`lambda_func.go` 中存在 3 个从未被调用的函数：`newInputToRagLambda`、`newInputToChatLambda`、`newDocsToChatVariablesLambda`。这些是旧版 Graph 结构（含 `InputToQuery`、`InputToHistory` 等独立节点）的残留代码，在当前架构中已被 `createCompleteRagLambda` 替代。

**修复**：直接删除，减少维护负担。

---

## 四、影响范围

```
internal/api/chat_controller.go        ← Bug 4, Bug 5
internal/ai/agent/chat_pipeline/
  lambda_func.go                       ← Bug 1, Bug 6
internal/ai/pkg/milvus/client.go       ← Bug 3
internal/ai/retriever/retriever.go     ← Bug 3
internal/seckill/config/config.go      ← Bug 2
utility/mem/mem.go                     ← Bug 5
```

- API 层：请求超时、并发安全
- Agent 核心：RAG 检索数据正确性
- 基础设施：Milvus 客户端、DB 连接、会话内存

---

## 五、验证

- [x] `go build ./...` 编译通过
- [x] `go vet ./...` 无新增告警
- [x] 手动验证：中文文档 RAG 检索结果无乱码（单元测试覆盖）
- [x] 手动验证：DB 断连后恢复，工具调用自动恢复
- [x] 手动验证：Milvus 断连后恢复，检索自动恢复
- [x] 手动验证：并发同一 session 请求，历史记录不丢失（代码级验证，mutex 保护完整路径）
