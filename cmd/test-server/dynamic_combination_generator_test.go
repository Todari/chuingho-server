package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDynamicCombinationGenerator_GenerateDynamicCombinations(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	tests := []struct {
		name         string
		resumeText   string
		topK         int
		expectedKeys []string
	}{
		{
			name:       "기술 개발자 자기소개서",
			resumeText: "안녕하세요. 저는 풀스택 개발자로서 React와 Node.js를 활용한 웹 애플리케이션 개발에 전문성을 가지고 있습니다. 새로운 기술을 학습하는 것을 좋아하며, 클라우드 환경에서의 DevOps와 자동화에 관심이 많습니다.",
			topK:       3,
			expectedKeys: []string{"combinations", "details", "processing_time", "total_generated", "filtered_adjectives", "filtered_nouns"},
		},
		{
			name:       "창의적 디자이너 자기소개서",
			resumeText: "저는 창의적인 UI/UX 디자이너입니다. 사용자 중심의 디자인 사고를 바탕으로 혁신적인 솔루션을 만들어내는 것을 좋아합니다. 아이디어를 현실로 만드는 과정에서 큰 보람을 느낍니다.",
			topK:       5,
			expectedKeys: []string{"combinations", "details", "processing_time", "total_generated", "filtered_adjectives", "filtered_nouns"},
		},
		{
			name:       "리더십 경영진 자기소개서",
			resumeText: "10년간의 팀 리더 경험을 바탕으로 조직을 이끌고 성과를 달성해왔습니다. 구성원들과의 소통을 중시하며, 협력을 통해 더 큰 시너지를 만들어내는 것이 제 강점입니다.",
			topK:       4,
			expectedKeys: []string{"combinations", "details", "processing_time", "total_generated", "filtered_adjectives", "filtered_nouns"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := generator.GenerateDynamicCombinations(tt.resumeText, tt.topK)

			// Then
			assert.NotNil(t, result)

			// 모든 필수 키가 존재하는지 확인
			for _, key := range tt.expectedKeys {
				assert.Contains(t, result, key, "결과에 %s 키가 없습니다", key)
			}

			// combinations 검증
			combinations, ok := result["combinations"].([]string)
			assert.True(t, ok, "combinations가 []string 타입이 아닙니다")
			assert.Len(t, combinations, tt.topK, "요청한 개수와 결과 개수가 다릅니다")

			// 모든 조합이 비어있지 않은지 확인
			for i, combination := range combinations {
				assert.NotEmpty(t, combination, "조합 %d가 비어있습니다", i)
				assert.Contains(t, combination, " ", "조합 %d에 공백이 없습니다 (형용사+명사 형태가 아님)", i)
			}

			// 중복 조합이 없는지 확인
			combinationSet := make(map[string]bool)
			for _, combination := range combinations {
				assert.False(t, combinationSet[combination], "중복된 조합이 발견됨: %s", combination)
				combinationSet[combination] = true
			}

			// details 검증
			details, ok := result["details"].([]map[string]interface{})
			assert.True(t, ok, "details가 올바른 타입이 아닙니다")
			assert.Len(t, details, tt.topK, "details 개수가 맞지 않습니다")

			// processing_time 검증
			processingTime, ok := result["processing_time"].(float64)
			assert.True(t, ok, "processing_time이 float64 타입이 아닙니다")
			assert.Greater(t, processingTime, 0.0, "처리 시간이 0보다 작거나 같습니다")
			assert.Less(t, processingTime, 10.0, "처리 시간이 너무 깁니다 (10초 초과)")

			// total_generated 검증
			totalGenerated, ok := result["total_generated"].(int)
			assert.True(t, ok, "total_generated가 int 타입이 아닙니다")
			assert.Greater(t, totalGenerated, 0, "생성된 총 조합 수가 0입니다")
			assert.LessOrEqual(t, totalGenerated, 600, "생성된 조합 수가 예상보다 많습니다 (20×30=600 초과)")

			// filtered counts 검증
			filteredAdjectives, ok := result["filtered_adjectives"].(int)
			assert.True(t, ok, "filtered_adjectives가 int 타입이 아닙니다")
			assert.Greater(t, filteredAdjectives, 0, "필터링된 형용사 수가 0입니다")
			assert.LessOrEqual(t, filteredAdjectives, 20, "필터링된 형용사 수가 20을 초과합니다")

			filteredNouns, ok := result["filtered_nouns"].(int)
			assert.True(t, ok, "filtered_nouns가 int 타입이 아닙니다")
			assert.Greater(t, filteredNouns, 0, "필터링된 명사 수가 0입니다")
			assert.LessOrEqual(t, filteredNouns, 30, "필터링된 명사 수가 30을 초과합니다")
		})
	}
}

func TestDynamicCombinationGenerator_ExtractKeywords(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	tests := []struct {
		name            string
		text            string
		expectedKeywords []string
	}{
		{
			name:            "기술 관련 텍스트",
			text:            "저는 소프트웨어 개발자로서 프로그래밍과 코딩에 전문성을 가지고 있습니다.",
			expectedKeywords: []string{"기술"},
		},
		{
			name:            "창의성 관련 텍스트",
			text:            "창의적인 아이디어로 혁신적인 솔루션을 만들어내는 것을 좋아합니다.",
			expectedKeywords: []string{"창의"},
		},
		{
			name:            "리더십 관련 텍스트",
			text:            "팀을 리더하며 구성원들을 이끌어 나가는 것이 제 강점입니다.",
			expectedKeywords: []string{"리더십"},
		},
		{
			name:            "협력 관련 텍스트",
			text:            "팀워크를 중시하며 동료들과의 소통과 협력을 통해 성과를 달성합니다.",
			expectedKeywords: []string{"협력"},
		},
		{
			name:            "분석 관련 텍스트",
			text:            "데이터를 체계적으로 분석하고 논리적으로 접근하여 정확한 판단을 내립니다.",
			expectedKeywords: []string{"기술", "분석"}, // "데이터"가 기술 키워드로도 분류됨
		},
		{
			name:            "열정 관련 텍스트",
			text:            "새로운 도전에 대한 열정과 목표 달성에 대한 강한 의욕을 가지고 있습니다.",
			expectedKeywords: []string{"열정"},
		},
		{
			name:            "복합 키워드 텍스트",
			text:            "창의적인 개발자로서 팀을 리더하며 혁신적인 기술 솔루션을 만들어냅니다.",
			expectedKeywords: []string{"창의", "기술", "리더십"},
		},
		{
			name:            "키워드 없는 텍스트",
			text:            "안녕하세요. 반갑습니다. 좋은 하루 되세요.",
			expectedKeywords: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := generator.extractKeywords(tt.text)

			// Then
			assert.ElementsMatch(t, tt.expectedKeywords, result, "추출된 키워드가 예상과 다릅니다")
		})
	}
}

func TestDynamicCombinationGenerator_FilterRelevantWords(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	words := []string{
		"혁신적인", "창의적인", "기술적인", "논리적인", "협력적인",
		"아름다운", "따뜻한", "차가운", "밝은", "어두운",
	}
	keywords := []string{"기술", "창의"}
	topK := 5

	// When
	result := generator.filterRelevantWords(words, keywords, topK)

	// Then
	assert.Len(t, result, topK)

	// 관련성이 높은 단어들이 포함되어 있는지 확인
	resultSet := make(map[string]bool)
	for _, word := range result {
		resultSet[word] = true
	}

	// 기술/창의 관련 단어들이 우선적으로 선택되었는지 확인
	assert.True(t, resultSet["혁신적인"] || resultSet["창의적인"] || resultSet["기술적인"], 
		"관련성이 높은 단어가 선택되지 않았습니다")
}

func TestDynamicCombinationGenerator_CalculateSemanticSimilarity(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	tests := []struct {
		name        string
		combination string
		keywords    []string
		expectHigh  bool
	}{
		{
			name:        "관련성 높은 조합",
			combination: "혁신적인 개발자",
			keywords:    []string{"기술", "창의"},
			expectHigh:  true,
		},
		{
			name:        "관련성 낮은 조합",
			combination: "차가운 바람",
			keywords:    []string{"기술", "창의"},
			expectHigh:  false,
		},
		{
			name:        "부분적 관련성",
			combination: "창의적인 바람",
			keywords:    []string{"기술", "창의"},
			expectHigh:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			result := generator.calculateSemanticSimilarity(tt.combination, tt.keywords)

			// Then
			assert.GreaterOrEqual(t, result, 0.0, "유사도는 0 이상이어야 합니다")
			assert.LessOrEqual(t, result, 1.0, "유사도는 1 이하여야 합니다")

			if tt.expectHigh {
				assert.Greater(t, result, 0.3, "관련성이 높은 조합의 유사도가 낮습니다")
			}
		})
	}
}

func TestDynamicCombinationGenerator_SelectDiverseCombinations(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	similarities := []map[string]interface{}{
		{"phrase": "창의적인 개발자", "similarity": 0.95},
		{"phrase": "창의적인 프로그래머", "similarity": 0.94}, // 유사함
		{"phrase": "분석적인 사고자", "similarity": 0.90},     // 다름
		{"phrase": "창의적인 설계자", "similarity": 0.89},     // 유사함
		{"phrase": "협력적인 리더", "similarity": 0.85},       // 다름
	}

	topK := 3

	// When
	result := generator.selectDiverseCombinations(similarities, topK)

	// Then
	assert.Len(t, result, topK)

	// 첫 번째는 가장 높은 유사도를 가져야 함
	assert.Equal(t, "창의적인 개발자", result[0]["phrase"])

	// 나머지는 다양성을 고려해서 선택되어야 함
	phrases := make([]string, len(result))
	for i, item := range result {
		phrases[i] = item["phrase"].(string)
	}

	// "창의적인"이라는 단어가 너무 많이 반복되지 않아야 함
	creativeCount := 0
	for _, phrase := range phrases {
		if strings.Contains(phrase, "창의적인") {
			creativeCount++
		}
	}
	assert.LessOrEqual(t, creativeCount, 2, "다양성이 부족합니다 (창의적인이 너무 많이 선택됨)")
}

func TestDynamicCombinationGenerator_CalculateDiversityScore(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	candidate := map[string]interface{}{"phrase": "분석적인 사고자"}
	selected := []map[string]interface{}{
		{"phrase": "창의적인 개발자"},
		{"phrase": "협력적인 리더"},
	}

	// When
	result := generator.calculateDiversityScore(candidate, selected)

	// Then
	assert.GreaterOrEqual(t, result, 0.0, "다양성 점수는 0 이상이어야 합니다")
	assert.LessOrEqual(t, result, 1.0, "다양성 점수는 1 이하여야 합니다")
	assert.Greater(t, result, 0.5, "다른 종류의 조합이므로 다양성 점수가 높아야 합니다")
}

func TestDynamicCombinationGenerator_EmptyInput(t *testing.T) {
	// Given
	generator := NewDynamicCombinationGenerator()

	// When
	result := generator.GenerateDynamicCombinations("", 3)

	// Then
	assert.NotNil(t, result)
	
	combinations, ok := result["combinations"].([]string)
	assert.True(t, ok)
	
	// 빈 입력에 대해서도 기본적인 조합이 생성되어야 함
	assert.Greater(t, len(combinations), 0, "빈 입력에 대해서도 기본 조합이 생성되어야 합니다")
}

// 성능 테스트
func BenchmarkDynamicCombinationGenerator_GenerateDynamicCombinations(b *testing.B) {
	generator := NewDynamicCombinationGenerator()
	resumeText := "안녕하세요. 저는 풀스택 개발자로서 React와 Node.js를 활용한 웹 애플리케이션 개발에 전문성을 가지고 있습니다. 새로운 기술을 학습하는 것을 좋아하며, 클라우드 환경에서의 DevOps와 자동화에 관심이 많습니다."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.GenerateDynamicCombinations(resumeText, 3)
	}
}

func BenchmarkDynamicCombinationGenerator_ExtractKeywords(b *testing.B) {
	generator := NewDynamicCombinationGenerator()
	text := "저는 창의적인 풀스택 개발자로서 팀 리더 역할을 맡아 혁신적인 기술 솔루션을 개발하고 있습니다. 데이터 분석과 협력을 통해 성과를 달성하는 것이 제 강점입니다."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generator.extractKeywords(text)
	}
}
