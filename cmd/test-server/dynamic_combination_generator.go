package main

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"
)

// DynamicCombinationGenerator 동적 조합 생성기
type DynamicCombinationGenerator struct {
	adjectives []string
	nouns      []string
	rand       *rand.Rand
}

// NewDynamicCombinationGenerator 새로운 동적 조합 생성기 생성
func NewDynamicCombinationGenerator() *DynamicCombinationGenerator {
	gen := &DynamicCombinationGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	
	// 형용사/명사 풀 로드
	gen.loadWordPools()
	
	return gen
}

// loadWordPools 형용사/명사 풀 로드
func (dcg *DynamicCombinationGenerator) loadWordPools() {
	// 형용사 로드 (여러 경로 시도)
	adjectivePaths := []string{"../../data/adjectives.txt", "./data/adjectives.txt", "data/adjectives.txt"}
	for _, path := range adjectivePaths {
		dcg.adjectives = dcg.loadWordPool(path)
		if len(dcg.adjectives) > 0 {
			break
		}
	}
	if len(dcg.adjectives) == 0 {
		// 기본 형용사 풀
		dcg.adjectives = []string{
			"아름다운", "따뜻한", "밝은", "새로운", "창의적인", "독창적인", "혁신적인",
			"차가운", "깊은", "높은", "부드러운", "강한", "빠른", "느린", "큰", "작은",
			"똑똑한", "현명한", "용감한", "친절한", "성실한", "부지런한", "정직한",
			"열정적인", "적극적인", "능동적인", "자발적인", "협력적인", "포용적인",
			"논리적인", "체계적인", "분석적인", "비판적인", "객관적인", "합리적인",
		}
	}
	
	// 명사 로드 (여러 경로 시도)
	nounPaths := []string{"../../data/nouns.txt", "./data/nouns.txt", "data/nouns.txt"}
	for _, path := range nounPaths {
		dcg.nouns = dcg.loadWordPool(path)
		if len(dcg.nouns) > 0 {
			break
		}
	}
	if len(dcg.nouns) == 0 {
		// 기본 명사 풀
		dcg.nouns = []string{
			"바람", "별", "꿈", "빛", "마음", "생각", "미래", "희망", "에너지", "열정",
			"바다", "하늘", "구름", "꽃", "나무", "물", "불", "길", "집", "문",
			"리더", "개발자", "설계자", "분석가", "기획자", "관리자", "전문가",
			"혁신가", "창작자", "탐험가", "도전자", "실행자", "완주자", "달성자",
			"사고자", "관찰자", "판단자", "의사결정자", "문제해결사", "커뮤니케이터",
		}
	}
}

// loadWordPool 단어 풀 파일 로드
func (dcg *DynamicCombinationGenerator) loadWordPool(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		// 에러 메시지를 줄여서 로그 스팸을 방지
		return nil
	}
	defer file.Close()

	var words []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			words = append(words, line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("파일 읽기 실패: %v\n", err)
		return nil
	}

	fmt.Printf("단어 풀 로드 완료 (%s): %d개\n", filename, len(words))
	return words
}

// GenerateDynamicCombinations 동적 조합 생성 (실제 ML 서비스 시뮬레이션)
func (dcg *DynamicCombinationGenerator) GenerateDynamicCombinations(resumeText string, topK int) map[string]interface{} {
	startTime := time.Now()
	
	// 1단계: 자기소개서 키워드 분석
	keywords := dcg.extractKeywords(resumeText)
	
	// 2단계: 키워드 기반 형용사/명사 필터링
	adjFilterCount := 20
	nounFilterCount := 30
	
	relevantAdjectives := dcg.filterRelevantWords(dcg.adjectives, keywords, adjFilterCount)
	relevantNouns := dcg.filterRelevantWords(dcg.nouns, keywords, nounFilterCount)
	
	// 3단계: 조합 생성
	combinations := []string{}
	for _, adj := range relevantAdjectives {
		for _, noun := range relevantNouns {
			combinations = append(combinations, fmt.Sprintf("%s %s", adj, noun))
		}
	}
	
	// 4단계: 의미적 유사도 계산 시뮬레이션
	similarities := []map[string]interface{}{}
	for _, combination := range combinations {
		// 실제로는 벡터 유사도를 계산하지만, 여기서는 키워드 기반으로 시뮬레이션
		similarity := dcg.calculateSemanticSimilarity(combination, keywords)
		similarities = append(similarities, map[string]interface{}{
			"phrase":     combination,
			"similarity": similarity,
		})
	}
	
	// 5단계: 유사도 기준 정렬
	dcg.sortBySimilarity(similarities)
	
    // 6단계: 다양성 고려 선택 (MMR 알고리즘)
    finalResults := dcg.selectDiverseCombinations(similarities, topK)

    // 상위 유사 조합 5개 추출
    topSimilar := []map[string]interface{}{}
    maxTop := 5
    if len(similarities) < maxTop { maxTop = len(similarities) }
    for i := 0; i < maxTop; i++ {
        topSimilar = append(topSimilar, similarities[i])
    }
	
	processingTime := time.Since(startTime).Seconds()
	
	// 결과 반환
	finalCombinations := []string{}
	for _, result := range finalResults {
		finalCombinations = append(finalCombinations, result["phrase"].(string))
	}
	
    return map[string]interface{}{
		"combinations":        finalCombinations,
		"details":             finalResults,
		"processing_time":     processingTime,
		"total_generated":     len(combinations),
		"filtered_adjectives": len(relevantAdjectives),
        "filtered_nouns":      len(relevantNouns),
        "top_similar":         topSimilar,
	}
}

// extractKeywords 텍스트에서 키워드 추출
func (dcg *DynamicCombinationGenerator) extractKeywords(text string) []string {
	text = strings.ToLower(text)
	keywords := []string{}
	
	// 기술 관련
	techKeywords := []string{"개발", "프로그래밍", "코딩", "기술", "엔지니어", "시스템", "소프트웨어", "웹", "앱", "데이터"}
	for _, keyword := range techKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, "기술")
			break
		}
	}
	
	// 창의성 관련
	creativityKeywords := []string{"창의", "아이디어", "혁신", "독창", "상상", "기획", "디자인", "예술"}
	for _, keyword := range creativityKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, "창의")
			break
		}
	}
	
	// 리더십 관련
	leadershipKeywords := []string{"리더", "지도", "이끌", "주도", "팀장", "관리", "책임", "지휘"}
	for _, keyword := range leadershipKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, "리더십")
			break
		}
	}
	
	// 협력 관련
	collaborationKeywords := []string{"협력", "소통", "팀워크", "화합", "관계", "네트워킹", "파트너십"}
	for _, keyword := range collaborationKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, "협력")
			break
		}
	}
	
	// 분석 관련
	analysisKeywords := []string{"분석", "논리", "체계", "정확", "꼼꼼", "세밀", "신중", "판단"}
	for _, keyword := range analysisKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, "분석")
			break
		}
	}
	
	// 열정 관련
	passionKeywords := []string{"열정", "적극", "도전", "목표", "성취", "노력", "의욕", "동기"}
	for _, keyword := range passionKeywords {
		if strings.Contains(text, keyword) {
			keywords = append(keywords, "열정")
			break
		}
	}
	
	return keywords
}

// filterRelevantWords 키워드 기반으로 관련성 높은 단어들 필터링
func (dcg *DynamicCombinationGenerator) filterRelevantWords(words []string, keywords []string, topK int) []string {
	if len(words) <= topK {
		return words
	}
	
	// 키워드와의 관련성 점수 계산
	scored := []struct {
		word  string
		score float64
	}{}
	
	for _, word := range words {
		score := dcg.calculateKeywordRelevance(word, keywords)
		scored = append(scored, struct {
			word  string
			score float64
		}{word, score})
	}
	
	// 점수 기준 정렬
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	
	// 상위 topK개 선택 + 일부 랜덤 선택 (다양성 확보)
	result := []string{}
	
	// 상위 70% 선택
	topCount := int(float64(topK) * 0.7)
	for i := 0; i < topCount && i < len(scored); i++ {
		result = append(result, scored[i].word)
	}
	
	// 나머지 30%는 랜덤 선택
	remaining := topK - len(result)
	if remaining > 0 && len(scored) > topCount {
		availableWords := []string{}
		for i := topCount; i < len(scored); i++ {
			availableWords = append(availableWords, scored[i].word)
		}
		
		// 랜덤 셔플
		for i := len(availableWords) - 1; i > 0; i-- {
			j := dcg.rand.Intn(i + 1)
			availableWords[i], availableWords[j] = availableWords[j], availableWords[i]
		}
		
		for i := 0; i < remaining && i < len(availableWords); i++ {
			result = append(result, availableWords[i])
		}
	}
	
	return result
}

// calculateKeywordRelevance 단어와 키워드들 간의 관련성 점수 계산
func (dcg *DynamicCombinationGenerator) calculateKeywordRelevance(word string, keywords []string) float64 {
	if len(keywords) == 0 {
		return dcg.rand.Float64() // 키워드가 없으면 랜덤 점수
	}
	
	score := 0.0
	wordLower := strings.ToLower(word)
	
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		
		// 직접 포함 관계
		if strings.Contains(wordLower, keywordLower) || strings.Contains(keywordLower, wordLower) {
			score += 1.0
			continue
		}
		
		// 의미적 연관성 (단순화된 버전)
		semanticScore := dcg.calculateSemanticRelevance(wordLower, keywordLower)
		score += semanticScore
	}
	
	// 기본 랜덤 점수 추가 (다양성 확보)
	score += dcg.rand.Float64() * 0.3
	
	return score
}

// calculateSemanticRelevance 의미적 연관성 계산 (단순화된 버전)
func (dcg *DynamicCombinationGenerator) calculateSemanticRelevance(word, keyword string) float64 {
	// 실제로는 워드 임베딩을 사용하지만, 여기서는 규칙 기반으로 시뮬레이션
	relevanceMap := map[string]map[string]float64{
		"기술": {
			"혁신적인": 0.8, "창의적인": 0.7, "전문적인": 0.9, "체계적인": 0.6,
			"개발자": 0.9, "엔지니어": 0.9, "전문가": 0.7, "설계자": 0.8,
		},
        "창의": {
            "독창적인": 0.9, "혁신적인": 0.8, "상상력있는": 0.8, "예술적인": 0.7,
            "창의적인": 0.9,
            "기획자": 0.8, "설계자": 0.7, "창작자": 0.9, "혁신가": 0.9,
            // 약한 연상: 자연어적 시적 조합 허용을 위한 소량 가중치
            "바람": 0.3,
        },
		"리더십": {
			"주도적인": 0.9, "카리스마있는": 0.8, "책임감있는": 0.7, "결단력있는": 0.8,
			"리더": 0.9, "지도자": 0.9, "관리자": 0.7, "지휘자": 0.8,
		},
		"협력": {
			"소통하는": 0.8, "화합하는": 0.8, "포용적인": 0.7, "친화적인": 0.6,
			"팀워커": 0.9, "브릿지": 0.7, "커뮤니케이터": 0.8, "조정자": 0.7,
		},
		"분석": {
			"논리적인": 0.9, "체계적인": 0.8, "정확한": 0.7, "신중한": 0.6,
			"분석가": 0.9, "관찰자": 0.7, "판단자": 0.8, "탐구자": 0.7,
		},
		"열정": {
			"적극적인": 0.8, "도전적인": 0.8, "의욕적인": 0.9, "에너지넘치는": 0.8,
			"도전자": 0.8, "실행자": 0.7, "달성자": 0.7, "추진자": 0.8,
		},
	}
	
	if keywordMap, exists := relevanceMap[keyword]; exists {
		if score, exists := keywordMap[word]; exists {
			return score
		}
	}
	
	// 문자열 유사도 기반 점수 (Levenshtein distance 단순화 버전)
	return dcg.calculateStringSimilarity(word, keyword) * 0.3
}

// calculateStringSimilarity 문자열 유사도 계산
func (dcg *DynamicCombinationGenerator) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}
	
	// 간단한 공통 문자 비율 계산
	runes1 := []rune(s1)
	runes2 := []rune(s2)
	
	common := 0
	for _, r1 := range runes1 {
		for _, r2 := range runes2 {
			if r1 == r2 {
				common++
				break
			}
		}
	}
	
	maxLen := math.Max(float64(len(runes1)), float64(len(runes2)))
	if maxLen == 0 {
		return 0
	}
	
	return float64(common) / maxLen
}

// calculateSemanticSimilarity 의미적 유사도 계산 (시뮬레이션)
func (dcg *DynamicCombinationGenerator) calculateSemanticSimilarity(combination string, keywords []string) float64 {
	if len(keywords) == 0 {
		return dcg.rand.Float64()
	}
	
	words := strings.Fields(strings.ToLower(combination))
	totalScore := 0.0
	
	for _, word := range words {
		for _, keyword := range keywords {
			score := dcg.calculateSemanticRelevance(word, strings.ToLower(keyword))
			totalScore += score
		}
	}
	
	// 정규화
	normalizedScore := totalScore / (float64(len(words)) * float64(len(keywords)))
	
	// 랜덤 노이즈 추가 (0.2 비중)
	randomNoise := dcg.rand.Float64() * 0.2
	
	return math.Min(1.0, normalizedScore+randomNoise)
}

// sortBySimilarity 유사도 기준으로 정렬
func (dcg *DynamicCombinationGenerator) sortBySimilarity(similarities []map[string]interface{}) {
	for i := 0; i < len(similarities)-1; i++ {
		for j := i + 1; j < len(similarities); j++ {
			score1 := similarities[i]["similarity"].(float64)
			score2 := similarities[j]["similarity"].(float64)
			if score1 < score2 {
				similarities[i], similarities[j] = similarities[j], similarities[i]
			}
		}
	}
}

// selectDiverseCombinations MMR 알고리즘으로 다양성 고려 선택
func (dcg *DynamicCombinationGenerator) selectDiverseCombinations(similarities []map[string]interface{}, topK int) []map[string]interface{} {
	if len(similarities) <= topK {
		return similarities
	}
	
	selected := []map[string]interface{}{}
	remaining := make([]map[string]interface{}, len(similarities))
	copy(remaining, similarities)
	
	// 첫 번째는 가장 유사도 높은 것 선택
	selected = append(selected, remaining[0])
	remaining = remaining[1:]
	
	// 나머지는 유사도와 다양성 고려
	for len(selected) < topK && len(remaining) > 0 {
		bestScore := -1.0
		bestIdx := 0
		
		for i, candidate := range remaining {
			// 유사도 점수 (70%)
			relevanceScore := candidate["similarity"].(float64) * 0.7
			
			// 다양성 점수 (30%)
			diversityScore := dcg.calculateDiversityScore(candidate, selected) * 0.3
			
			totalScore := relevanceScore + diversityScore
			
			if totalScore > bestScore {
				bestScore = totalScore
				bestIdx = i
			}
		}
		
		selected = append(selected, remaining[bestIdx])
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}
	
	return selected
}

// calculateDiversityScore 다양성 점수 계산
func (dcg *DynamicCombinationGenerator) calculateDiversityScore(candidate map[string]interface{}, selected []map[string]interface{}) float64 {
	if len(selected) == 0 {
		return 1.0
	}
	
	candidatePhrase := candidate["phrase"].(string)
	candidateWords := strings.Fields(candidatePhrase)
	
	minSimilarity := 1.0
	for _, sel := range selected {
		selectedPhrase := sel["phrase"].(string)
		selectedWords := strings.Fields(selectedPhrase)
		
		// Jaccard 유사도 계산
		intersection := 0
		candidateSet := make(map[string]bool)
		for _, word := range candidateWords {
			candidateSet[word] = true
		}
		
		for _, word := range selectedWords {
			if candidateSet[word] {
				intersection++
			}
		}
		
		union := len(candidateWords) + len(selectedWords) - intersection
		var jaccardSim float64
		if union > 0 {
			jaccardSim = float64(intersection) / float64(union)
		}
		
		if jaccardSim < minSimilarity {
			minSimilarity = jaccardSim
		}
	}
	
	return 1.0 - minSimilarity // 유사도가 낮을수록 다양성 높음
}
