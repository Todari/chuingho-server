package util

import (
	"regexp"
	"strings"
)

// ExtractText 파일 확장자에 따라 텍스트 추출 (Deprecated: 텍스트 입력 방식으로 변경됨)
// 기존 코드 호환성을 위해 유지, 새로운 코드에서는 사용하지 말 것
func ExtractText(content []byte, filename string) (string, error) {
	// 간단히 바이트를 문자열로 변환
	return string(content), nil
}

// CleanText 텍스트 정리 및 정규화
func CleanText(text string) string {
	// 연속된 공백 제거
	re := regexp.MustCompile(`\s+`)
	cleaned := re.ReplaceAllString(text, " ")
	
	// 앞뒤 공백 제거
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// GetMimeType 파일명에서 MIME 타입 추정 (Deprecated: 텍스트 입력 방식으로 변경됨)
// 기존 코드 호환성을 위해 유지, 새로운 코드에서는 사용하지 말 것
func GetMimeType(filename string) string {
	return "text/plain"
}

// IsValidTextFile 유효한 텍스트 파일인지 확인 (Deprecated: 텍스트 입력 방식으로 변경됨)
// 기존 코드 호환성을 위해 유지, 새로운 코드에서는 사용하지 말 것
func IsValidTextFile(filename string) bool {
	return true
}

// TruncateText 텍스트를 지정된 길이로 자르기
func TruncateText(text string, maxLength int) string {
	runes := []rune(text)
	if len(runes) <= maxLength {
		return text
	}
	
	if maxLength <= 3 {
		return "..."
	}
	
	return string(runes[:maxLength-3]) + "..."
}