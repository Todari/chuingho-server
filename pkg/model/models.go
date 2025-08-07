package model

import (
	"time"

	"github.com/google/uuid"
)

// User 사용자 엔티티
type User struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Resume 자기소개서 엔티티
type Resume struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Content     string    `json:"content" db:"content"`           // 자기소개서 텍스트 내용
	ContentHash string    `json:"content_hash" db:"content_hash"` // 텍스트 내용 해시
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ResumeStatus 자기소개서 처리 상태
type ResumeStatus string

const (
	ResumeStatusUploaded   ResumeStatus = "uploaded"
	ResumeStatusProcessing ResumeStatus = "processing"
	ResumeStatusCompleted  ResumeStatus = "completed"
	ResumeStatusFailed     ResumeStatus = "failed"
)

// TitleRecommendation 췽호 추천 결과 엔티티
type TitleRecommendation struct {
	ID                       uuid.UUID              `json:"id" db:"id"`
	ResumeID                 uuid.UUID              `json:"resume_id" db:"resume_id"`
	Titles                   []string               `json:"titles" db:"titles"`
	VectorSimilarityScores   map[string]float32     `json:"vector_similarity_scores,omitempty" db:"vector_similarity_scores"`
	ProcessingTimeMs         *int                   `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	MLModelVersion           *string                `json:"ml_model_version,omitempty" db:"ml_model_version"`
	CreatedAt                time.Time              `json:"created_at" db:"created_at"`
}

// ProcessingLog 처리 로그 엔티티
type ProcessingLog struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	RequestID        string     `json:"request_id" db:"request_id"`
	UserIDHash       *string    `json:"user_id_hash,omitempty" db:"user_id_hash"`
	Operation        string     `json:"operation" db:"operation"`
	Status           string     `json:"status" db:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty" db:"error_message"`
	ProcessingTimeMs *int       `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// PhraseCandidate 형용사+명사 후보 엔티티
type PhraseCandidate struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	Phrase           string     `json:"phrase" db:"phrase"`
	Adjective        string     `json:"adjective" db:"adjective"`
	Noun             string     `json:"noun" db:"noun"`
	FrequencyScore   float64    `json:"frequency_score" db:"frequency_score"`
	SemanticCategory *string    `json:"semantic_category,omitempty" db:"semantic_category"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// DTO (Data Transfer Objects)

// UploadResumeRequest 자기소개서 업로드 요청
type UploadResumeRequest struct {
	Text string `json:"text" binding:"required,min=10,max=50000"` // 자기소개서 텍스트 (10자~50000자)
}

// UploadResumeResponse 자기소개서 업로드 응답
type UploadResumeResponse struct {
	ResumeID uuid.UUID    `json:"resumeId"`
	Status   ResumeStatus `json:"status"`
}

// GenerateTitlesRequest 췽호 생성 요청
type GenerateTitlesRequest struct {
	ResumeID uuid.UUID `json:"resumeId" binding:"required"`
}

// GenerateTitlesResponse 췽호 생성 응답
type GenerateTitlesResponse struct {
    Titles     []string               `json:"titles"`
    TopSimilar []CombinationDetail    `json:"top_similar,omitempty"`
}

// MLEmbeddingRequest ML 서비스 임베딩 요청
type MLEmbeddingRequest struct {
	Text string `json:"text"`
}

// MLEmbeddingResponse ML 서비스 임베딩 응답
type MLEmbeddingResponse struct {
	Vector []float32 `json:"vector"`
}

// MLBatchEmbeddingRequest ML 서비스 배치 임베딩 요청
type MLBatchEmbeddingRequest struct {
	Phrases []string `json:"phrases"`
}

// MLBatchEmbeddingResponse ML 서비스 배치 임베딩 응답
type MLBatchEmbeddingResponse struct {
	Results []struct {
		Phrase string    `json:"phrase"`
		Vector []float32 `json:"vector"`
	} `json:"results"`
}

// VectorSearchResult 벡터 검색 결과
type VectorSearchResult struct {
	Phrase string  `json:"phrase"`
	Score  float32 `json:"score"`
}

// 동적 조합 생성 요청/응답 모델
type DynamicCombinationRequest struct {
	ResumeText       string `json:"resume_text" binding:"required"`
	TopK             int    `json:"top_k"`
	AdjFilterCount   int    `json:"adj_filter_count"`
	NounFilterCount  int    `json:"noun_filter_count"`
}

type CombinationDetail struct {
	Phrase     string  `json:"phrase"`
	Similarity float64 `json:"similarity"`
}

type DynamicCombinationResponse struct {
	Combinations        []string            `json:"combinations"`
	Details             []CombinationDetail `json:"details"`
	ProcessingTime      float64             `json:"processing_time"`
	TotalGenerated      int                 `json:"total_generated"`
	FilteredAdjectives  int                 `json:"filtered_adjectives"`
	FilteredNouns       int                 `json:"filtered_nouns"`
    TopSimilar          []CombinationDetail `json:"top_similar,omitempty"`
}

// ErrorResponse API 에러 응답
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// HealthCheckResponse 헬스체크 응답
type HealthCheckResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]interface{} `json:"services,omitempty"`
}