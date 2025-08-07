package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/pkg/util"
)

// RequestLogger 요청 로깅 미들웨어
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if recovered != nil {
			logger.Error("패닉 복구",
				zap.Any("error", recovered),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method))
			c.AbortWithStatus(500)
		}
	})
}

// RequestID 요청 ID 생성 미들웨어
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			if id, err := util.GenerateRequestID(); err == nil {
				requestID = id
			} else {
				requestID = "unknown"
			}
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// CORS CORS 헤더 설정 미들웨어
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Process-Time")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ProcessTime 처리 시간 측정 미들웨어
func ProcessTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		processTime := time.Since(start)
		c.Header("X-Process-Time", processTime.String())
	}
}

// SecurityHeaders 보안 헤더 설정 미들웨어
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}