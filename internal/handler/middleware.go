package handler

import (
	"os"
	"strings"
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
    // 환경변수 기반 설정 (없으면 안전한 개발 기본값 사용)
    allowedOriginsEnv := os.Getenv("CHUINGHO_CORS_ALLOWED_ORIGINS")
    allowCredentials := strings.EqualFold(os.Getenv("CHUINGHO_CORS_ALLOW_CREDENTIALS"), "true")
    allowedMethods := os.Getenv("CHUINGHO_CORS_ALLOWED_METHODS")
    allowedHeadersEnv := os.Getenv("CHUINGHO_CORS_ALLOWED_HEADERS")

    if allowedOriginsEnv == "" {
        // 개발 기본 허용 오리진
        allowedOriginsEnv = "http://localhost:3000,http://127.0.0.1:3000"
    }
    if allowedMethods == "" {
        allowedMethods = "GET, POST, PUT, DELETE, OPTIONS"
    }
    if allowedHeadersEnv == "" {
        allowedHeadersEnv = "Origin, Content-Type, Accept, Authorization, X-Request-ID"
    }

    allowedOrigins := map[string]struct{}{}
    wildcard := false
    for _, o := range strings.Split(allowedOriginsEnv, ",") {
        o = strings.TrimSpace(o)
        if o == "*" {
            wildcard = true
        } else if o != "" {
            allowedOrigins[o] = struct{}{}
        }
    }

    return func(c *gin.Context) {
        origin := c.GetHeader("Origin")

        // 기본 헤더
        c.Header("Access-Control-Expose-Headers", "X-Request-ID, X-Process-Time")
        c.Header("Access-Control-Max-Age", "86400")
        c.Header("Access-Control-Allow-Methods", allowedMethods)

        // 요청된 커스텀 헤더를 그대로 허용 (프리플라이트 안정성)
        reqHeaders := c.GetHeader("Access-Control-Request-Headers")
        if reqHeaders != "" {
            c.Header("Access-Control-Allow-Headers", reqHeaders)
            c.Header("Vary", "Origin, Access-Control-Request-Headers, Access-Control-Request-Method")
        } else {
            c.Header("Access-Control-Allow-Headers", allowedHeadersEnv)
            c.Header("Vary", "Origin")
        }

        // 오리진 허용 정책
        if origin != "" {
            if wildcard {
                // 와일드카드 허용 (credentials 사용 시 브라우저가 거부할 수 있음)
                c.Header("Access-Control-Allow-Origin", "*")
            } else if _, ok := allowedOrigins[origin]; ok {
                c.Header("Access-Control-Allow-Origin", origin)
                if allowCredentials {
                    c.Header("Access-Control-Allow-Credentials", "true")
                }
            }
        } else if wildcard {
            c.Header("Access-Control-Allow-Origin", "*")
        }

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