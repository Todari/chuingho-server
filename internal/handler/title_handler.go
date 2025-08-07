package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/service"
	"github.com/Todari/chuingho-server/pkg/model"
)

// TitleHandler 췽호 관련 HTTP 핸들러
type TitleHandler struct {
	titleService *service.TitleService
	logger       *zap.Logger
}

// NewTitleHandler 새로운 췽호 핸들러 생성
func NewTitleHandler(titleService *service.TitleService, logger *zap.Logger) *TitleHandler {
	return &TitleHandler{
		titleService: titleService,
		logger:       logger,
	}
}

// GenerateTitles 췽호 생성
// @Summary 췽호 추천 생성
// @Description 자기소개서를 분석하여 3개의 췽호(형용사+명사) 추천
// @Tags titles
// @Accept json
// @Produce json
// @Param request body model.GenerateTitlesRequest true "자기소개서 ID"
// @Success 200 {object} model.GenerateTitlesResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/titles [post]
func (h *TitleHandler) GenerateTitles(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	var req model.GenerateTitlesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("잘못된 요청 본문",
			zap.String("request_id", requestID),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error:   "잘못된 요청 형식입니다",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info("췽호 생성 요청",
		zap.String("request_id", requestID),
		zap.String("resume_id", req.ResumeID.String()))

	// 서비스 호출
	response, err := h.titleService.GenerateTitles(c.Request.Context(), req.ResumeID)
	if err != nil {
		h.logger.Error("췽호 생성 실패",
			zap.String("request_id", requestID),
			zap.String("resume_id", req.ResumeID.String()),
			zap.Error(err))
		
		// 에러 타입에 따른 상태 코드 결정
		statusCode := http.StatusInternalServerError
		errorMessage := err.Error()
		
		switch {
		case errorMessage == "자기소개서를 찾을 수 없습니다":
			statusCode = http.StatusNotFound
		case errorMessage == "자기소개서 내용이 너무 짧습니다":
			statusCode = http.StatusBadRequest
		}
		
		c.JSON(statusCode, model.ErrorResponse{
			Error:   "췽호 생성에 실패했습니다",
			Details: errorMessage,
		})
		return
	}

	h.logger.Info("췽호 생성 성공",
		zap.String("request_id", requestID),
		zap.String("resume_id", req.ResumeID.String()),
		zap.Strings("titles", response.Titles))

	c.JSON(http.StatusOK, response)
}

// GetTitleHistory 췽호 추천 기록 조회
// @Summary 췽호 추천 기록 조회
// @Description 특정 자기소개서의 췽호 추천 기록을 시간순으로 조회
// @Tags titles
// @Produce json
// @Param resumeId path string true "자기소개서 ID (UUID)"
// @Success 200 {array} model.TitleRecommendation
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/titles/history/{resumeId} [get]
func (h *TitleHandler) GetTitleHistory(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	// 경로 파라미터 파싱
	resumeIDStr := c.Param("resumeId")
	resumeID, err := uuid.Parse(resumeIDStr)
	if err != nil {
		h.logger.Warn("잘못된 자기소개서 ID",
			zap.String("request_id", requestID),
			zap.String("resume_id", resumeIDStr))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "잘못된 자기소개서 ID 형식입니다",
			Code:  "INVALID_RESUME_ID",
		})
		return
	}

	// 서비스 호출
	history, err := h.titleService.GetTitleHistory(c.Request.Context(), resumeID)
	if err != nil {
		h.logger.Error("췽호 기록 조회 실패",
			zap.String("request_id", requestID),
			zap.String("resume_id", resumeID.String()),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "췽호 기록 조회에 실패했습니다",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info("췽호 기록 조회 완료",
		zap.String("request_id", requestID),
		zap.String("resume_id", resumeID.String()),
		zap.Int("count", len(history)))

	c.JSON(http.StatusOK, history)
}