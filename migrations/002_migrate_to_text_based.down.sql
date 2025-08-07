-- 텍스트 기반에서 파일 기반으로 롤백

-- 제약조건 제거
ALTER TABLE resumes DROP CONSTRAINT IF EXISTS resumes_content_or_file_check;

-- 인덱스 제거
DROP INDEX IF EXISTS idx_resumes_content_hash;

-- 새로운 컬럼 제거
ALTER TABLE resumes DROP COLUMN IF EXISTS content;

-- 기존 컬럼들을 다시 NOT NULL로 변경
ALTER TABLE resumes 
ALTER COLUMN original_filename SET NOT NULL,
ALTER COLUMN file_size SET NOT NULL,
ALTER COLUMN mime_type SET NOT NULL,
ALTER COLUMN storage_key SET NOT NULL;

-- 백업 테이블 제거
DROP TABLE IF EXISTS resumes_backup;