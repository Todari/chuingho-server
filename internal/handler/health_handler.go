package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/database"
	"github.com/Todari/chuingho-server/internal/service"
	"github.com/Todari/chuingho-server/internal/storage"
	"github.com/Todari/chuingho-server/internal/vector"
	"github.com/Todari/chuingho-server/pkg/model"
)

// HealthHandler 헬스체크 관련 HTTP 핸들러
type HealthHandler struct {
	db        *database.DB
	storage   *storage.Storage
	vectorDB  vector.VectorDB
    mlClient  service.MLClientAPI
	logger    *zap.Logger
}

// NewHealthHandler 새로운 헬스체크 핸들러 생성
func NewHealthHandler(
	db *database.DB,
	storage *storage.Storage,
	vectorDB vector.VectorDB,
    mlClient service.MLClientAPI,
	logger *zap.Logger,
) *HealthHandler {
	return &HealthHandler{
		db:       db,
		storage:  storage,
		vectorDB: vectorDB,
		mlClient: mlClient,
		logger:   logger,
	}
}

// HealthCheck 전체 시스템 헬스체크
// @Summary 시스템 헬스체크
// @Description 데이터베이스, 스토리지, 벡터DB, ML서비스의 전체 상태 확인
// @Tags health
// @Produce json
// @Success 200 {object} model.HealthCheckResponse
// @Failure 503 {object} model.ErrorResponse
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := c.GetString("request_id")
	
	h.logger.Debug("헬스체크 시작", zap.String("request_id", requestID))
	
	services := make(map[string]interface{})
	overallStatus := "healthy"

	// 데이터베이스 상태 확인
	if err := h.db.HealthCheck(ctx); err != nil {
		h.logger.Error("데이터베이스 헬스체크 실패", zap.Error(err))
		services["database"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		dbStats := h.db.GetStats()
		services["database"] = map[string]interface{}{
			"status":           "healthy",
			"total_conns":      dbStats.TotalConns(),
			"acquired_conns":   dbStats.AcquiredConns(),
			"idle_conns":       dbStats.IdleConns(),
			"constructed_conns": dbStats.ConstructingConns(),
		}
	}

	// 스토리지 상태 확인
	if err := h.storage.HealthCheck(ctx); err != nil {
		h.logger.Error("스토리지 헬스체크 실패", zap.Error(err))
		services["storage"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		services["storage"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	// 벡터 DB 상태 확인
	if err := h.vectorDB.HealthCheck(ctx); err != nil {
		h.logger.Error("벡터DB 헬스체크 실패", zap.Error(err))
		services["vector_db"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		vectorStats, _ := h.vectorDB.GetStats(ctx)
		services["vector_db"] = map[string]interface{}{
			"status":        "healthy",
			"total_vectors": vectorStats.TotalVectors,
			"dimension":     vectorStats.Dimension,
			"index_type":    vectorStats.IndexType,
		}
	}

	// ML 서비스 상태 확인
	if err := h.mlClient.HealthCheck(ctx); err != nil {
		h.logger.Error("ML 서비스 헬스체크 실패", zap.Error(err))
		services["ml_service"] = map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		overallStatus = "unhealthy"
	} else {
		services["ml_service"] = map[string]interface{}{
			"status": "healthy",
		}
	}

	response := model.HealthCheckResponse{
		Status:    overallStatus,
		Timestamp: time.Now(),
		Services:  services,
	}

	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	h.logger.Info("헬스체크 완료",
		zap.String("request_id", requestID),
		zap.String("status", overallStatus))

	c.JSON(statusCode, response)
}

// ReadinessCheck 준비 상태 확인 (K8s용)
// @Summary 준비 상태 확인
// @Description 서비스가 요청을 받을 준비가 되었는지 확인
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} model.ErrorResponse
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
	ctx := c.Request.Context()
	
	// 핵심 서비스들만 간단히 확인
	if err := h.db.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error: "데이터베이스가 준비되지 않았습니다",
		})
		return
	}

	if err := h.storage.HealthCheck(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{
			Error: "스토리지가 준비되지 않았습니다",
		})
		return
	}

	c.JSON(http.StatusOK, map[string]string{
		"status": "ready",
	})
}

// LivenessCheck 생존 상태 확인 (K8s용)
// @Summary 생존 상태 확인
// @Description 서비스가 살아있는지 확인
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{
		"status": "alive",
	})
}