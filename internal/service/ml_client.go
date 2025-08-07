package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/config"
	"github.com/Todari/chuingho-server/pkg/model"
)

// MLClient ML 서비스 클라이언트
type MLClient struct {
	config     config.MLConfig
	httpClient *http.Client
	logger     *zap.Logger
}

// NewMLClient 새로운 ML 클라이언트 생성
func NewMLClient(cfg config.MLConfig, logger *zap.Logger) *MLClient {
	return &MLClient{
		config: cfg,
		httpClient: &http.Client{
			Timeout: time.Duration(cfg.Timeout) * time.Second,
		},
		logger: logger,
	}
}

// GetEmbedding 단일 텍스트 임베딩 생성
func (c *MLClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	requestBody := model.MLEmbeddingRequest{
		Text: text,
	}

	response, err := c.makeRequest(ctx, "/embed", requestBody)
	if err != nil {
		return nil, fmt.Errorf("임베딩 요청 실패: %w", err)
	}

	var embeddingResponse model.MLEmbeddingResponse
	if err := json.Unmarshal(response, &embeddingResponse); err != nil {
		return nil, fmt.Errorf("임베딩 응답 파싱 실패: %w", err)
	}

	c.logger.Debug("임베딩 생성 완료",
		zap.Int("text_length", len(text)),
		zap.Int("vector_dimension", len(embeddingResponse.Vector)))

	return embeddingResponse.Vector, nil
}

// GetBatchEmbeddings 배치 텍스트 임베딩 생성
func (c *MLClient) GetBatchEmbeddings(ctx context.Context, phrases []string) (map[string][]float32, error) {
	requestBody := model.MLBatchEmbeddingRequest{
		Phrases: phrases,
	}

	response, err := c.makeRequest(ctx, "/embed/phrases", requestBody)
	if err != nil {
		return nil, fmt.Errorf("배치 임베딩 요청 실패: %w", err)
	}

	var batchResponse model.MLBatchEmbeddingResponse
	if err := json.Unmarshal(response, &batchResponse); err != nil {
		return nil, fmt.Errorf("배치 임베딩 응답 파싱 실패: %w", err)
	}

	// 결과를 맵으로 변환
	result := make(map[string][]float32)
	for _, item := range batchResponse.Results {
		result[item.Phrase] = item.Vector
	}

	c.logger.Debug("배치 임베딩 생성 완료",
		zap.Int("input_phrases", len(phrases)),
		zap.Int("output_embeddings", len(result)))

	return result, nil
}

// GenerateDynamicCombinations 동적 형용사+명사 조합 생성
func (c *MLClient) GenerateDynamicCombinations(ctx context.Context, resumeText string, topK int) (*model.DynamicCombinationResponse, error) {
	requestBody := model.DynamicCombinationRequest{
		ResumeText:        resumeText,
		TopK:              topK,
		AdjFilterCount:    20, // 상위 형용사 20개
		NounFilterCount:   30, // 상위 명사 30개
	}

	response, err := c.makeRequest(ctx, "/generate_dynamic_combinations", requestBody)
	if err != nil {
		return nil, fmt.Errorf("동적 조합 생성 요청 실패: %w", err)
	}

	var combinationResponse model.DynamicCombinationResponse
	if err := json.Unmarshal(response, &combinationResponse); err != nil {
		return nil, fmt.Errorf("동적 조합 응답 파싱 실패: %w", err)
	}

	c.logger.Info("동적 조합 생성 완료",
		zap.Int("combinations_count", len(combinationResponse.Combinations)),
		zap.Int("total_generated", combinationResponse.TotalGenerated),
		zap.Float64("processing_time", combinationResponse.ProcessingTime))

	return &combinationResponse, nil
}

// HealthCheck ML 서비스 상태 확인
func (c *MLClient) HealthCheck(ctx context.Context) error {
	url := c.config.ServiceURL + "/health"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("헬스체크 요청 생성 실패: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("헬스체크 요청 실패: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ML 서비스 비정상 상태: %d", resp.StatusCode)
	}

	return nil
}

// makeRequest HTTP 요청 실행
func (c *MLClient) makeRequest(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
	url := c.config.ServiceURL + endpoint

	var requestBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("요청 본문 직렬화 실패: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonData)
	}

	var lastErr error
	for attempt := 0; attempt < c.config.RetryCount; attempt++ {
		if attempt > 0 {
			c.logger.Debug("ML 서비스 요청 재시도",
				zap.String("endpoint", endpoint),
				zap.Int("attempt", attempt+1))
			
			// 재시도 전 잠깐 대기
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(attempt) * time.Second):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, requestBody)
		if err != nil {
			lastErr = fmt.Errorf("요청 생성 실패: %w", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("HTTP 요청 실패: %w", err)
			continue
		}

		responseBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("응답 읽기 실패: %w", err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return responseBody, nil
		}

		// 4xx 오류는 재시도하지 않음
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return nil, fmt.Errorf("클라이언트 오류 %d: %s", resp.StatusCode, string(responseBody))
		}

		// 5xx 오류는 재시도
		lastErr = fmt.Errorf("서버 오류 %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil, fmt.Errorf("최대 재시도 횟수 초과 (%d회): %w", c.config.RetryCount, lastErr)
}