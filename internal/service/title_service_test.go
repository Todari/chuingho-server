package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/Todari/chuingho-server/internal/vector"
	"github.com/Todari/chuingho-server/pkg/model"
)

// MockMLClient ML 클라이언트 모킹
type MockMLClient struct {
	mock.Mock
}

func (m *MockMLClient) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
    args := m.Mock.Called(ctx, text)
	return args.Get(0).([]float32), args.Error(1)
}

func (m *MockMLClient) GetBatchEmbeddings(ctx context.Context, phrases []string) (map[string][]float32, error) {
    args := m.Mock.Called(ctx, phrases)
	return args.Get(0).(map[string][]float32), args.Error(1)
}

func (m *MockMLClient) GenerateDynamicCombinations(ctx context.Context, resumeText string, topK int) (*model.DynamicCombinationResponse, error) {
    args := m.Mock.Called(ctx, resumeText, topK)
	return args.Get(0).(*model.DynamicCombinationResponse), args.Error(1)
}

func (m *MockMLClient) HealthCheck(ctx context.Context) error {
    args := m.Mock.Called(ctx)
	return args.Error(0)
}

// MockResumeService Resume 서비스 모킹
type MockResumeService struct {
	mock.Mock
}

func (m *MockResumeService) UploadResume(ctx context.Context, text string) (*model.UploadResumeResponse, error) {
    args := m.Mock.Called(ctx, text)
	return args.Get(0).(*model.UploadResumeResponse), args.Error(1)
}

func (m *MockResumeService) GetResumeContent(ctx context.Context, resumeID uuid.UUID) (string, error) {
    args := m.Mock.Called(ctx, resumeID)
	return args.Get(0).(string), args.Error(1)
}

func (m *MockResumeService) UpdateResumeStatus(ctx context.Context, resumeID uuid.UUID, status model.ResumeStatus) error {
    args := m.Mock.Called(ctx, resumeID, status)
	return args.Error(0)
}

func TestTitleService_GenerateTitles_DynamicCombination_Success(t *testing.T) {
	// Given
	logger := zaptest.NewLogger(t)
	mockMLClient := &MockMLClient{}
	mockResumeService := &MockResumeService{}

    titleService := &TitleService{
        mlClient:      mockMLClient, // MLClientAPI를 만족
        resumeService: mockResumeService,
        logger:        logger,
    }

	resumeID := uuid.New()
	resumeText := "안녕하세요. 저는 풀스택 개발자로서 React와 Node.js를 활용한 웹 애플리케이션 개발에 전문성을 가지고 있습니다. 새로운 기술을 학습하는 것을 좋아하며, 클라우드 환경에서의 DevOps와 자동화에 관심이 많습니다."

	// Mock 설정
	mockResumeService.On("UpdateResumeStatus", mock.Anything, resumeID, model.ResumeStatusProcessing).Return(nil)
	mockResumeService.On("GetResumeContent", mock.Anything, resumeID).Return(resumeText, nil)
	
	expectedResponse := &model.DynamicCombinationResponse{
		Combinations: []string{
			"혁신적인 개발자",
			"자동화 전문가",
			"학습하는 엔지니어",
		},
		Details: []model.CombinationDetail{
			{Phrase: "혁신적인 개발자", Similarity: 0.92},
			{Phrase: "자동화 전문가", Similarity: 0.88},
			{Phrase: "학습하는 엔지니어", Similarity: 0.85},
		},
		ProcessingTime:      0.245,
		TotalGenerated:      600,
		FilteredAdjectives:  20,
		FilteredNouns:       30,
	}
	
	mockMLClient.On("GenerateDynamicCombinations", mock.Anything, resumeText, 3).Return(expectedResponse, nil)
	mockResumeService.On("UpdateResumeStatus", mock.Anything, resumeID, model.ResumeStatusCompleted).Return(nil)

	// When
	result, err := titleService.GenerateTitles(context.Background(), resumeID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Titles, 3)
	assert.Equal(t, "혁신적인 개발자", result.Titles[0])
	assert.Equal(t, "자동화 전문가", result.Titles[1])
	assert.Equal(t, "학습하는 엔지니어", result.Titles[2])

	// 모든 모킹이 호출되었는지 확인
	mockMLClient.AssertExpectations(t)
	mockResumeService.AssertExpectations(t)
}

func TestTitleService_GenerateTitles_DynamicCombination_Fallback_Success(t *testing.T) {
	// Given
	logger := zaptest.NewLogger(t)
	mockMLClient := &MockMLClient{}
	mockResumeService := &MockResumeService{}
	mockVectorDB := &MockVectorDB{}

	titleService := &TitleService{
		mlClient:      mockMLClient,
		resumeService: mockResumeService,
		vectorDB:      mockVectorDB,
		logger:        logger,
	}

	resumeID := uuid.New()
	resumeText := "창의적이고 열정적인 개발자입니다."

	// Mock 설정 - 동적 조합 생성 실패
	mockResumeService.On("UpdateResumeStatus", mock.Anything, resumeID, model.ResumeStatusProcessing).Return(nil)
	mockResumeService.On("GetResumeContent", mock.Anything, resumeID).Return(resumeText, nil)
	mockMLClient.On("GenerateDynamicCombinations", mock.Anything, resumeText, 3).Return((*model.DynamicCombinationResponse)(nil), assert.AnError)
	
	// 폴백 - 기존 방식
	mockEmbedding := []float32{0.1, 0.2, 0.3}
	mockMLClient.On("GetEmbedding", mock.Anything, resumeText).Return(mockEmbedding, nil)
	
	mockSearchResults := []model.VectorSearchResult{
		{Phrase: "창의적 설계자", Score: 0.95},
		{Phrase: "열정적 개발자", Score: 0.90},
		{Phrase: "협력적 리더", Score: 0.85},
	}
	mockVectorDB.On("Search", mock.Anything, mockEmbedding, 50).Return(mockSearchResults, nil)

	// When
	result, err := titleService.GenerateTitles(context.Background(), resumeID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Titles, 3)

	// 모든 모킹이 호출되었는지 확인
	mockMLClient.AssertExpectations(t)
	mockResumeService.AssertExpectations(t)
	mockVectorDB.AssertExpectations(t)
}

func TestTitleService_GenerateTitles_EmptyDynamicResponse_Fallback(t *testing.T) {
	// Given
	logger := zaptest.NewLogger(t)
	mockMLClient := &MockMLClient{}
	mockResumeService := &MockResumeService{}
	mockVectorDB := &MockVectorDB{}

	titleService := &TitleService{
		mlClient:      mockMLClient,
		resumeService: mockResumeService,
		vectorDB:      mockVectorDB,
		logger:        logger,
	}

	resumeID := uuid.New()
	resumeText := "개발자입니다."

	// Mock 설정 - 동적 조합 생성 성공하지만 결과 없음
	mockResumeService.On("UpdateResumeStatus", mock.Anything, resumeID, model.ResumeStatusProcessing).Return(nil)
	mockResumeService.On("GetResumeContent", mock.Anything, resumeID).Return(resumeText, nil)
	
	emptyResponse := &model.DynamicCombinationResponse{
		Combinations:        []string{}, // 빈 결과
		Details:             []model.CombinationDetail{},
		ProcessingTime:      0.1,
		TotalGenerated:      0,
		FilteredAdjectives:  0,
		FilteredNouns:       0,
	}
	mockMLClient.On("GenerateDynamicCombinations", mock.Anything, resumeText, 3).Return(emptyResponse, nil)
	
	// 폴백 - 기존 방식
	mockEmbedding := []float32{0.1, 0.2, 0.3}
	mockMLClient.On("GetEmbedding", mock.Anything, resumeText).Return(mockEmbedding, nil)
	
	mockSearchResults := []model.VectorSearchResult{
		{Phrase: "기본 개발자", Score: 0.80},
		{Phrase: "일반 프로그래머", Score: 0.75},
		{Phrase: "평범한 엔지니어", Score: 0.70},
	}
	mockVectorDB.On("Search", mock.Anything, mockEmbedding, 50).Return(mockSearchResults, nil)

	// When
	result, err := titleService.GenerateTitles(context.Background(), resumeID)

	// Then
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Titles, 3)

	// 모든 모킹이 호출되었는지 확인
	mockMLClient.AssertExpectations(t)
	mockResumeService.AssertExpectations(t)
	mockVectorDB.AssertExpectations(t)
}

func TestTitleService_DiversityRanking(t *testing.T) {
	// Given
	logger := zaptest.NewLogger(t)
	titleService := &TitleService{
		logger: logger,
	}

	searchResults := []model.VectorSearchResult{
		{Phrase: "창의적 개발자", Score: 0.95},
		{Phrase: "창의적 프로그래머", Score: 0.94}, // 유사함
		{Phrase: "분석적 사고자", Score: 0.90},     // 다름
		{Phrase: "창의적 설계자", Score: 0.89},     // 유사함
		{Phrase: "협력적 리더", Score: 0.85},       // 다름
	}

	// When
	result := titleService.diversityRanking(searchResults, 3)

	// Then
	assert.Len(t, result, 3)
	assert.Equal(t, "창의적 개발자", result[0]) // 가장 높은 점수
	
	// 나머지 두 개는 다양성을 고려해서 선택되어야 함
	// "창의적 프로그래머"보다는 "분석적 사고자"나 "협력적 리더"가 선택될 가능성이 높음
	assert.NotEqual(t, "창의적 프로그래머", result[1])
	assert.NotEqual(t, "창의적 설계자", result[2])
}

func TestTitleService_CalculateStringSimilarity(t *testing.T) {
	// Given
	logger := zaptest.NewLogger(t)
	titleService := &TitleService{
		logger: logger,
	}

	tests := []struct {
		name     string
		a        string
		b        string
		expected float32
	}{
		{
			name:     "동일한 문자열",
			a:        "창의적 개발자",
			b:        "창의적 개발자",
			expected: 1.0,
		},
		{
			name:     "완전히 다른 문자열",
			a:        "창의적 개발자",
			b:        "협력적 리더",
			expected: 0.0,
		},
		{
			name:     "부분적으로 유사한 문자열",
			a:        "창의적 개발자",
			b:        "창의적 설계자",
			expected: 0.6, // "창의적"과 "자" 공통
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := titleService.calculateStringSimilarity(tt.a, tt.b)

			// Then
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTitleService_GetRandomTitles(t *testing.T) {
	// Given
	logger := zaptest.NewLogger(t)
	titleService := &TitleService{
		logger: logger,
	}

	// When
	result := titleService.GetRandomTitles(context.Background())

	// Then
	assert.Len(t, result, 3)
	
	// 모든 결과가 비어있지 않은지 확인
	for _, title := range result {
		assert.NotEmpty(t, title)
	}
	
	// 중복이 없는지 확인
	titleSet := make(map[string]bool)
	for _, title := range result {
		assert.False(t, titleSet[title], "중복된 칭호가 발견됨: %s", title)
		titleSet[title] = true
	}
}

// MockVectorDB 벡터 DB 모킹
type MockVectorDB struct {
	mock.Mock
}

// VectorDB 인터페이스 호환을 위해 AddVectors 구현
func (m *MockVectorDB) AddVectors(ctx context.Context, vectors []vector.VectorRecord) error {
    args := m.Called(ctx, vectors)
    return args.Error(0)
}

func (m *MockVectorDB) Search(ctx context.Context, query []float32, topK int) ([]model.VectorSearchResult, error) {
	args := m.Called(ctx, query, topK)
	return args.Get(0).([]model.VectorSearchResult), args.Error(1)
}

func (m *MockVectorDB) Delete(ctx context.Context, ids []string) error {
    args := m.Called(ctx, ids)
    return args.Error(0)
}

func (m *MockVectorDB) Update(ctx context.Context, id string, vector []float32, metadata map[string]interface{}) error {
	args := m.Called(ctx, id, vector, metadata)
	return args.Error(0)
}

func (m *MockVectorDB) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVectorDB) GetStats(ctx context.Context) (*vector.VectorStats, error) {
    args := m.Called(ctx)
    return args.Get(0).(*vector.VectorStats), args.Error(1)
}

func (m *MockVectorDB) Initialize(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}

func (m *MockVectorDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

// 성능 테스트
func BenchmarkTitleService_DiversityRanking(b *testing.B) {
	logger := zaptest.NewLogger(b)
	titleService := &TitleService{
		logger: logger,
	}

	// 큰 데이터셋 생성
	searchResults := make([]model.VectorSearchResult, 1000)
	for i := 0; i < 1000; i++ {
		searchResults[i] = model.VectorSearchResult{
			Phrase: fmt.Sprintf("테스트 칭호 %d", i),
			Score:  float32(1000-i) / 1000.0,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = titleService.diversityRanking(searchResults, 10)
	}
}
