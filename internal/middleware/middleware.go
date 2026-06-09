package middleware

import (
	"net/http"

	"SuperBizAgent/config"

	"github.com/gin-gonic/gin"
)

// CORS 中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// Logger 中间件 - 使用 Zap 记录日志
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		config.Info("[HTTP] %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
	}
}

// Recovery 中间件 - 捕获 panic 并记录日志
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				config.Error("[PANIC] %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "Internal server error",
					"error":   err,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
