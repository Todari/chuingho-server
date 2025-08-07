package util

import (
	"testing"
)

func TestExtractText(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		filename string
		expected string
		hasError bool
	}{
		{
			name:     "txt 파일",
			content:  []byte("안녕하세요. 테스트 텍스트입니다."),
			filename: "test.txt",
			expected: "안녕하세요. 테스트 텍스트입니다.",
			hasError: false,
		},
		{
			name:     "md 파일",
			content:  []byte("# 제목\n\n내용입니다."),
			filename: "test.md",
			expected: "# 제목\n\n내용입니다.",
			hasError: false,
		},
		{
			name:     "docx 파일 (간단 처리)",
			content:  []byte("문서 내용"),
			filename: "test.docx",
			expected: "문서 내용",
			hasError: false,
		},
		{
			name:     "지원하지 않는 확장자",
			content:  []byte("내용"),
			filename: "test.pdf",
			expected: "내용", // UTF-8 유효하면 반환
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractText(tt.content, tt.filename)
			
			if tt.hasError && err == nil {
				t.Error("ExtractText()에서 에러가 예상되었지만 nil을 반환")
			}
			
			if !tt.hasError && err != nil {
				t.Errorf("ExtractText() 예상치 못한 에러 = %v", err)
			}
			
			if result != tt.expected {
				t.Errorf("ExtractText() = %v, 예상 = %v", result, tt.expected)
			}
		})
	}
}

func TestCleanText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "연속된 공백 제거",
			input:    "안녕하세요    세계입니다",
			expected: "안녕하세요 세계입니다",
		},
		{
			name:     "앞뒤 공백 제거",
			input:    "  텍스트  ",
			expected: "텍스트",
		},
		{
			name:     "탭과 개행 처리",
			input:    "라인1\n\n\t라인2\t\t라인3",
			expected: "라인1 라인2 라인3",
		},
		{
			name:     "이미 정리된 텍스트",
			input:    "정상적인 텍스트",
			expected: "정상적인 텍스트",
		},
		{
			name:     "빈 문자열",
			input:    "",
			expected: "",
		},
		{
			name:     "공백만 있는 문자열",
			input:    "   \t\n   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanText(tt.input)
			if result != tt.expected {
				t.Errorf("CleanText() = %q, 예상 = %q", result, tt.expected)
			}
		})
	}
}

func TestGetMimeType(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "txt 파일",
			filename: "test.txt",
			expected: "text/plain",
		},
		{
			name:     "md 파일",
			filename: "readme.md",
			expected: "text/plain", // Deprecated: 이제 모든 파일이 text/plain
		},
		{
			name:     "docx 파일",
			filename: "document.docx",
			expected: "text/plain", // Deprecated: 이제 모든 파일이 text/plain
		},
		{
			name:     "알 수 없는 확장자",
			filename: "unknown.unknown",
			expected: "text/plain", // Deprecated: 이제 모든 파일이 text/plain
		},
		{
			name:     "확장자 없음",
			filename: "filename",
			expected: "text/plain", // Deprecated: 이제 모든 파일이 text/plain
		},
		{
			name:     "대문자 확장자",
			filename: "test.TXT",
			expected: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetMimeType(tt.filename)
			if result != tt.expected {
				t.Errorf("GetMimeType() = %v, 예상 = %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidTextFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected bool
	}{
		{
			name:     "유효한 txt 파일",
			filename: "test.txt",
			expected: true,
		},
		{
			name:     "유효한 md 파일",
			filename: "readme.md",
			expected: true,
		},
		{
			name:     "유효한 docx 파일",
			filename: "document.docx",
			expected: true,
		},
		{
			name:     "유효하지 않은 pdf 파일",
			filename: "document.pdf",
			expected: true, // Deprecated: 이제 모든 파일이 유효함
		},
		{
			name:     "유효하지 않은 이미지 파일",
			filename: "image.jpg",
			expected: true, // Deprecated: 이제 모든 파일이 유효함
		},
		{
			name:     "확장자 없음",
			filename: "filename",
			expected: true, // Deprecated: 이제 모든 파일이 유효함
		},
		{
			name:     "대문자 확장자",
			filename: "test.TXT",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidTextFile(tt.filename)
			if result != tt.expected {
				t.Errorf("IsValidTextFile() = %v, 예상 = %v", result, tt.expected)
			}
		})
	}
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		maxLength int
		expected  string
	}{
		{
			name:      "짧은 텍스트",
			text:      "짧은 텍스트",
			maxLength: 100,
			expected:  "짧은 텍스트",
		},
		{
			name:      "긴 영어 텍스트",
			text:      "This is a very long text that should be truncated",
			maxLength: 20,
			expected:  "This is a very lo...",
		},
		{
			name:      "긴 한글 텍스트",
			text:      "이것은 매우 긴 한글 텍스트입니다 잘려야 합니다",
			maxLength: 10,
			expected:  "이것은 매우 ...",
		},
		{
			name:      "정확히 최대 길이",
			text:      "정확히열글자입니다",
			maxLength: 10,
			expected:  "정확히열글자입니다",
		},
		{
			name:      "빈 문자열",
			text:      "",
			maxLength: 10,
			expected:  "",
		},
		{
			name:      "최대 길이 0",
			text:      "텍스트",
			maxLength: 0,
			expected:  "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateText(tt.text, tt.maxLength)
			if result != tt.expected {
				t.Errorf("TruncateText() = %q, 예상 = %q", result, tt.expected)
			}
		})
	}
}