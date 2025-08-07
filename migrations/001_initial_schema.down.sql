-- 초기 스키마 롤백

-- 트리거 삭제
DROP TRIGGER IF EXISTS update_phrase_candidates_updated_at ON phrase_candidates;
DROP TRIGGER IF EXISTS update_resumes_updated_at ON resumes;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- 함수 삭제
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 테이블 삭제 (외래키 종속성 순서에 따라)
DROP TABLE IF EXISTS title_recommendations;
DROP TABLE IF EXISTS processing_logs;
DROP TABLE IF EXISTS phrase_candidates;
DROP TABLE IF EXISTS resumes;
DROP TABLE IF EXISTS users;

-- 확장 삭제 (다른 용도로 사용 중일 수 있으므로 주석 처리)
-- DROP EXTENSION IF EXISTS "uuid-ossp";