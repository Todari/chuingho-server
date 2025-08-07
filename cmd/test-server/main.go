package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/pkg/model"
)

// 간단한 인메모리 스토리지 (실제로는 데이터베이스 사용)
var resumeStorage = make(map[string]string)

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
		
		// 텍스트를 스토리지에 저장 (실제로는 데이터베이스에 저장)
		resumeStorage[resumeID.String()] = req.Text
		
		response := model.UploadResumeResponse{
			ResumeID: resumeID,
			Status:   model.ResumeStatusUploaded,
		}

		logger.Info("자기소개서 업로드 성공",
			zap.String("resume_id", resumeID.String()),
			zap.Int("text_length", len([]rune(req.Text))))

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

		// 향상된 칭호 생성 (실제로는 ML 서비스와 벡터 검색 사용)
		// resumeId로 원본 텍스트를 찾아서 스마트한 칭호 생성
		var resumeText string
		if storedText, exists := resumeStorage[req.ResumeID.String()]; exists {
			resumeText = storedText
		} else {
			resumeText = "창의적이고 열정적인 개발자입니다. 팀워크를 중시하며 지속적인 학습과 성장을 추구합니다."
		}
		
		titleGenerator := NewTitleGenerator()
		mockTitles := titleGenerator.GenerateSmartTitles(resumeText, 3)

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