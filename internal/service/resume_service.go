package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/database"
	"github.com/Todari/chuingho-server/pkg/model"
	"github.com/Todari/chuingho-server/pkg/util"
)

// ResumeService 자기소개서 관련 비즈니스 로직
type ResumeService struct {
	db     *database.DB
	logger *zap.Logger
}

// NewResumeService 새로운 자기소개서 서비스 생성
func NewResumeService(db *database.DB, logger *zap.Logger) *ResumeService {
	return &ResumeService{
		db:     db,
		logger: logger,
	}
}

// UploadResume 자기소개서 업로드
func (s *ResumeService) UploadResume(ctx context.Context, text string) (*model.UploadResumeResponse, error) {
	requestID, _ := util.GenerateRequestID()
	s.logger.Info("자기소개서 업로드 시작",
		zap.String("request_id", requestID),
		zap.Int("text_length", len([]rune(text))))

	// 텍스트 정리
	cleanedText := util.CleanText(text)
	if len(cleanedText) < 10 {
		return nil, fmt.Errorf("자기소개서 내용이 너무 짧습니다 (최소 10자)")
	}

	// 트랜잭션 시작
	tx, err := s.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("트랜잭션 시작 실패: %w", err)
	}
	defer tx.Rollback(ctx)

	// 사용자 생성 (임시 사용자)
	userID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO users (id, created_at, updated_at) 
		VALUES ($1, NOW(), NOW())`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("사용자 생성 실패: %w", err)
	}

	// 텍스트 해시 생성
	contentHash := util.HashContent([]byte(cleanedText))

	// 자기소개서 저장
	resumeID := uuid.New()
	_, err = tx.Exec(ctx, `
		INSERT INTO resumes (
			id, user_id, content, content_hash, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`,
		resumeID, userID, cleanedText, contentHash, model.ResumeStatusUploaded)
	if err != nil {
		return nil, fmt.Errorf("자기소개서 저장 실패: %w", err)
	}

	// 처리 로그 저장
	_, err = tx.Exec(ctx, `
		INSERT INTO processing_logs (
			request_id, user_id_hash, operation, status, created_at
		) VALUES ($1, $2, $3, $4, NOW())`,
		requestID, util.HashUserID(userID.String()), "upload", "success")
	if err != nil {
		s.logger.Warn("처리 로그 저장 실패", zap.Error(err))
	}

	// 트랜잭션 커밋
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("트랜잭션 커밋 실패: %w", err)
	}

	s.logger.Info("자기소개서 업로드 완료",
		zap.String("request_id", requestID),
		zap.String("resume_id", resumeID.String()),
		zap.String("content_hash", contentHash))

	return &model.UploadResumeResponse{
		ResumeID: resumeID,
		Status:   model.ResumeStatusUploaded,
	}, nil
}

// GetResume 자기소개서 조회
func (s *ResumeService) GetResume(ctx context.Context, resumeID uuid.UUID) (*model.Resume, error) {
	var resume model.Resume
	err := s.db.Pool.QueryRow(ctx, `
		SELECT id, user_id, content, content_hash, status, created_at, updated_at
		FROM resumes 
		WHERE id = $1`,
		resumeID).Scan(
		&resume.ID, &resume.UserID, &resume.Content, &resume.ContentHash,
		&resume.Status, &resume.CreatedAt, &resume.UpdatedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("자기소개서를 찾을 수 없습니다: %s", resumeID.String())
		}
		return nil, fmt.Errorf("자기소개서 조회 실패: %w", err)
	}

	return &resume, nil
}

// GetResumeContent 자기소개서 내용 조회
func (s *ResumeService) GetResumeContent(ctx context.Context, resumeID uuid.UUID) (string, error) {
	// 자기소개서 조회
	resume, err := s.GetResume(ctx, resumeID)
	if err != nil {
		return "", err
	}

	s.logger.Debug("자기소개서 내용 조회 완료",
		zap.String("resume_id", resumeID.String()),
		zap.Int("content_length", len(resume.Content)))

	return resume.Content, nil
}

// UpdateResumeStatus 자기소개서 상태 업데이트
func (s *ResumeService) UpdateResumeStatus(ctx context.Context, resumeID uuid.UUID, status model.ResumeStatus) error {
	_, err := s.db.Pool.Exec(ctx, `
		UPDATE resumes 
		SET status = $1, updated_at = NOW() 
		WHERE id = $2`,
		status, resumeID)

	if err != nil {
		return fmt.Errorf("자기소개서 상태 업데이트 실패: %w", err)
	}

	s.logger.Debug("자기소개서 상태 업데이트",
		zap.String("resume_id", resumeID.String()),
		zap.String("status", string(status)))

	return nil
}

// ListResumes 자기소개서 목록 조회 (관리용)
func (s *ResumeService) ListResumes(ctx context.Context, limit, offset int) ([]model.Resume, error) {
	rows, err := s.db.Pool.Query(ctx, `
		SELECT id, user_id, content, content_hash, status, created_at, updated_at
		FROM resumes 
		ORDER BY created_at DESC 
		LIMIT $1 OFFSET $2`,
		limit, offset)

	if err != nil {
		return nil, fmt.Errorf("자기소개서 목록 조회 실패: %w", err)
	}
	defer rows.Close()

	var resumes []model.Resume
	for rows.Next() {
		var resume model.Resume
		err := rows.Scan(
			&resume.ID, &resume.UserID, &resume.Content, &resume.ContentHash,
			&resume.Status, &resume.CreatedAt, &resume.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("자기소개서 행 스캔 실패: %w", err)
		}
		resumes = append(resumes, resume)
	}

	return resumes, nil
}