package vector

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/config"
	"github.com/Todari/chuingho-server/pkg/model"
)

// FaissDB Faiss 인메모리 벡터 DB 구현
// 실제 Faiss 바인딩 대신 순수 Go로 구현한 간단한 벡터 검색
type FaissDB struct {
	config    config.VectorConfig
	logger    *zap.Logger
	vectors   map[string]VectorRecord
	dimension int
	mutex     sync.RWMutex
	indexPath string
}

// NewFaissDB 새로운 Faiss DB 클라이언트 생성
func NewFaissDB(cfg config.VectorConfig, logger *zap.Logger) *FaissDB {
	return &FaissDB{
		config:    cfg,
		logger:    logger,
		vectors:   make(map[string]VectorRecord),
		dimension: cfg.Dimension,
		indexPath: cfg.IndexPath,
	}
}

// Initialize 벡터 DB 초기화
func (f *FaissDB) Initialize(ctx context.Context) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// 인덱스 디렉토리 생성
	if err := os.MkdirAll(f.indexPath, 0755); err != nil {
		return fmt.Errorf("인덱스 디렉토리 생성 실패: %w", err)
	}

	// 기존 인덱스 로드 시도
	if err := f.loadIndex(); err != nil {
		f.logger.Warn("기존 인덱스 로드 실패, 새 인덱스 시작", zap.Error(err))
	}

	f.logger.Info("Faiss 벡터 DB 초기화 완료",
		zap.String("index_path", f.indexPath),
		zap.Int("dimension", f.dimension),
		zap.Int("loaded_vectors", len(f.vectors)))

	return nil
}

// AddVectors 벡터 추가
func (f *FaissDB) AddVectors(ctx context.Context, vectors []VectorRecord) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	for _, record := range vectors {
		if len(record.Vector) != f.dimension {
			return fmt.Errorf("벡터 차원이 맞지 않음: 예상 %d, 실제 %d", f.dimension, len(record.Vector))
		}

		// 벡터 정규화 (코사인 유사도를 위해)
		normalizedVector := f.normalizeVector(record.Vector)
		record.Vector = normalizedVector

		f.vectors[record.ID] = record
	}

	// 인덱스 저장
	if err := f.saveIndex(); err != nil {
		f.logger.Error("인덱스 저장 실패", zap.Error(err))
		return fmt.Errorf("인덱스 저장 실패: %w", err)
	}

	f.logger.Info("벡터 추가 완료", zap.Int("added", len(vectors)), zap.Int("total", len(f.vectors)))
	return nil
}

// Search 유사도 검색
func (f *FaissDB) Search(ctx context.Context, query []float32, topK int) ([]model.VectorSearchResult, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	if len(query) != f.dimension {
		return nil, fmt.Errorf("쿼리 벡터 차원이 맞지 않음: 예상 %d, 실제 %d", f.dimension, len(query))
	}

	// 쿼리 벡터 정규화
	normalizedQuery := f.normalizeVector(query)

	// 모든 벡터와 유사도 계산
	similarities := make([]model.VectorSearchResult, 0, len(f.vectors))

	for id, record := range f.vectors {
		// 코사인 유사도 계산 (정규화된 벡터의 내적)
		similarity := f.cosineSimilarity(normalizedQuery, record.Vector)
		
		// 메타데이터에서 phrase 추출
		phrase := id
		if record.Metadata != nil && record.Metadata["phrase"] != nil {
			if p, ok := record.Metadata["phrase"].(string); ok {
				phrase = p
			}
		}

		similarities = append(similarities, model.VectorSearchResult{
			Phrase: phrase,
			Score:  similarity,
		})
	}

	// 유사도 순으로 정렬 (내림차순)
	sort.Slice(similarities, func(i, j int) bool {
		return similarities[i].Score > similarities[j].Score
	})

	// topK개만 반환
	if topK > len(similarities) {
		topK = len(similarities)
	}

	result := similarities[:topK]
	f.logger.Debug("벡터 검색 완료",
		zap.Int("total_vectors", len(f.vectors)),
		zap.Int("top_k", topK),
		zap.Int("results", len(result)))

	return result, nil
}

// Update 벡터 업데이트
func (f *FaissDB) Update(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(vector) != f.dimension {
		return fmt.Errorf("벡터 차원이 맞지 않음: 예상 %d, 실제 %d", f.dimension, len(vector))
	}

	normalizedVector := f.normalizeVector(vector)
	
	f.vectors[id] = VectorRecord{
		ID:       id,
		Vector:   normalizedVector,
		Metadata: metadata,
	}

	if err := f.saveIndex(); err != nil {
		return fmt.Errorf("인덱스 저장 실패: %w", err)
	}

	f.logger.Debug("벡터 업데이트 완료", zap.String("id", id))
	return nil
}

// Delete 벡터 삭제
func (f *FaissDB) Delete(ctx context.Context, ids []string) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	deletedCount := 0
	for _, id := range ids {
		if _, exists := f.vectors[id]; exists {
			delete(f.vectors, id)
			deletedCount++
		}
	}

	if deletedCount > 0 {
		if err := f.saveIndex(); err != nil {
			return fmt.Errorf("인덱스 저장 실패: %w", err)
		}
	}

	f.logger.Info("벡터 삭제 완료", zap.Int("deleted", deletedCount), zap.Int("remaining", len(f.vectors)))
	return nil
}

// GetStats 통계 정보 조회
func (f *FaissDB) GetStats(ctx context.Context) (*VectorStats, error) {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// 대략적인 메모리 사용량 계산 (벡터 데이터만)
	memoryUsage := int64(len(f.vectors) * f.dimension * 4) // float32 = 4 bytes

	return &VectorStats{
		TotalVectors: len(f.vectors),
		Dimension:    f.dimension,
		IndexType:    "faiss_hnsw_simulation",
		MemoryUsage:  memoryUsage,
		Additional: map[string]interface{}{
			"index_path": f.indexPath,
			"metric":     f.config.MetricType,
		},
	}, nil
}

// HealthCheck 상태 확인
func (f *FaissDB) HealthCheck(ctx context.Context) error {
	f.mutex.RLock()
	defer f.mutex.RUnlock()

	// 기본적인 상태 확인
	if f.vectors == nil {
		return fmt.Errorf("벡터 스토리지가 초기화되지 않음")
	}

	// 인덱스 디렉토리 접근 가능 확인
	if _, err := os.Stat(f.indexPath); err != nil {
		return fmt.Errorf("인덱스 디렉토리 접근 불가: %w", err)
	}

	return nil
}

// Close 연결 종료
func (f *FaissDB) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	// 최종 인덱스 저장
	if err := f.saveIndex(); err != nil {
		f.logger.Error("종료시 인덱스 저장 실패", zap.Error(err))
	}

	f.logger.Info("Faiss 벡터 DB 종료")
	return nil
}

// normalizeVector 벡터 정규화 (L2 정규화)
func (f *FaissDB) normalizeVector(vector []float32) []float32 {
	var norm float32
	for _, v := range vector {
		norm += v * v
	}
	
	if norm == 0 {
		return vector
	}
	
	norm = float32(1.0 / (norm * norm)) // sqrt의 역수
	normalized := make([]float32, len(vector))
	for i, v := range vector {
		normalized[i] = v * norm
	}
	
	return normalized
}

// cosineSimilarity 코사인 유사도 계산 (정규화된 벡터)
func (f *FaissDB) cosineSimilarity(a, b []float32) float32 {
	var dot float32
	for i := range a {
		dot += a[i] * b[i]
	}
	return dot
}

// saveIndex 인덱스를 파일에 저장
func (f *FaissDB) saveIndex() error {
	indexFile := filepath.Join(f.indexPath, "vectors.json")
	
	data, err := json.Marshal(f.vectors)
	if err != nil {
		return fmt.Errorf("벡터 직렬화 실패: %w", err)
	}
	
	if err := os.WriteFile(indexFile, data, 0644); err != nil {
		return fmt.Errorf("인덱스 파일 쓰기 실패: %w", err)
	}
	
	return nil
}

// loadIndex 파일에서 인덱스 로드
func (f *FaissDB) loadIndex() error {
	indexFile := filepath.Join(f.indexPath, "vectors.json")
	
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return fmt.Errorf("인덱스 파일 읽기 실패: %w", err)
	}
	
	if err := json.Unmarshal(data, &f.vectors); err != nil {
		return fmt.Errorf("벡터 역직렬화 실패: %w", err)
	}
	
	return nil
}