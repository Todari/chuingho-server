-- 파일 기반에서 텍스트 기반으로 자기소개서 테이블 업데이트

-- 기존 resumes 테이블 백업 (롤백용)
CREATE TABLE resumes_backup AS SELECT * FROM resumes;

-- 새로운 컬럼 추가
ALTER TABLE resumes 
ADD COLUMN content TEXT;

-- 기존 파일 관련 컬럼들을 optional로 변경
ALTER TABLE resumes 
ALTER COLUMN original_filename DROP NOT NULL,
ALTER COLUMN file_size DROP NOT NULL,
ALTER COLUMN mime_type DROP NOT NULL,
ALTER COLUMN storage_key DROP NOT NULL;

-- 기존 데이터가 있다면 마이그레이션 로직 (파일에서 텍스트 추출 필요)
-- 실제 프로덕션에서는 데이터 마이그레이션 스크립트가 필요함

-- 새로운 제약조건 추가 (content 또는 storage_key 중 하나는 반드시 존재)
ALTER TABLE resumes 
ADD CONSTRAINT resumes_content_or_file_check 
CHECK (
    (content IS NOT NULL AND LENGTH(content) >= 10) OR 
    (storage_key IS NOT NULL)
);

-- 인덱스 추가
CREATE INDEX idx_resumes_content_hash ON resumes(content_hash);

-- 주석 업데이트
COMMENT ON TABLE resumes IS '자기소개서 테이블 - 텍스트 기반 저장 지원';
COMMENT ON COLUMN resumes.content IS '자기소개서 텍스트 내용 (직접 입력)';
COMMENT ON COLUMN resumes.original_filename IS '원본 파일명 (파일 업로드시에만)';
COMMENT ON COLUMN resumes.file_size IS '파일 크기 (파일 업로드시에만)';
COMMENT ON COLUMN resumes.mime_type IS '파일 MIME 타입 (파일 업로드시에만)';
COMMENT ON COLUMN resumes.storage_key IS 'S3 저장소 키 (파일 업로드시에만)';
COMMENT ON COLUMN resumes.content_hash IS '콘텐츠 해시 (텍스트 또는 파일 내용)';