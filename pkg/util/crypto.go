package util

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
)

// HashUserID 사용자 ID를 해시화 (개인정보 보호)
func HashUserID(userID string) string {
	hasher := sha256.New()
	hasher.Write([]byte(userID))
	return hex.EncodeToString(hasher.Sum(nil))
}

// HashContent 컨텐츠 해시 생성 (무결성 검증)
func HashContent(content []byte) string {
	hasher := sha256.New()
	hasher.Write(content)
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateRequestID 요청 추적용 고유 ID 생성
func GenerateRequestID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("요청 ID 생성 실패: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CalculateFileHash 파일 내용의 해시 계산
func CalculateFileHash(reader io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", fmt.Errorf("파일 해시 계산 실패: %w", err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}