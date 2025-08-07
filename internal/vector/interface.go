package vector

import (
	"context"

	"github.com/Todari/chuingho-server/pkg/model"
)

// VectorDB 벡터 데이터베이스 인터페이스
type VectorDB interface {
	// Initialize 벡터 DB 초기화
	Initialize(ctx context.Context) error

	// AddVectors 벡터 추가
	AddVectors(ctx context.Context, vectors []VectorRecord) error

	// Search 유사도 검색
	Search(ctx context.Context, query []float32, topK int) ([]model.VectorSearchResult, error)

	// Update 벡터 업데이트
	Update(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error

	// Delete 벡터 삭제
	Delete(ctx context.Context, ids []string) error

	// GetStats 통계 정보 조회
	GetStats(ctx context.Context) (*VectorStats, error)

	// HealthCheck 상태 확인
	HealthCheck(ctx context.Context) error

	// Close 연결 종료
	Close() error
}

// VectorRecord 벡터 레코드
type VectorRecord struct {
	ID       string                 `json:"id"`
	Vector   []float32              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// VectorStats 벡터 DB 통계
type VectorStats struct {
	TotalVectors int                    `json:"total_vectors"`
	Dimension    int                    `json:"dimension"`
	IndexType    string                 `json:"index_type"`
	MemoryUsage  int64                  `json:"memory_usage_bytes"`
	Additional   map[string]interface{} `json:"additional,omitempty"`
}