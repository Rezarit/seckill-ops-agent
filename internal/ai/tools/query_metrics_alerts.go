// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// PrometheusAlert Prometheus 告警信息结构
type PrometheusAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	State       string            `json:"state"`
	ActiveAt    string            `json:"activeAt"`
	Value       string            `json:"value"`
}

// PrometheusAlertsResult 告警查询结果结构
type PrometheusAlertsResult struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []PrometheusAlert `json:"alerts"`
	} `json:"data"`
	Error     string `json:"error,omitempty"`
	ErrorType string `json:"errorType,omitempty"`
}

// SimplifiedAlert 简化的告警信息（用于输出）
type SimplifiedAlert struct {
	AlertName   string `json:"alert_name" jsonschema:"description=告警名称，从 Prometheus 告警的 labels.alertname 字段提取"`
	Description string `json:"description" jsonschema:"description=告警描述信息，从 Prometheus 告警的 annotations.description 字段提取"`
	State       string `json:"state" jsonschema:"description=告警状态，通常为 'firing'（触发中）或 'pending'（待触发）"`
	ActiveAt    string `json:"active_at" jsonschema:"description=告警激活时间，RFC3339 格式的时间戳，例如 '2025-10-29T08:48:42.496134755Z'"`
	Duration    string `json:"duration" jsonschema:"description=告警持续时间，从激活时间到当前时间的时长，格式如 '2h30m15s'、'30m15s' 或 '15s'"`
}

// PrometheusAlertsOutput 告警查询输出结构
type PrometheusAlertsOutput struct {
	Success bool              `json:"success" jsonschema:"description=查询是否成功"`
	Alerts  []SimplifiedAlert `json:"alerts,omitempty" jsonschema:"description=活动告警列表，每个告警包含名称、描述、状态、激活时间和持续时间。相同 alertname 的告警只保留第一个"`
	Message string            `json:"message,omitempty" jsonschema:"description=操作结果的状态消息"`
	Error   string            `json:"error,omitempty" jsonschema:"description=如果查询失败，包含错误信息"`
}

// queryPrometheusAlerts 查询 Prometheus 告警
// 注意：当前版本返回空结果（Prometheus 服务未启用）
func queryPrometheusAlerts() (PrometheusAlertsResult, error) {
	// 开关：Prometheus 服务未启用时返回空结果
	config.Warn("[Tool-query_prometheus_alerts] Prometheus 服务当前未启用，返回空结果")
	return PrometheusAlertsResult{}, nil

	// 以下为启用 Prometheus 时的代码
	// baseURL := "http://127.0.0.1:9090"
	// apiURL := fmt.Sprintf("%s/api/v1/alerts", baseURL)
	// config.Info("[Tool-query_prometheus_alerts] 查询 Prometheus 告警: %s", apiURL)
	// client := &http.Client{Timeout: 10 * time.Second}
	// resp, err := client.Get(apiURL)
	// if err != nil {
	// 	return PrometheusAlertsResult{}, fmt.Errorf("查询失败: %v", err)
	// }
	// defer resp.Body.Close()
	// body, _ := io.ReadAll(resp.Body)
	// var result PrometheusAlertsResult
	// json.Unmarshal(body, &result)
	// return result, nil
}

// calculateDuration 计算从激活时间到当前的持续时间
func calculateDuration(activeAtStr string) string {
	activeAt, err := time.Parse(time.RFC3339Nano, activeAtStr)
	if err != nil {
		return "unknown"
	}

	duration := time.Since(activeAt)
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	} else {
		return fmt.Sprintf("%ds", seconds)
	}
}

// NewPrometheusAlertsQueryTool 创建 Prometheus 告警查询工具
// 查询 Prometheus 告警系统中的活动告警
func NewPrometheusAlertsQueryTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"query_prometheus_alerts",
		"Query active alerts from Prometheus alerting system. This tool retrieves all currently active/firing alerts including their labels, annotations, state, and values.",
		func(ctx context.Context, input *struct{}, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-query_prometheus_alerts] 工具调用开始")

			// 调用 Prometheus Alerts API
			result, err := queryPrometheusAlerts()
			if err != nil {
				alertsOut := PrometheusAlertsOutput{
					Success: false,
					Error:   err.Error(),
					Message: "查询 Prometheus 告警失败",
				}
				jsonBytes, _ := json.MarshalIndent(alertsOut, "", "  ")
				config.Error("[Tool-query_prometheus_alerts] 查询失败: %v", err)
				return string(jsonBytes), err
			}

			// 转换为简化格式（相同 alertname 只保留第一个）
			seenAlertNames := make(map[string]bool)
			simplifiedAlerts := make([]SimplifiedAlert, 0)
			for _, alert := range result.Data.Alerts {
				alertName := alert.Labels["alertname"]
				if seenAlertNames[alertName] {
					continue
				}
				seenAlertNames[alertName] = true

				simplified := SimplifiedAlert{
					AlertName:   alertName,
					Description: alert.Annotations["description"],
					State:       alert.State,
					ActiveAt:    alert.ActiveAt,
					Duration:    calculateDuration(alert.ActiveAt),
				}
				simplifiedAlerts = append(simplifiedAlerts, simplified)
			}

			// 构建成功响应
			alertsOut := PrometheusAlertsOutput{
				Success: true,
				Alerts:  simplifiedAlerts,
				Message: fmt.Sprintf("成功获取 %d 个活动告警", len(simplifiedAlerts)),
			}

			// 转换为 JSON
			jsonBytes, err := json.MarshalIndent(alertsOut, "", "  ")
			if err != nil {
				config.Error("[Tool-query_prometheus_alerts] JSON 序列化失败: %v", err)
				return "", err
			}

			config.Info("[Tool-query_prometheus_alerts] 查询完成，找到 %d 个告警", len(simplifiedAlerts))
			return string(jsonBytes), nil
		})
	if err != nil {
		config.Error("[Tool-query_prometheus_alerts] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}
	return t
}
