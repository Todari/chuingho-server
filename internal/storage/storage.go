package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/config"
)

// Storage S3 호환 객체 스토리지 클라이언트
type Storage struct {
	client     *minio.Client
	bucketName string
	config     config.StorageConfig
	logger     *zap.Logger
}

// UploadResult 업로드 결과
type UploadResult struct {
	Key         string `json:"key"`
	Size        int64  `json:"size"`
	ContentHash string `json:"content_hash"`
	ETag        string `json:"etag"`
}

// New 새로운 스토리지 클라이언트 생성
func New(ctx context.Context, cfg config.StorageConfig, logger *zap.Logger) (*Storage, error) {
	// MinIO 클라이언트 생성
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("MinIO 클라이언트 생성 실패: %w", err)
	}

	storage := &Storage{
		client:     client,
		bucketName: cfg.BucketName,
		config:     cfg,
		logger:     logger,
	}

	// 버킷 존재 확인 및 생성
	if err := storage.ensureBucket(ctx); err != nil {
		return nil, fmt.Errorf("버킷 확인/생성 실패: %w", err)
	}

	logger.Info("스토리지 클라이언트 초기화 완료",
		zap.String("endpoint", cfg.Endpoint),
		zap.String("bucket", cfg.BucketName),
		zap.Bool("ssl", cfg.UseSSL))

	return storage, nil
}

// ensureBucket 버킷 존재 확인 및 생성
func (s *Storage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("버킷 존재 확인 실패: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{
			Region: s.config.Region,
		})
		if err != nil {
			return fmt.Errorf("버킷 생성 실패: %w", err)
		}

		s.logger.Info("새 버킷 생성됨", zap.String("bucket", s.bucketName))
	}

	// 서버측 암호화 설정 (AES-256)
	if err := s.setupEncryption(ctx); err != nil {
		s.logger.Warn("서버측 암호화 설정 실패", zap.Error(err))
	}

	return nil
}

// setupEncryption 버킷 서버측 암호화 설정
func (s *Storage) setupEncryption(ctx context.Context) error {
	// MinIO는 기본적으로 AES-256-GCM 암호화를 지원
	// 추가적인 암호화 정책 설정은 MinIO 관리자 인터페이스에서 수행
	return nil
}

// UploadFile 파일 업로드
func (s *Storage) UploadFile(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*UploadResult, error) {
	// 컨텐츠 해시 계산을 위한 TeeReader 사용
	hasher := sha256.New()
	teeReader := io.TeeReader(reader, hasher)

	// 업로드 옵션 설정
	options := minio.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"uploaded-at": time.Now().UTC().Format(time.RFC3339),
		},
		ServerSideEncryption: nil, // MinIO 기본 암호화 사용
	}

	// 파일 업로드
	info, err := s.client.PutObject(ctx, s.bucketName, key, teeReader, size, options)
	if err != nil {
		return nil, fmt.Errorf("파일 업로드 실패: %w", err)
	}

	// 해시 값 계산
	contentHash := fmt.Sprintf("%x", hasher.Sum(nil))

	result := &UploadResult{
		Key:         key,
		Size:        info.Size,
		ContentHash: contentHash,
		ETag:        strings.Trim(info.ETag, "\""),
	}

	s.logger.Info("파일 업로드 완료",
		zap.String("key", key),
		zap.Int64("size", info.Size),
		zap.String("content_type", contentType))

	return result, nil
}

// DownloadFile 파일 다운로드
func (s *Storage) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("파일 다운로드 실패: %w", err)
	}

	// 객체 존재 확인을 위해 Stat 호출
	_, err = object.Stat()
	if err != nil {
		object.Close()
		return nil, fmt.Errorf("파일 정보 확인 실패: %w", err)
	}

	s.logger.Debug("파일 다운로드 시작", zap.String("key", key))
	return object, nil
}

// GetPresignedURL 미리 서명된 URL 생성
func (s *Storage) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	reqParams := make(url.Values)
	presignedURL, err := s.client.PresignedGetObject(ctx, s.bucketName, key, expires, reqParams)
	if err != nil {
		return "", fmt.Errorf("미리 서명된 URL 생성 실패: %w", err)
	}

	s.logger.Debug("미리 서명된 URL 생성",
		zap.String("key", key),
		zap.Duration("expires", expires))

	return presignedURL.String(), nil
}

// DeleteFile 파일 삭제
func (s *Storage) DeleteFile(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("파일 삭제 실패: %w", err)
	}

	s.logger.Info("파일 삭제 완료", zap.String("key", key))
	return nil
}

// FileExists 파일 존재 확인
func (s *Storage) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("파일 존재 확인 실패: %w", err)
	}
	return true, nil
}

// GetFileInfo 파일 정보 조회
func (s *Storage) GetFileInfo(ctx context.Context, key string) (*minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("파일 정보 조회 실패: %w", err)
	}
	return &info, nil
}

// ListFiles 파일 목록 조회
func (s *Storage) ListFiles(ctx context.Context, prefix string, recursive bool) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo

	objectCh := s.client.ListObjects(ctx, s.bucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("파일 목록 조회 실패: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// HealthCheck 스토리지 상태 확인
func (s *Storage) HealthCheck(ctx context.Context) error {
	// 버킷 존재 확인으로 헬스체크 수행
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("스토리지 헬스체크 실패: %w", err)
	}
	if !exists {
		return fmt.Errorf("버킷이 존재하지 않음: %s", s.bucketName)
	}
	return nil
}

// GenerateKey 객체 키 생성 (타임스탬프 기반 계층 구조)
func GenerateKey(prefix, filename string) string {
	now := time.Now().UTC()
	datePrefix := now.Format("2006/01/02")
	
	// 파일명에서 확장자 분리
	parts := strings.Split(filename, ".")
	name := parts[0]
	ext := ""
	if len(parts) > 1 {
		ext = "." + parts[len(parts)-1]
	}
	
	// 타임스탬프와 함께 유니크한 키 생성
	timestamp := now.Format("150405")
	key := fmt.Sprintf("%s/%s/%s_%s%s", prefix, datePrefix, name, timestamp, ext)
	
	return key
}