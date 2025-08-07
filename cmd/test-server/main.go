package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/pkg/model"
)

func main() {
	// 로거 초기화
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Gin 라우터 생성
	router := gin.Default()

	// 로깅 미들웨어
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health"},
	}))

	// 헬스체크
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now(),
			"service":   "chuingho-test-server",
		})
	})

	// 자기소개서 업로드 (새로운 텍스트 기반 API)
	router.POST("/v1/resumes", func(c *gin.Context) {
		var req model.UploadResumeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("잘못된 요청 데이터", zap.Error(err))
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "올바른 JSON 형식과 텍스트가 필요합니다",
				Code:    "INVALID_REQUEST",
				Details: err.Error(),
			})
			return
		}

		// 텍스트 길이 검증
		textLength := len([]rune(req.Text))
		if textLength < 10 {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error: "자기소개서는 최소 10자 이상이어야 합니다",
				Code:  "TEXT_TOO_SHORT",
			})
			return
		}
		if textLength > 50000 {
			c.JSON(http.StatusRequestEntityTooLarge, model.ErrorResponse{
				Error: "자기소개서는 최대 50,000자까지 입력 가능합니다",
				Code:  "TEXT_TOO_LONG",
			})
			return
		}

		logger.Info("자기소개서 업로드 요청",
			zap.Int("text_length", textLength))

		// Mock 응답 생성
		resumeID := uuid.New()
		response := model.UploadResumeResponse{
			ResumeID: resumeID,
			Status:   model.ResumeStatusUploaded,
		}

		logger.Info("자기소개서 업로드 성공",
			zap.String("resume_id", resumeID.String()))

		c.JSON(http.StatusOK, response)
	})

	// 췽호 추천 생성
	router.POST("/v1/titles", func(c *gin.Context) {
		var req model.GenerateTitlesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{
				Error:   "resumeId가 필요합니다",
				Code:    "INVALID_REQUEST",
				Details: err.Error(),
			})
			return
		}

		logger.Info("췽호 생성 요청",
			zap.String("resume_id", req.ResumeID.String()))

		// Mock 췽호 추천 (실제로는 ML 서비스와 벡터 검색 사용)
		mockTitles := []string{
			"창의적 설계자",
			"열정적 개발자", 
			"협력적 리더",
		}

		response := model.GenerateTitlesResponse{
			Titles: mockTitles,
		}

		logger.Info("췽호 생성 완료",
			zap.String("resume_id", req.ResumeID.String()),
			zap.Strings("titles", mockTitles))

		c.JSON(http.StatusOK, response)
	})

	// 서버 시작
	logger.Info("테스트 서버 시작", zap.String("port", "8080"))
	if err := router.Run(":8080"); err != nil {
		logger.Fatal("서버 시작 실패", zap.Error(err))
	}
}