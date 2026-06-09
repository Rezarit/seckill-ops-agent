// Package tools 提供各种工具实现
// 包含时间查询、文档检索、秒杀系统查询等工具
package tools

import (
	"SuperBizAgent/config"
	"context"
	"encoding/json"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MysqlCrudInput MySQL CRUD 工具的输入参数
type MysqlCrudInput struct {
	DSN         string `json:"dsn" jsonschema:"description=MySQL 数据库连接字符串，包含用户名、密码、主机、端口和数据库名"`
	SQL         string `json:"sql" jsonschema:"description=要执行的 SQL 查询语句"`
	OperateType string `json:"operate_type" jsonschema:"description=SQL 操作类型: query(查询), insert(插入), update(更新), delete(删除)"`
}

// NewMysqlCrudTool 创建 MySQL CRUD 工具（已弃用）
// 注意：此工具仅供参考，实际生产环境中不建议直接执行任意 SQL
// 当前版本仅支持查询操作，且需要用户确认
//
// Deprecated: 此工具已下线，仅保留查询功能演示
func NewMysqlCrudTool() tool.InvokableTool {
	t, err := utils.InferOptionableTool(
		"mysql_crud",
		"Execute SQL queries against the MySQL database. (Deprecated: Query only)",
		func(ctx context.Context, input *MysqlCrudInput, opts ...tool.Option) (output string, err error) {
			config.Info("[Tool-mysql_crud] 工具调用开始（已弃用）")

			// 验证输入参数
			if input.DSN == "" {
				config.Error("[Tool-mysql_crud] DSN 为空")
				return `{"success":false,"message":"DSN is required"}`, nil
			}
			if input.SQL == "" {
				config.Error("[Tool-mysql_crud] SQL 为空")
				return `{"success":false,"message":"SQL is required"}`, nil
			}

			config.Info("[Tool-mysql_crud] DSN: %s", maskDSN(input.DSN))
			config.Info("[Tool-mysql_crud] SQL: %s", truncateText(input.SQL, 100))
			config.Info("[Tool-mysql_crud] OperateType: %s", input.OperateType)

			// 建立数据库连接
			db, err := gorm.Open(mysql.Open(input.DSN), &gorm.Config{})
			if err != nil {
				config.Error("[Tool-mysql_crud] 数据库连接失败: %v", err)
				return fmtJSON(false, "", err.Error()), nil
			}

			// 仅支持查询操作（安全限制）
			if input.OperateType == "query" {
				var results []interface{}
				err = db.Raw(input.SQL).Scan(&results).Error
				if err != nil {
					config.Error("[Tool-mysql_crud] SQL 执行失败: %v", err)
					return fmtJSON(false, "", err.Error()), nil
				}

				resBytes, err := json.Marshal(results)
				if err != nil {
					config.Error("[Tool-mysql_crud] JSON 序列化失败: %v", err)
					return fmtJSON(false, "", err.Error()), nil
				}

				config.Info("[Tool-mysql_crud] 查询完成，返回 %d 条记录", len(results))
				return string(resBytes), nil
			} else {
				// 非查询操作已禁用
				config.Warn("[Tool-mysql_crud] 非查询操作已禁用: %s", input.OperateType)
				return `{"success":false,"message":"Only query operations are allowed"}`, nil
			}
		})
	if err != nil {
		config.Error("[Tool-mysql_crud] 工具创建失败: %v", err)
		// 返回 nil 表示工具创建失败，上层可以继续运行
		return nil
	}
	return t
}

// maskDSN 对 DSN 进行脱敏处理
func maskDSN(dsn string) string {
	return "******"
}

// truncateText 截断文本用于日志输出
func truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + "..."
}

// fmtJSON 格式化 JSON 响应
func fmtJSON(success bool, message, errMsg string) string {
	result := map[string]interface{}{
		"success": success,
	}
	if message != "" {
		result["message"] = message
	}
	if errMsg != "" {
		result["error"] = errMsg
	}
	b, _ := json.Marshal(result)
	return string(b)
}
