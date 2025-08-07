package vector

import (
	"context"
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/config"
)

func TestFaissDB_Initialize(t *testing.T) {
	// 임시 디렉토리 생성
	tempDir, err := os.MkdirTemp("", "faiss_test")
	if err != nil {
		t.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.VectorConfig{
		Type:      "faiss",
		IndexPath: tempDir,
		Dimension: 768,
	}

	logger, _ := zap.NewDevelopment()
	db := NewFaissDB(cfg, logger)

	ctx := context.Background()
	err = db.Initialize(ctx)
	if err != nil {
		t.Errorf("Initialize() 에러 = %v", err)
	}

	// 인덱스 디렉토리가 생성되었는지 확인
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("인덱스 디렉토리가 생성되지 않음")
	}
}

func TestFaissDB_AddVectors(t *testing.T) {
	db := setupTestFaissDB(t)
	ctx := context.Background()

	vectors := []VectorRecord{
		{
			ID:     "test1",
			Vector: []float32{1.0, 0.5, 0.2, 0.1},
			Metadata: map[string]interface{}{
				"phrase": "창의적 설계자",
			},
		},
		{
			ID:     "test2",
			Vector: []float32{0.8, 0.6, 0.3, 0.2},
			Metadata: map[string]interface{}{
				"phrase": "세심한 분석가",
			},
		},
	}

	err := db.AddVectors(ctx, vectors)
	if err != nil {
		t.Errorf("AddVectors() 에러 = %v", err)
	}

	// 벡터가 저장되었는지 확인
	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() 에러 = %v", err)
	}

	if stats.TotalVectors != 2 {
		t.Errorf("예상 벡터 수 = 2, 실제 = %d", stats.TotalVectors)
	}
}

func TestFaissDB_AddVectors_DimensionMismatch(t *testing.T) {
	db := setupTestFaissDB(t)
	ctx := context.Background()

	vectors := []VectorRecord{
		{
			ID:     "test1",
			Vector: []float32{1.0, 0.5}, // 차원이 맞지 않음 (4 대신 2)
		},
	}

	err := db.AddVectors(ctx, vectors)
	if err == nil {
		t.Error("차원 불일치 시 에러가 반환되어야 함")
	}
}

func TestFaissDB_Search(t *testing.T) {
	db := setupTestFaissDB(t)
	ctx := context.Background()

	// 테스트 벡터 추가
	vectors := []VectorRecord{
		{
			ID:     "test1",
			Vector: []float32{1.0, 0.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "창의적 설계자",
			},
		},
		{
			ID:     "test2",
			Vector: []float32{0.0, 1.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "세심한 분석가",
			},
		},
		{
			ID:     "test3",
			Vector: []float32{0.0, 0.0, 1.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "적극적 리더",
			},
		},
	}

	err := db.AddVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("벡터 추가 실패: %v", err)
	}

	// 첫 번째 벡터와 유사한 쿼리
	queryVector := []float32{0.9, 0.1, 0.0, 0.0}
	results, err := db.Search(ctx, queryVector, 2)
	if err != nil {
		t.Errorf("Search() 에러 = %v", err)
	}

	if len(results) != 2 {
		t.Errorf("예상 결과 수 = 2, 실제 = %d", len(results))
	}

	// 첫 번째 결과가 가장 유사해야 함 (test1)
	if results[0].Phrase != "창의적 설계자" {
		t.Errorf("첫 번째 결과 = %s, 예상 = 창의적 설계자", results[0].Phrase)
	}

	// 유사도가 내림차순으로 정렬되어야 함
	if len(results) > 1 && results[0].Score < results[1].Score {
		t.Error("검색 결과가 유사도 순으로 정렬되지 않음")
	}
}

func TestFaissDB_Update(t *testing.T) {
	db := setupTestFaissDB(t)
	ctx := context.Background()

	// 초기 벡터 추가
	vectors := []VectorRecord{
		{
			ID:     "test1",
			Vector: []float32{1.0, 0.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "원래 구문",
			},
		},
	}

	err := db.AddVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("벡터 추가 실패: %v", err)
	}

	// 벡터 업데이트
	newVector := []float32{0.0, 1.0, 0.0, 0.0}
	newMetadata := map[string]interface{}{
		"phrase": "업데이트된 구문",
	}

	err = db.Update(ctx, "test1", newVector, newMetadata)
	if err != nil {
		t.Errorf("Update() 에러 = %v", err)
	}

	// 업데이트된 벡터로 검색
	queryVector := []float32{0.0, 0.9, 0.0, 0.0}
	results, err := db.Search(ctx, queryVector, 1)
	if err != nil {
		t.Errorf("Search() 에러 = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("예상 결과 수 = 1, 실제 = %d", len(results))
	}

	if results[0].Phrase != "업데이트된 구문" {
		t.Errorf("업데이트 결과 = %s, 예상 = 업데이트된 구문", results[0].Phrase)
	}
}

func TestFaissDB_Delete(t *testing.T) {
	db := setupTestFaissDB(t)
	ctx := context.Background()

	// 테스트 벡터 추가
	vectors := []VectorRecord{
		{
			ID:     "test1",
			Vector: []float32{1.0, 0.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "삭제될 구문",
			},
		},
		{
			ID:     "test2",
			Vector: []float32{0.0, 1.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "유지될 구문",
			},
		},
	}

	err := db.AddVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("벡터 추가 실패: %v", err)
	}

	// 벡터 삭제
	err = db.Delete(ctx, []string{"test1"})
	if err != nil {
		t.Errorf("Delete() 에러 = %v", err)
	}

	// 삭제 후 통계 확인
	stats, err := db.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() 에러 = %v", err)
	}

	if stats.TotalVectors != 1 {
		t.Errorf("삭제 후 벡터 수 = %d, 예상 = 1", stats.TotalVectors)
	}

	// 삭제된 벡터로 검색해도 결과가 나오지 않아야 함
	queryVector := []float32{1.0, 0.0, 0.0, 0.0}
	results, err := db.Search(ctx, queryVector, 10)
	if err != nil {
		t.Errorf("Search() 에러 = %v", err)
	}

	// 유지된 구문만 검색되어야 함
	found := false
	for _, result := range results {
		if result.Phrase == "삭제될 구문" {
			t.Error("삭제된 구문이 검색 결과에 나타남")
		}
		if result.Phrase == "유지될 구문" {
			found = true
		}
	}

	if !found {
		t.Error("유지되어야 할 구문이 검색되지 않음")
	}
}

func TestFaissDB_HealthCheck(t *testing.T) {
	db := setupTestFaissDB(t)
	ctx := context.Background()

	err := db.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck() 에러 = %v", err)
	}
}

// setupTestFaissDB 테스트용 FaissDB 인스턴스 생성
func setupTestFaissDB(t *testing.T) *FaissDB {
	tempDir, err := os.MkdirTemp("", "faiss_test")
	if err != nil {
		t.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}

	// 테스트 종료 시 정리
	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	cfg := config.VectorConfig{
		Type:      "faiss",
		IndexPath: tempDir,
		Dimension: 4, // 테스트용으로 작은 차원 사용
	}

	logger, _ := zap.NewDevelopment()
	db := NewFaissDB(cfg, logger)

	ctx := context.Background()
	err = db.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() 실패: %v", err)
	}

	return db
}

func TestFaissDB_SaveAndLoadIndex(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "faiss_test")
	if err != nil {
		t.Fatalf("임시 디렉토리 생성 실패: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := config.VectorConfig{
		Type:      "faiss",
		IndexPath: tempDir,
		Dimension: 4,
	}

	logger, _ := zap.NewDevelopment()

	// 첫 번째 DB 인스턴스에서 벡터 저장
	db1 := NewFaissDB(cfg, logger)
	ctx := context.Background()
	err = db1.Initialize(ctx)
	if err != nil {
		t.Fatalf("Initialize() 실패: %v", err)
	}

	vectors := []VectorRecord{
		{
			ID:     "test1",
			Vector: []float32{1.0, 0.0, 0.0, 0.0},
			Metadata: map[string]interface{}{
				"phrase": "테스트 구문",
			},
		},
	}

	err = db1.AddVectors(ctx, vectors)
	if err != nil {
		t.Fatalf("벡터 추가 실패: %v", err)
	}

	db1.Close()

	// 두 번째 DB 인스턴스에서 저장된 인덱스 로드
	db2 := NewFaissDB(cfg, logger)
	err = db2.Initialize(ctx)
	if err != nil {
		t.Fatalf("두 번째 Initialize() 실패: %v", err)
	}

	stats, err := db2.GetStats(ctx)
	if err != nil {
		t.Errorf("GetStats() 에러 = %v", err)
	}

	if stats.TotalVectors != 1 {
		t.Errorf("로드된 벡터 수 = %d, 예상 = 1", stats.TotalVectors)
	}

	// 저장된 벡터로 검색
	queryVector := []float32{1.0, 0.0, 0.0, 0.0}
	results, err := db2.Search(ctx, queryVector, 1)
	if err != nil {
		t.Errorf("Search() 에러 = %v", err)
	}

	if len(results) != 1 || results[0].Phrase != "테스트 구문" {
		t.Error("저장된 인덱스에서 벡터를 올바르게 로드하지 못함")
	}

	db2.Close()
}