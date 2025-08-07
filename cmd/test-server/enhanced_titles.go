package main

import (
	"math/rand"
	"strings"
	"time"
)

// 확장된 칭호 풀 - 실제 서비스에서는 벡터 DB에서 가져옴
var enhancedTitlePool = []string{
	// 리더십 관련
	"적극적 리더", "주도적 지도자", "카리스마틱 리더", "혁신적 리더", "변화지향적 리더",
	"전략적 지휘자", "비전제시형 리더", "결단력있는 지도자", "포용적 리더", "멘토링형 지도자",
	"영감을주는 리더", "겸손한 지도자", "섬기는 리더", "변혁적 리더", "상황적응형 리더",
	
	// 창의성 관련
	"창의적 혁신가", "독창적 아이디어맨", "상상력풍부한 기획자", "아이디어 뱅크형 인재", "혁신적 사고자",
	"창조적 문제해결사", "독특한 발상가", "트렌드세터형 개발자", "새로운 관점의 제안자", "브레인스토밍 마스터",
	"예술적 창작자", "파괴적 혁신가", "융합적 사고자", "실험적 도전자", "비관습적 접근자",
	
	// 분석력 관련
	"논리적 분석가", "체계적 사고자", "데이터기반 분석가", "꼼꼼한 검토자", "세밀한 관찰자",
	"정확한 판단자", "신중한 의사결정자", "비판적 사고자", "합리적 추론가", "객관적 평가자",
	"통계적 분석가", "정량적 평가자", "심층적 탐구자", "과학적 접근자", "증거기반 판단자",
	
	// 협력 관련
	"협력적 팀워커", "소통지향적 브릿지", "화합형 조정자", "포용적 협력자", "네트워킹 전문가",
	"관계지향적 커뮤니케이터", "상호작용 전문가", "팀워크 촉진자", "갈등조정형 중재자", "협업 스페셜리스트",
	"다양성 포용 전문가", "문화간 소통 전문가", "팀 하모니 메이커", "협력적 문제해결사", "네트워킹 마스터",
	
	// 실행력 관련
	"실행력있는 추진자", "끈기있는 완주자", "목표지향적 달성자", "집중력있는 실행자", "결과지향적 완수자",
	"책임감강한 수행자", "신속한 실행자", "지속가능한 실천가", "성과중심 달성자", "완벽주의 실행자",
	"신속한 의사결정자", "민첩한 실행자", "효율성 추구자", "생산성 최적화 전문가", "프로세스 개선자",
	
	// 기술 관련
	"기술지향적 개발자", "혁신적 엔지니어", "문제해결형 프로그래머", "효율성추구 개발자", "품질중심 아키텍트",
	"사용자중심 디자이너", "데이터기반 분석가", "클라우드 전문가", "보안중심 관리자", "자동화 스페셜리스트",
	"풀스택 개발자", "DevOps 엔지니어", "AI/ML 전문가", "블록체인 개발자", "모바일 앱 개발자",
	
	// 커뮤니케이션 관련
	"설득력있는 스피커", "명확한 전달자", "매력적인 발표자", "논리적 설명자", "감동적인 스토리텔러",
	"효과적인 커뮤니케이터", "다언어 소통 전문가", "브랜드 스토리텔러", "콘텐츠 크리에이터", "영향력있는 인플루언서",
	"비주얼 커뮤니케이터", "디지털 마케터", "PR 전문가", "미디어 플래너", "소셜미디어 전문가",
	
	// 기업가정신 관련
	"혁신적 기업가", "위험감수형 모험가", "기회포착 전문가", "사업개발 전문가", "벤처 캐피탈리스트",
	"스타트업 창업자", "비즈니스 모델 혁신가", "시장창조형 개척자", "투자 전문가", "사업전략 수립자",
	
	// 감성 관련
	"감성적 커뮤니케이터", "공감능력이 높은 상담자", "따뜻한 서포터", "배려심깊은 도우미", "친화력있는 관계자",
	"감정이입형 조력자", "힐링형 치유자", "정서적 안정자", "위로형 동반자", "긍정에너지 전파자",
	
	// 글로벌 역량 관련
	"글로벌 마인드셋 소유자", "다문화 소통 전문가", "국제업무 전문가", "크로스 컬처 브릿지", "글로벌 프로젝트 매니저",
	"해외시장 개척자", "국제협력 전문가", "글로벌 네트워커", "다국적 팀 리더", "해외사업 개발자",
	
	// 지속가능성 관련
	"지속가능 발전 추구자", "ESG 전문가", "친환경 솔루션 개발자", "사회적 가치 창출자", "그린테크 개발자",
	"순환경제 설계자", "탄소중립 기획자", "사회적 기업가", "임팩트 메이커", "지속가능 혁신가",
}

// TitleGenerator 칭호 생성기
type TitleGenerator struct {
	rand *rand.Rand
}

// NewTitleGenerator 새로운 칭호 생성기 생성
func NewTitleGenerator() *TitleGenerator {
	return &TitleGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateSmartTitles 텍스트 내용을 분석하여 관련성 있는 칭호 생성
func (tg *TitleGenerator) GenerateSmartTitles(text string, count int) []string {
	if count <= 0 {
		count = 3
	}
	
	// 텍스트에서 키워드 추출 및 카테고리 매핑
	keywords := tg.extractKeywords(text)
	relevantTitles := tg.getRelevantTitles(keywords)
	
	// 관련성 있는 칭호가 충분하지 않으면 전체 풀에서 선택
	if len(relevantTitles) < count*2 {
		relevantTitles = append(relevantTitles, tg.getRandomTitles(count*2)...)
	}
	
	// 다양성을 고려하여 최종 선택
	return tg.selectDiverseTitles(relevantTitles, count)
}

// extractKeywords 텍스트에서 핵심 키워드 추출
func (tg *TitleGenerator) extractKeywords(text string) []string {
	text = strings.ToLower(text)
	keywords := []string{}
	
	// 리더십 관련 키워드
	if strings.Contains(text, "리더") || strings.Contains(text, "지도") || strings.Contains(text, "이끌") || 
	   strings.Contains(text, "주도") || strings.Contains(text, "팀장") || strings.Contains(text, "관리") {
		keywords = append(keywords, "leadership")
	}
	
	// 창의성 관련 키워드
	if strings.Contains(text, "창의") || strings.Contains(text, "아이디어") || strings.Contains(text, "혁신") || 
	   strings.Contains(text, "독창") || strings.Contains(text, "상상") || strings.Contains(text, "기획") {
		keywords = append(keywords, "creativity")
	}
	
	// 기술 관련 키워드
	if strings.Contains(text, "개발") || strings.Contains(text, "프로그래밍") || strings.Contains(text, "코딩") ||
	   strings.Contains(text, "기술") || strings.Contains(text, "엔지니어") || strings.Contains(text, "시스템") {
		keywords = append(keywords, "technology")
	}
	
	// 분석 관련 키워드
	if strings.Contains(text, "분석") || strings.Contains(text, "데이터") || strings.Contains(text, "논리") ||
	   strings.Contains(text, "체계") || strings.Contains(text, "정확") || strings.Contains(text, "꼼꼼") {
		keywords = append(keywords, "analysis")
	}
	
	// 협력 관련 키워드
	if strings.Contains(text, "협력") || strings.Contains(text, "소통") || strings.Contains(text, "팀워크") ||
	   strings.Contains(text, "화합") || strings.Contains(text, "관계") || strings.Contains(text, "네트워킹") {
		keywords = append(keywords, "collaboration")
	}
	
	// 실행력 관련 키워드
	if strings.Contains(text, "실행") || strings.Contains(text, "추진") || strings.Contains(text, "완수") ||
	   strings.Contains(text, "성과") || strings.Contains(text, "목표") || strings.Contains(text, "결과") {
		keywords = append(keywords, "execution")
	}
	
	// 커뮤니케이션 관련 키워드
	if strings.Contains(text, "발표") || strings.Contains(text, "설득") || strings.Contains(text, "전달") ||
	   strings.Contains(text, "스피치") || strings.Contains(text, "커뮤니케이션") || strings.Contains(text, "소통") {
		keywords = append(keywords, "communication")
	}
	
	// 기업가정신 관련 키워드
	if strings.Contains(text, "창업") || strings.Contains(text, "사업") || strings.Contains(text, "기업가") ||
	   strings.Contains(text, "투자") || strings.Contains(text, "벤처") || strings.Contains(text, "스타트업") {
		keywords = append(keywords, "entrepreneurship")
	}
	
	return keywords
}

// getRelevantTitles 키워드에 기반하여 관련성 있는 칭호 반환
func (tg *TitleGenerator) getRelevantTitles(keywords []string) []string {
	titleMap := map[string][]string{
		"leadership": {
			"적극적 리더", "주도적 지도자", "카리스마틱 리더", "혁신적 리더", "변화지향적 리더",
			"전략적 지휘자", "비전제시형 리더", "결단력있는 지도자", "포용적 리더", "멘토링형 지도자",
		},
		"creativity": {
			"창의적 혁신가", "독창적 아이디어맨", "상상력풍부한 기획자", "아이디어 뱅크형 인재", "혁신적 사고자",
			"창조적 문제해결사", "독특한 발상가", "트렌드세터형 개발자", "새로운 관점의 제안자", "브레인스토밍 마스터",
		},
		"technology": {
			"기술지향적 개발자", "혁신적 엔지니어", "문제해결형 프로그래머", "효율성추구 개발자", "품질중심 아키텍트",
			"풀스택 개발자", "DevOps 엔지니어", "AI/ML 전문가", "블록체인 개발자", "클라우드 전문가",
		},
		"analysis": {
			"논리적 분석가", "체계적 사고자", "데이터기반 분석가", "꼼꼼한 검토자", "세밀한 관찰자",
			"정확한 판단자", "신중한 의사결정자", "비판적 사고자", "통계적 분석가", "심층적 탐구자",
		},
		"collaboration": {
			"협력적 팀워커", "소통지향적 브릿지", "화합형 조정자", "포용적 협력자", "네트워킹 전문가",
			"관계지향적 커뮤니케이터", "상호작용 전문가", "팀워크 촉진자", "협력적 문제해결사", "네트워킹 마스터",
		},
		"execution": {
			"실행력있는 추진자", "끈기있는 완주자", "목표지향적 달성자", "집중력있는 실행자", "결과지향적 완수자",
			"신속한 실행자", "성과중심 달성자", "효율성 추구자", "생산성 최적화 전문가", "프로세스 개선자",
		},
		"communication": {
			"설득력있는 스피커", "명확한 전달자", "매력적인 발표자", "논리적 설명자", "감동적인 스토리텔러",
			"효과적인 커뮤니케이터", "브랜드 스토리텔러", "콘텐츠 크리에이터", "영향력있는 인플루언서", "PR 전문가",
		},
		"entrepreneurship": {
			"혁신적 기업가", "위험감수형 모험가", "기회포착 전문가", "사업개발 전문가", "벤처 캐피탈리스트",
			"스타트업 창업자", "비즈니스 모델 혁신가", "시장창조형 개척자", "투자 전문가", "사업전략 수립자",
		},
	}
	
	var relevantTitles []string
	for _, keyword := range keywords {
		if titles, exists := titleMap[keyword]; exists {
			relevantTitles = append(relevantTitles, titles...)
		}
	}
	
	return tg.removeDuplicates(relevantTitles)
}

// getRandomTitles 전체 풀에서 랜덤하게 칭호 선택
func (tg *TitleGenerator) getRandomTitles(count int) []string {
	if count >= len(enhancedTitlePool) {
		return enhancedTitlePool
	}
	
	// Fisher-Yates 셔플 알고리즘
	shuffled := make([]string, len(enhancedTitlePool))
	copy(shuffled, enhancedTitlePool)
	
	for i := len(shuffled) - 1; i > 0; i-- {
		j := tg.rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	
	return shuffled[:count]
}

// selectDiverseTitles 다양성을 고려하여 최종 칭호 선택
func (tg *TitleGenerator) selectDiverseTitles(candidates []string, count int) []string {
	if len(candidates) <= count {
		return candidates
	}
	
	selected := []string{}
	remaining := make([]string, len(candidates))
	copy(remaining, candidates)
	
	// 첫 번째는 랜덤하게 선택
	firstIdx := tg.rand.Intn(len(remaining))
	selected = append(selected, remaining[firstIdx])
	remaining = append(remaining[:firstIdx], remaining[firstIdx+1:]...)
	
	// 나머지는 다양성을 고려하여 선택
	for len(selected) < count && len(remaining) > 0 {
		bestIdx := 0
		bestScore := -1.0
		
		for i, candidate := range remaining {
			diversityScore := tg.calculateDiversityScore(candidate, selected)
			if diversityScore > bestScore {
				bestScore = diversityScore
				bestIdx = i
			}
		}
		
		selected = append(selected, remaining[bestIdx])
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}
	
	return selected
}

// calculateDiversityScore 다양성 점수 계산
func (tg *TitleGenerator) calculateDiversityScore(candidate string, selected []string) float64 {
	if len(selected) == 0 {
		return 1.0
	}
	
	minSimilarity := 1.0
	candidateWords := strings.Fields(candidate)
	
	for _, sel := range selected {
		selectedWords := strings.Fields(sel)
		similarity := tg.calculateJaccardSimilarity(candidateWords, selectedWords)
		if similarity < minSimilarity {
			minSimilarity = similarity
		}
	}
	
	return 1.0 - minSimilarity // 유사도가 낮을수록 다양성 점수가 높음
}

// calculateJaccardSimilarity Jaccard 유사도 계산
func (tg *TitleGenerator) calculateJaccardSimilarity(set1, set2 []string) float64 {
	if len(set1) == 0 && len(set2) == 0 {
		return 1.0
	}
	
	intersection := 0
	set1Map := make(map[string]bool)
	for _, word := range set1 {
		set1Map[word] = true
	}
	
	for _, word := range set2 {
		if set1Map[word] {
			intersection++
		}
	}
	
	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(union)
}

// removeDuplicates 중복 제거
func (tg *TitleGenerator) removeDuplicates(titles []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, title := range titles {
		if !seen[title] {
			seen[title] = true
			result = append(result, title)
		}
	}
	
	return result
}