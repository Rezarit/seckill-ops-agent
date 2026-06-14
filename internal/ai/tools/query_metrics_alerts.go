// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	AlertName   string            `json:"alert_name" jsonschema:"description=告警名称，从 Prometheus 告警的 labels.alertname 字段提取"`
	Description string            `json:"description" jsonschema:"description=告警描述信息，从 Prometheus 告警的 annotations.description 字段提取"`
	State       string            `json:"state" jsonschema:"description=告警状态，通常为 'firing'（触发中）或 'pending'（待触发）"`
	ActiveAt    string            `json:"active_at" jsonschema:"description=告警激活时间，RFC3339 格式的时间戳，例如 '2025-10-29T08:48:42.496134755Z'"`
	Duration    string            `json:"duration" jsonschema:"description=告警持续时间，从激活时间到当前时间的时长，格式如 '2h30m15s'、'30m15s' 或 '15s'"`
	Labels      map[string]string `json:"labels,omitempty" jsonschema:"description=告警标签，包含实例、严重级别等元数据"`
	Annotations map[string]string `json:"annotations,omitempty" jsonschema:"description=告警注释，包含详细描述信息"`
	Value       string            `json:"value,omitempty" jsonschema:"description=告警触发时的指标值"`
}

// PrometheusAlertsOutput 告警查询输出结构
type PrometheusAlertsOutput struct {
	Success          bool              `json:"success" jsonschema:"description=查询是否成功"`
	Alerts           []SimplifiedAlert `json:"alerts,omitempty" jsonschema:"description=活动告警列表，每个告警包含名称、描述、状态、激活时间和持续时间。相同 alertname 的告警只保留第一个"`
	Message          string            `json:"message,omitempty" jsonschema:"description=操作结果的状态消息"`
	Error            string            `json:"error,omitempty" jsonschema:"description=如果查询失败，包含错误信息"`
	PrometheusStatus string            `json:"prometheus_status,omitempty" jsonschema:"description=Prometheus 服务状态: 'healthy'、'unhealthy' 或 'unknown'"`
	AlertCount       int               `json:"alert_count" jsonschema:"description=告警总数"`
	FiringCount      int               `json:"firing_count" jsonschema:"description=触发中的告警数量"`
}

// queryPrometheusAlerts 查询 Prometheus 告警
// 从配置的 Prometheus 服务获取活动告警信息
func queryPrometheusAlerts() (PrometheusAlertsResult, error) {
	// 检查配置
	promConfig := config.AppConfig.Prometheus
	if !promConfig.Enabled {
		config.Warn("[Tool-query_prometheus_alerts] Prometheus 未启用")
		return PrometheusAlertsResult{}, nil
	}

	if promConfig.URL == "" {
		config.Warn("[Tool-query_prometheus_alerts] Prometheus 地址未配置")
		return PrometheusAlertsResult{}, nil
	}

	// 构建 API URL
	apiURL := fmt.Sprintf("%s/api/v1/alerts", promConfig.URL)
	config.Info("[Tool-query_prometheus_alerts] 查询 Prometheus 告警: %s", apiURL)

	// 设置超时时间
	timeout := time.Duration(promConfig.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// 创建 HTTP 客户端
	client := &http.Client{Timeout: timeout}

	// 发送请求
	resp, err := client.Get(apiURL)
	if err != nil {
		config.Error("[Tool-query_prometheus_alerts] HTTP 请求失败: %v", err)
		return PrometheusAlertsResult{}, fmt.Errorf("HTTP 请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		config.Error("[Tool-query_prometheus_alerts] 读取响应失败: %v", err)
		return PrometheusAlertsResult{}, fmt.Errorf("读取响应失败: %v", err)
	}

	// 解析 JSON
	var result PrometheusAlertsResult
	if err := json.Unmarshal(body, &result); err != nil {
		config.Error("[Tool-query_prometheus_alerts] JSON 解析失败: %v", err)
		return PrometheusAlertsResult{}, fmt.Errorf("JSON 解析失败: %v", err)
	}

	// 检查错误
	if result.Status == "error" {
		config.Error("[Tool-query_prometheus_alerts] Prometheus API 错误: %s", result.Error)
		return result, fmt.Errorf("Prometheus API 错误: %s", result.Error)
	}

	config.Info("[Tool-query_prometheus_alerts] 查询成功，获取到 %d 个告警", len(result.Data.Alerts))
	return result, nil
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

// checkPrometheusHealth 检查 Prometheus 服务健康状态
func checkPrometheusHealth() string {
	promConfig := config.AppConfig.Prometheus
	if !promConfig.Enabled || promConfig.URL == "" {
		return "disabled"
	}

	healthURL := fmt.Sprintf("%s/-/healthy", promConfig.URL)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(healthURL)
	if err != nil {
		return "unhealthy"
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return "healthy"
	}
	return "unhealthy"
}

// NewPrometheusAlertsQueryTool 创建 Prometheus 告警查询工具
// 查询 Prometheus 告警系统中的活动告警
func NewPrometheusAlertsQueryTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"query_prometheus_alerts",
		"Query active alerts from Prometheus alerting system. This tool retrieves all currently active/firing alerts including their labels, annotations, state, and values.",
		func(ctx context.Context, input *struct{}, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-query_prometheus_alerts] 工具调用开始")

			// 检查 Prometheus 健康状态
			promStatus := checkPrometheusHealth()
			config.Info("[Tool-query_prometheus_alerts] Prometheus 状态: %s", promStatus)

			// 调用 Prometheus Alerts API
			result, err := queryPrometheusAlerts()
			if err != nil {
				alertsOut := PrometheusAlertsOutput{
					Success:          false,
					Error:            err.Error(),
					Message:          "查询 Prometheus 告警失败",
					PrometheusStatus: promStatus,
					AlertCount:       0,
					FiringCount:      0,
				}
				jsonBytes, _ := json.MarshalIndent(alertsOut, "", "  ")
				config.Error("[Tool-query_prometheus_alerts] 查询失败: %v", err)
				return string(jsonBytes), err
			}

			// 转换为简化格式（相同 alertname 只保留第一个）
			seenAlertNames := make(map[string]bool)
			simplifiedAlerts := make([]SimplifiedAlert, 0)
			firingCount := 0

			for _, alert := range result.Data.Alerts {
				alertName := alert.Labels["alertname"]
				if seenAlertNames[alertName] {
					continue
				}
				seenAlertNames[alertName] = true

				// 统计触发中的告警
				if alert.State == "firing" {
					firingCount++
				}

				simplified := SimplifiedAlert{
					AlertName:   alertName,
					Description: alert.Annotations["description"],
					State:       alert.State,
					ActiveAt:    alert.ActiveAt,
					Duration:    calculateDuration(alert.ActiveAt),
					Labels:      alert.Labels,
					Annotations: alert.Annotations,
					Value:       alert.Value,
				}
				simplifiedAlerts = append(simplifiedAlerts, simplified)
			}

			// 构建成功响应
			alertsOut := PrometheusAlertsOutput{
				Success:          true,
				Alerts:           simplifiedAlerts,
				Message:          fmt.Sprintf("成功获取 %d 个告警，其中 %d 个正在触发", len(simplifiedAlerts), firingCount),
				PrometheusStatus: promStatus,
				AlertCount:       len(simplifiedAlerts),
				FiringCount:      firingCount,
			}

			// 转换为 JSON
			jsonBytes, err := json.MarshalIndent(alertsOut, "", "  ")
			if err != nil {
				config.Error("[Tool-query_prometheus_alerts] JSON 序列化失败: %v", err)
				return "", err
			}

			config.Info("[Tool-query_prometheus_alerts] 查询完成，找到 %d 个告警，%d 个正在触发", len(simplifiedAlerts), firingCount)
			return string(jsonBytes), nil
		})
	if err != nil {
		config.Error("[Tool-query_prometheus_alerts] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}
	return t
}
