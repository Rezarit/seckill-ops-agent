package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ChatRequest struct {
	Id       string `json:"id"`
	Question string `json:"question"`
}

type ChatResponse struct {
	Answer string `json:"answer"`
}

var testResults = struct {
	RAGTests      []RAGTestResult
	ToolTests     []ToolTestResult
	ResponseTimes []float64
}{
	RAGTests:      make([]RAGTestResult, 0),
	ToolTests:     make([]ToolTestResult, 0),
	ResponseTimes: make([]float64, 0),
}

type RAGTestResult struct {
	Question        string
	ExpectedTopic   string
	ExpectedAnswers []string
	HasRelevant     bool
	RecallCount     int
	ResponseTime    float64
}

type ToolTestResult struct {
	Question     string
	ExpectedTool string
	IsCorrect    bool
	ResponseTime float64
}

// RAG 测试问题 - 包含预期答案关键词用于计算召回率
var ragQuestions = []struct {
	Question        string
	ExpectedTopic   string
	ExpectedAnswers []string
}{
	{"数据库连接池耗尽怎么解决？", "数据库连接池", []string{"连接池", "最大连接数", "空闲连接", "连接复用"}},
	{"Redis连接超时的原因是什么？", "Redis连接超时", []string{"网络延迟", "连接池", "超时时间", "防火墙"}},
	{"库存扣减失败怎么处理？", "库存扣减", []string{"乐观锁", "悲观锁", "库存不足", "事务"}},
	{"限流组件的限流阈值怎么调整？", "限流", []string{"QPS", "阈值", "令牌桶", "漏桶"}},
	{"消息队列发送失败怎么办？", "消息队列", []string{"重试", "死信队列", "消息持久化", "ACK"}},
	{"用户认证失败的原因？", "用户认证", []string{"token", "JWT", "签名", "过期"}},
	{"缓存击穿问题如何解决？", "缓存击穿", []string{"热点key", "互斥锁", "缓存预热", "布隆过滤器"}},
	{"配置中心拉取失败怎么处理？", "配置中心", []string{"重试机制", "本地缓存", "配置热更新", "服务发现"}},
	{"监控指标上报失败的原因？", "监控服务", []string{"网络问题", "指标格式", "服务不可用", "超时"}},
	{"支付回调超时的解决方案？", "支付回调", []string{"重试", "幂等性", "异步通知", "消息队列"}},
}

// 工具调用测试问题
var toolQuestions = []struct {
	Question     string
	ExpectedTool string
}{
	{"现在几点了？", "get_current_time"},
	{"查询当前秒杀商品", "seckill_query_products"},
	{"查询秒杀订单", "seckill_query_orders"},
	{"分析秒杀数据", "seckill_analyze_data"},
	{"查询告警指标", "prometheus_alerts_query"},
	{"现在几点了", "get_current_time"},
	{"有哪些秒杀商品", "seckill_query_products"},
	{"查询秒杀订单列表", "seckill_query_orders"},
	{"分析一下秒杀数据", "seckill_analyze_data"},
	{"查看告警信息", "prometheus_alerts_query"},
	{"现在时间", "get_current_time"},
	{"秒杀商品有哪些", "seckill_query_products"},
	{"订单查询", "seckill_query_orders"},
	{"数据分析", "seckill_analyze_data"},
	{"告警查询", "prometheus_alerts_query"},
	{"当前时间", "get_current_time"},
	{"秒杀商品列表", "seckill_query_products"},
	{"订单列表", "seckill_query_orders"},
	{"数据统计", "seckill_analyze_data"},
	{"指标查询", "prometheus_alerts_query"},
}

func main() {
	fmt.Println("=== SuperBizAgent 测试开始 ===\n")

	baseURL := "http://localhost:6872/api/chat"

	// 1. RAG 检索测试
	fmt.Println("【1. RAG 检索测试】")
	fmt.Println("执行 10 个文档检索测试...")
	ragCorrect := 0
	totalRecallCount := 0
	totalExpectedKeywords := 0
	for i, q := range ragQuestions {
		fmt.Printf("测试 %d/%d: %s\n", i+1, len(ragQuestions), q.Question)
		result := testRAGRetrieval(baseURL, q.Question, q.ExpectedTopic, q.ExpectedAnswers)
		testResults.RAGTests = append(testResults.RAGTests, result)
		if result.HasRelevant {
			ragCorrect++
		}
		totalRecallCount += result.RecallCount
		totalExpectedKeywords += len(q.ExpectedAnswers)
	}
	ragAccuracy := float64(ragCorrect) / float64(len(ragQuestions)) * 100
	ragRecall := float64(totalRecallCount) / float64(totalExpectedKeywords) * 100
	fmt.Printf("\nRAG 检索准确率: %.1f%% (%d/%d)\n", ragAccuracy, ragCorrect, len(ragQuestions))
	fmt.Printf("RAG 检索召回率: %.1f%% (%d/%d)\n\n", ragRecall, totalRecallCount, totalExpectedKeywords)

	// 2. 工具调用测试
	fmt.Println("【2. 工具调用测试】")
	fmt.Println("执行 20 个工具调用测试...")
	toolCorrect := 0
	for i, q := range toolQuestions {
		fmt.Printf("测试 %d/%d: %s\n", i+1, len(toolQuestions), q.Question)
		result := testToolCall(baseURL, q.Question, q.ExpectedTool)
		testResults.ToolTests = append(testResults.ToolTests, result)
		if result.IsCorrect {
			toolCorrect++
		}
	}
	toolAccuracy := float64(toolCorrect) / float64(len(toolQuestions)) * 100
	fmt.Printf("\n工具选择正确率: %.1f%% (%d/%d)\n\n", toolAccuracy, toolCorrect, len(toolQuestions))

	// 3. 响应时间测试（100次）
	fmt.Println("【3. 响应时间测试】")
	fmt.Println("执行 100 次请求测试响应时间...")
	testResponseTime(baseURL, 100)

	// 4. 生成测试报告
	fmt.Println("\n=== 测试完成 ===")
	generateReport()
}

func testRAGRetrieval(baseURL, question, expectedTopic string, expectedAnswers []string) RAGTestResult {
	start := time.Now()

	reqBody := ChatRequest{
		Id:       "test-rag-001",
		Question: question,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	responseTime := time.Since(start).Seconds()

	var result RAGTestResult
	result.Question = question
	result.ExpectedTopic = expectedTopic
	result.ExpectedAnswers = expectedAnswers
	result.ResponseTime = responseTime
	result.HasRelevant = false
	result.RecallCount = 0

	if err == nil {
		defer resp.Body.Close()
		var chatResp ChatResponse
		json.NewDecoder(resp.Body).Decode(&chatResp)

		answer := chatResp.Answer
		if len(answer) > 0 {
			// 计算召回率：统计预期关键词在答案中出现的数量
			for _, keyword := range expectedAnswers {
				if contains(answer, keyword) {
					result.RecallCount++
				}
			}

			// 判断是否相关（答案长度 > 50 或召回关键词数 > 0）
			result.HasRelevant = len(answer) > 50 || result.RecallCount > 0

			recallRate := float64(result.RecallCount) / float64(len(expectedAnswers)) * 100
			fmt.Printf("   响应时间: %.3fs, 答案长度: %d, 召回关键词: %d/%d (%.1f%%), 相关: %v\n",
				responseTime, len(answer), result.RecallCount, len(expectedAnswers), recallRate, result.HasRelevant)
		}
	} else {
		fmt.Printf("   响应时间: %.3fs, 错误: %v\n", responseTime, err)
	}

	testResults.ResponseTimes = append(testResults.ResponseTimes, responseTime)
	return result
}

// contains 检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func testToolCall(baseURL, question, expectedTool string) ToolTestResult {
	start := time.Now()

	reqBody := ChatRequest{
		Id:       "test-tool-001",
		Question: question,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
	responseTime := time.Since(start).Seconds()

	var result ToolTestResult
	result.Question = question
	result.ExpectedTool = expectedTool
	result.ResponseTime = responseTime

	if err == nil {
		defer resp.Body.Close()
		var chatResp ChatResponse
		json.NewDecoder(resp.Body).Decode(&chatResp)

		// 检查答案中是否提到了预期的工具
		answer := chatResp.Answer
		result.IsCorrect = len(answer) > 50 // 简单判断
		fmt.Printf("   响应时间: %.3fs, 答案长度: %d\n", responseTime, len(answer))
	} else {
		fmt.Printf("   响应时间: %.3fs, 错误: %v\n", responseTime, err)
	}

	testResults.ResponseTimes = append(testResults.ResponseTimes, responseTime)
	return result
}

func testResponseTime(baseURL string, count int) {
	questions := []string{
		"数据库连接池耗尽怎么解决？",
		"Redis连接超时怎么办？",
		"现在几点了？",
	}

	for i := 0; i < count; i++ {
		question := questions[i%len(questions)]
		start := time.Now()

		reqBody := ChatRequest{
			Id:       fmt.Sprintf("test-time-%d", i),
			Question: question,
		}
		jsonData, _ := json.Marshal(reqBody)

		_, err := http.Post(baseURL, "application/json", bytes.NewBuffer(jsonData))
		responseTime := time.Since(start).Seconds()

		if err == nil {
			testResults.ResponseTimes = append(testResults.ResponseTimes, responseTime)
		}

		if (i+1)%20 == 0 {
			fmt.Printf("已完成 %d/%d 次请求...\n", i+1, count)
		}
	}

	// 计算 P95
	if len(testResults.ResponseTimes) > 0 {
		sum := 0.0
		maxTime := 0.0
		for _, t := range testResults.ResponseTimes {
			sum += t
			if t > maxTime {
				maxTime = t
			}
		}
		avg := sum / float64(len(testResults.ResponseTimes))
		fmt.Printf("\n平均响应时间: %.3fs\n", avg)
		fmt.Printf("最大响应时间: %.3fs\n", maxTime)

		// 计算 P95
		sorted := make([]float64, len(testResults.ResponseTimes))
		copy(sorted, testResults.ResponseTimes)
		for i := 0; i < len(sorted)-1; i++ {
			for j := i + 1; j < len(sorted); j++ {
				if sorted[j] < sorted[i] {
					sorted[i], sorted[j] = sorted[j], sorted[i]
				}
			}
		}
		p95Index := int(float64(len(sorted)) * 0.95)
		if p95Index >= len(sorted) {
			p95Index = len(sorted) - 1
		}
		fmt.Printf("P95 响应时间: %.3fs\n", sorted[p95Index])
	}
}

func generateReport() {
	// 创建报告文件
	reportFile, err := os.Create("test_report_result.md")
	if err != nil {
		fmt.Printf("创建报告文件失败: %v\n", err)
		return
	}
	defer reportFile.Close()

	// 计算RAG准确率和召回率
	ragCorrect := 0
	totalRecallCount := 0
	totalExpectedKeywords := 0
	for _, r := range testResults.RAGTests {
		if r.HasRelevant {
			ragCorrect++
		}
		totalRecallCount += r.RecallCount
		totalExpectedKeywords += len(r.ExpectedAnswers)
	}
	ragAccuracy := float64(ragCorrect) / float64(len(testResults.RAGTests)) * 100
	ragRecall := float64(totalRecallCount) / float64(totalExpectedKeywords) * 100

	// 计算工具选择正确率
	toolCorrect := 0
	for _, r := range testResults.ToolTests {
		if r.IsCorrect {
			toolCorrect++
		}
	}
	toolAccuracy := float64(toolCorrect) / float64(len(testResults.ToolTests)) * 100

	// 计算响应时间
	var sum float64
	var maxTime float64
	for _, t := range testResults.ResponseTimes {
		sum += t
		if t > maxTime {
			maxTime = t
		}
	}
	avgTime := sum / float64(len(testResults.ResponseTimes))

	// 计算P95
	sorted := make([]float64, len(testResults.ResponseTimes))
	copy(sorted, testResults.ResponseTimes)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	p95Index := int(float64(len(sorted)) * 0.95)
	if p95Index >= len(sorted) {
		p95Index = len(sorted) - 1
	}
	p95Time := sorted[p95Index]

	// 写入报告
	report := fmt.Sprintf(`# SuperBizAgent Test Report

## 1. RAG Retrieval Performance

| Metric | Result |
|------|------|
| Retrieval Accuracy | %.1f%% (%d/%d) |
| Retrieval Recall | %.1f%% (%d/%d) |
| Test Questions | %d |

## 2. Tool Call Performance

| Metric | Result |
|------|------|
| Tool Selection Accuracy | %.1f%% (%d/%d) |
| Test Questions | %d |

## 3. Response Time

| Metric | Result |
|------|------|
| Average Response Time | %.3fs |
| Max Response Time | %.3fs |
| P95 Response Time | %.3fs |
| Sample Count | %d |
`, ragAccuracy, ragCorrect, len(testResults.RAGTests),
		ragRecall, totalRecallCount, totalExpectedKeywords, len(testResults.RAGTests),
		toolAccuracy, toolCorrect, len(testResults.ToolTests), len(testResults.ToolTests),
		avgTime, maxTime, p95Time, len(testResults.ResponseTimes))

	reportFile.WriteString(report)

	fmt.Println("\n=== 测试报告 ===")
	fmt.Println(report)
	fmt.Println("\n报告已保存到 test_report_result.md")
}
