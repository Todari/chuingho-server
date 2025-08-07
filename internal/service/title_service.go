package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/database"
	"github.com/Todari/chuingho-server/internal/vector"
	"github.com/Todari/chuingho-server/pkg/model"
	"github.com/Todari/chuingho-server/pkg/util"
)

// TitleService 췽호 추천 관련 비즈니스 로직
type TitleService struct {
	db           *database.DB
	vectorDB     vector.VectorDB
	mlClient     *MLClient
	resumeService *ResumeService
	logger       *zap.Logger
}

// NewTitleService 새로운 췽호 서비스 생성
func NewTitleService(
	db *database.DB,
	vectorDB vector.VectorDB,
	mlClient *MLClient,
	resumeService *ResumeService,
	logger *zap.Logger,
) *TitleService {
	return &TitleService{
		db:           db,
		vectorDB:     vectorDB,
		mlClient:     mlClient,
		resumeService: resumeService,
		logger:       logger,
	}
}

// GenerateTitles 췽호 추천 생성
func (s *TitleService) GenerateTitles(ctx context.Context, resumeID uuid.UUID) (*model.GenerateTitlesResponse, error) {
	startTime := time.Now()
	requestID, _ := util.GenerateRequestID()

	s.logger.Info("췽호 생성 시작",
		zap.String("request_id", requestID),
		zap.String("resume_id", resumeID.String()))

	// 자기소개서 상태 업데이트
	if err := s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusProcessing); err != nil {
		return nil, fmt.Errorf("상태 업데이트 실패: %w", err)
	}

	// 자기소개서 내용 조회
	content, err := s.resumeService.GetResumeContent(ctx, resumeID)
	if err != nil {
		s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusFailed)
		return nil, fmt.Errorf("자기소개서 내용 조회 실패: %w", err)
	}

	// 텍스트가 너무 짧은 경우 처리
	if len(content) < 50 {
		s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusFailed)
		return nil, fmt.Errorf("자기소개서 내용이 너무 짧습니다 (최소 50자 필요)")
	}

	// ML 서비스로 임베딩 생성
	embedding, err := s.mlClient.GetEmbedding(ctx, content)
	if err != nil {
		s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusFailed)
		return nil, fmt.Errorf("임베딩 생성 실패: %w", err)
	}

	// 벡터 검색으로 유사한 췽호 후보 찾기
	searchResults, err := s.vectorDB.Search(ctx, embedding, 50) // top-50 후보
	if err != nil {
		s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusFailed)
		return nil, fmt.Errorf("벡터 검색 실패: %w", err)
	}

	if len(searchResults) == 0 {
		s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusFailed)
		return nil, fmt.Errorf("적합한 췽호 후보를 찾을 수 없습니다")
	}

	// 다양성 기반 재순위화 후 상위 3개 선택
	finalTitles := s.diversityRanking(searchResults, 3)

	// 결과 저장
	processingTime := int(time.Since(startTime).Milliseconds())
	if err := s.saveTitleRecommendation(ctx, resumeID, finalTitles, searchResults, processingTime); err != nil {
		s.logger.Error("췽호 추천 결과 저장 실패", zap.Error(err))
	}

	// 자기소개서 상태 업데이트
	if err := s.resumeService.UpdateResumeStatus(ctx, resumeID, model.ResumeStatusCompleted); err != nil {
		s.logger.Error("완료 상태 업데이트 실패", zap.Error(err))
	}

	s.logger.Info("췽호 생성 완료",
		zap.String("request_id", requestID),
		zap.String("resume_id", resumeID.String()),
		zap.Int("processing_time_ms", processingTime),
		zap.Strings("titles", finalTitles))

	return &model.GenerateTitlesResponse{
		Titles: finalTitles,
	}, nil
}

// diversityRanking 다양성 기반 재순위화
func (s *TitleService) diversityRanking(results []model.VectorSearchResult, topK int) []string {
	if len(results) <= topK {
		titles := make([]string, len(results))
		for i, result := range results {
			titles[i] = result.Phrase
		}
		return titles
	}

	// MMR (Maximal Marginal Relevance) 알고리즘 유사 구현
	selected := make([]model.VectorSearchResult, 0, topK)
	remaining := make([]model.VectorSearchResult, len(results))
	copy(remaining, results)

	// 첫 번째는 유사도가 가장 높은 것 선택
	selected = append(selected, remaining[0])
	remaining = remaining[1:]

	// 나머지는 유사도와 다양성을 고려하여 선택
	for len(selected) < topK && len(remaining) > 0 {
		bestIdx := 0
		bestScore := float32(-1)

		for i, candidate := range remaining {
			// 유사도 점수 (0.7 가중치)
			relevanceScore := candidate.Score * 0.7

			// 다양성 점수 (0.3 가중치) - 이미 선택된 것들과의 차이
			diversityScore := s.calculateDiversity(candidate.Phrase, selected) * 0.3

			totalScore := relevanceScore + diversityScore

			if totalScore > bestScore {
				bestScore = totalScore
				bestIdx = i
			}
		}

		selected = append(selected, remaining[bestIdx])
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	// 최종 결과 추출
	titles := make([]string, len(selected))
	for i, result := range selected {
		titles[i] = result.Phrase
	}

	return titles
}

// calculateDiversity 다양성 점수 계산 (단순 구현)
func (s *TitleService) calculateDiversity(candidate string, selected []model.VectorSearchResult) float32 {
	if len(selected) == 0 {
		return 1.0
	}

	minSimilarity := float32(1.0)
	for _, sel := range selected {
		similarity := s.calculateStringSimilarity(candidate, sel.Phrase)
		if similarity < minSimilarity {
			minSimilarity = similarity
		}
	}

	return 1.0 - minSimilarity
}

// calculateStringSimilarity 문자열 유사도 계산 (Jaccard 유사도)
func (s *TitleService) calculateStringSimilarity(a, b string) float32 {
	// 단어 단위로 분할
	wordsA := make(map[string]bool)
	wordsB := make(map[string]bool)

	for _, word := range []rune(a) {
		wordsA[string(word)] = true
	}
	for _, word := range []rune(b) {
		wordsB[string(word)] = true
	}

	// 교집합과 합집합 계산
	intersection := 0
	union := len(wordsA)

	for word := range wordsB {
		if wordsA[word] {
			intersection++
		} else {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}

	return float32(intersection) / float32(union)
}

// saveTitleRecommendation 췽호 추천 결과 저장
func (s *TitleService) saveTitleRecommendation(
	ctx context.Context,
	resumeID uuid.UUID,
	titles []string,
	searchResults []model.VectorSearchResult,
	processingTime int,
) error {
	// 유사도 점수 맵 생성
	scores := make(map[string]float32)
	for _, result := range searchResults {
		scores[result.Phrase] = result.Score
	}

	// 선택된 췽호들의 점수만 추출
	selectedScores := make(map[string]float32)
	for _, title := range titles {
		if score, exists := scores[title]; exists {
			selectedScores[title] = score
		}
	}

	_, err := s.db.Pool.Exec(ctx, `
		INSERT INTO title_recommendations (
			resume_id, titles, vector_similarity_scores, 
			processing_time_ms, ml_model_version, created_at
		) VALUES ($1, $2, $3, $4, $5, NOW())`,
		resumeID, titles, selectedScores, processingTime, "KoSimCSE-bert-v1")

	if err != nil {
		return fmt.Errorf("췽호 추천 결과 저장 실패: %w", err)
	}

	return nil
}

// GetTitleHistory 췽호 추천 기록 조회
func (s *TitleService) GetTitleHistory(ctx context.Context, resumeID uuid.UUID) ([]model.TitleRecommendation, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, resume_id, titles, vector_similarity_scores,
			   processing_time_ms, ml_model_version, created_at
		FROM title_recommendations 
		WHERE resume_id = $1 
		ORDER BY created_at DESC`,
		resumeID)

	if err != nil {
		return nil, fmt.Errorf("췽호 추천 기록 조회 실패: %w", err)
	}
	defer rows.Close()

	var recommendations []model.TitleRecommendation
	for rows.Next() {
		var rec model.TitleRecommendation
		err := rows.Scan(
			&rec.ID, &rec.ResumeID, &rec.Titles, &rec.VectorSimilarityScores,
			&rec.ProcessingTimeMs, &rec.MLModelVersion, &rec.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("췽호 추천 행 스캔 실패: %w", err)
		}
		recommendations = append(recommendations, rec)
	}

	return recommendations, nil
}

// GetRandomTitles 랜덤 췽호 추천 (벡터 DB가 비어있을 때)
func (s *TitleService) GetRandomTitles(ctx context.Context) []string {
	defaultTitles := []string{
		"창의적 혁신가", "열정적 도전자", "섬세한 분석가",
		"적극적 리더", "신중한 전략가", "유연한 커뮤니케이터",
		"끈기있는 실행자", "협력적 팀워커", "논리적 사고자",
		"감성적 기획자", "체계적 관리자", "직관적 문제해결사",
	}

	// 무작위로 3개 선택
	rand.Seed(time.Now().UnixNano())
	selected := make([]string, 3)
	used := make(map[int]bool)

	for i := 0; i < 3; i++ {
		for {
			idx := rand.Intn(len(defaultTitles))
			if !used[idx] {
				selected[i] = defaultTitles[idx]
				used[idx] = true
				break
			}
		}
	}

	s.logger.Info("기본 췽호 반환", zap.Strings("titles", selected))
	return selected
}