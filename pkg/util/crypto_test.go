package util

import (
	"strings"
	"testing"
)

func TestHashUserID(t *testing.T) {
	tests := []struct {
		name     string
		userID   string
		expected int // 예상 해시 길이
	}{
		{
			name:     "일반 UUID",
			userID:   "123e4567-e89b-12d3-a456-426614174000",
			expected: 64, // SHA-256 hex는 64글자
		},
		{
			name:     "빈 문자열",
			userID:   "",
			expected: 64,
		},
		{
			name:     "한글 사용자 ID",
			userID:   "사용자123",
			expected: 64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashUserID(tt.userID)
			
			if len(result) != tt.expected {
				t.Errorf("HashUserID() 길이 = %v, 예상 = %v", len(result), tt.expected)
			}
			
			// 16진수 문자만 포함하는지 확인
			for _, char := range result {
				if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f')) {
					t.Errorf("HashUserID() 잘못된 16진수 문자: %c", char)
				}
			}
		})
	}
}

func TestHashUserIDConsistency(t *testing.T) {
	userID := "test-user-123"
	
	hash1 := HashUserID(userID)
	hash2 := HashUserID(userID)
	
	if hash1 != hash2 {
		t.Error("HashUserID()는 같은 입력에 대해 동일한 결과를 반환해야 함")
	}
}

func TestHashUserIDUniqueness(t *testing.T) {
	userID1 := "user1"
	userID2 := "user2"
	
	hash1 := HashUserID(userID1)
	hash2 := HashUserID(userID2)
	
	if hash1 == hash2 {
		t.Error("HashUserID()는 다른 입력에 대해 다른 결과를 반환해야 함")
	}
}

func TestHashContent(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{
			name:    "일반 텍스트",
			content: []byte("Hello, World!"),
		},
		{
			name:    "한글 텍스트",
			content: []byte("안녕하세요, 세계!"),
		},
		{
			name:    "빈 내용",
			content: []byte(""),
		},
		{
			name:    "바이너리 데이터",
			content: []byte{0x00, 0x01, 0x02, 0xFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HashContent(tt.content)
			
			if len(result) != 64 { // SHA-256 hex
				t.Errorf("HashContent() 길이 = %v, 예상 = 64", len(result))
			}
		})
	}
}

func TestGenerateRequestID(t *testing.T) {
	// 여러 번 생성하여 유니크한지 확인
	ids := make(map[string]bool)
	
	for i := 0; i < 100; i++ {
		id, err := GenerateRequestID()
		if err != nil {
			t.Fatalf("GenerateRequestID() 에러 = %v", err)
		}
		
		if len(id) != 32 { // 16바이트 hex = 32글자
			t.Errorf("GenerateRequestID() 길이 = %v, 예상 = 32", len(id))
		}
		
		if ids[id] {
			t.Errorf("GenerateRequestID() 중복 ID 생성: %s", id)
		}
		ids[id] = true
	}
}

func TestCalculateFileHash(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "일반 파일",
			content: "This is a test file content.",
		},
		{
			name:    "한글 파일",
			content: "이것은 테스트 파일 내용입니다.",
		},
		{
			name:    "빈 파일",
			content: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			hash, err := CalculateFileHash(reader)
			
			if err != nil {
				t.Errorf("CalculateFileHash() 에러 = %v", err)
			}
			
			if len(hash) != 64 {
				t.Errorf("CalculateFileHash() 길이 = %v, 예상 = 64", len(hash))
			}
			
			// 같은 내용은 같은 해시를 생성해야 함
			reader2 := strings.NewReader(tt.content)
			hash2, err2 := CalculateFileHash(reader2)
			
			if err2 != nil {
				t.Errorf("CalculateFileHash() 두 번째 호출 에러 = %v", err2)
			}
			
			if hash != hash2 {
				t.Error("CalculateFileHash()는 같은 내용에 대해 동일한 해시를 반환해야 함")
			}
		})
	}
}