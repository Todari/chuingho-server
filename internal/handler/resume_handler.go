package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/service"
	"github.com/Todari/chuingho-server/pkg/model"
)

// ResumeHandler 자기소개서 관련 HTTP 핸들러
type ResumeHandler struct {
	resumeService *service.ResumeService
	logger        *zap.Logger
}

// NewResumeHandler 새로운 자기소개서 핸들러 생성
func NewResumeHandler(resumeService *service.ResumeService, logger *zap.Logger) *ResumeHandler {
	return &ResumeHandler{
		resumeService: resumeService,
		logger:        logger,
	}
}

// UploadResume 자기소개서 업로드
// @Summary 자기소개서 텍스트 업로드
// @Description 자기소개서 텍스트를 JSON으로 전송하여 등록
// @Tags resumes
// @Accept json
// @Produce json
// @Param request body model.UploadResumeRequest true "자기소개서 텍스트"
// @Success 200 {object} model.UploadResumeResponse
// @Failure 400 {object} model.ErrorResponse
// @Failure 413 {object} model.ErrorResponse "텍스트 길이 초과"
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/resumes [post]
func (h *ResumeHandler) UploadResume(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	// JSON 요청 바인딩
	var req model.UploadResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("잘못된 요청 데이터",
			zap.String("request_id", requestID),
			zap.Error(err))
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "올바른 JSON 형식과 텍스트가 필요합니다",
			Code:  "INVALID_REQUEST",
			Details: err.Error(),
		})
		return
	}

	// 텍스트 길이 검증 (추가 검증)
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

	h.logger.Info("자기소개서 업로드 요청",
		zap.String("request_id", requestID),
		zap.Int("text_length", textLength))

	// 서비스 호출
	response, err := h.resumeService.UploadResume(
		c.Request.Context(),
		req.Text,
	)
	if err != nil {
		h.logger.Error("자기소개서 업로드 실패",
			zap.String("request_id", requestID),
			zap.Error(err))
		
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "자기소개서 업로드에 실패했습니다",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info("자기소개서 업로드 성공",
		zap.String("request_id", requestID),
		zap.String("resume_id", response.ResumeID.String()))

	c.JSON(http.StatusOK, response)
}

// GetResume 자기소개서 조회
// @Summary 자기소개서 메타데이터 조회
// @Description 자기소개서 ID로 메타데이터 조회
// @Tags resumes
// @Produce json
// @Param id path string true "자기소개서 ID (UUID)"
// @Success 200 {object} model.Resume
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/resumes/{id} [get]
func (h *ResumeHandler) GetResume(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	// 경로 파라미터 파싱
	resumeIDStr := c.Param("id")
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
	resume, err := h.resumeService.GetResume(c.Request.Context(), resumeID)
	if err != nil {
		h.logger.Error("자기소개서 조회 실패",
			zap.String("request_id", requestID),
			zap.String("resume_id", resumeID.String()),
			zap.Error(err))
		
		statusCode := http.StatusInternalServerError
		if err.Error() == "자기소개서를 찾을 수 없습니다" {
			statusCode = http.StatusNotFound
		}
		
		c.JSON(statusCode, model.ErrorResponse{
			Error:   "자기소개서 조회에 실패했습니다",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resume)
}

// ListResumes 자기소개서 목록 조회 (관리용)
// @Summary 자기소개서 목록 조회
// @Description 전체 자기소개서 목록을 페이지네이션하여 조회 (관리자용)
// @Tags resumes
// @Produce json
// @Param limit query int false "조회할 개수 (기본: 20, 최대: 100)"
// @Param offset query int false "건너뛸 개수 (기본: 0)"
// @Success 200 {array} model.Resume
// @Failure 400 {object} model.ErrorResponse
// @Failure 500 {object} model.ErrorResponse
// @Router /v1/resumes [get]
func (h *ResumeHandler) ListResumes(c *gin.Context) {
	requestID := c.GetString("request_id")
	
	// 쿼리 파라미터 파싱
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "잘못된 limit 값입니다 (1-100)",
			Code:  "INVALID_LIMIT",
		})
		return
	}
	
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{
			Error: "잘못된 offset 값입니다 (≥0)",
			Code:  "INVALID_OFFSET",
		})
		return
	}

	// 서비스 호출
	resumes, err := h.resumeService.ListResumes(c.Request.Context(), limit, offset)
	if err != nil {
		h.logger.Error("자기소개서 목록 조회 실패",
			zap.String("request_id", requestID),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{
			Error:   "자기소개서 목록 조회에 실패했습니다",
			Details: err.Error(),
		})
		return
	}

	h.logger.Info("자기소개서 목록 조회 완료",
		zap.String("request_id", requestID),
		zap.Int("count", len(resumes)),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	c.JSON(http.StatusOK, resumes)
}