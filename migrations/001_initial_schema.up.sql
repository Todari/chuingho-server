-- 초기 데이터베이스 스키마 생성

-- UUID 확장 활성화
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 사용자 테이블
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 자기소개서 테이블
CREATE TABLE resumes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    original_filename VARCHAR(255) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    storage_key VARCHAR(500) NOT NULL, -- S3 객체 키
    content_hash VARCHAR(64) NOT NULL, -- SHA-256 해시
    status VARCHAR(50) NOT NULL DEFAULT 'uploaded', -- uploaded, processing, completed, failed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 췽호 추천 결과 테이블
CREATE TABLE title_recommendations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    resume_id UUID NOT NULL REFERENCES resumes(id) ON DELETE CASCADE,
    titles JSONB NOT NULL, -- 추천된 3개 췽호 배열
    vector_similarity_scores JSONB, -- 각 췽호의 유사도 점수
    processing_time_ms INTEGER, -- 처리 시간 (밀리초)
    ml_model_version VARCHAR(100), -- 사용된 ML 모델 버전
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 처리 로그 테이블 (개인정보 제거된 로그)
CREATE TABLE processing_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id VARCHAR(100) NOT NULL, -- 요청 추적 ID
    user_id_hash VARCHAR(64), -- 사용자 ID 해시 (개인정보 보호)
    operation VARCHAR(100) NOT NULL, -- upload, recommend, etc.
    status VARCHAR(50) NOT NULL, -- success, failed, processing
    error_message TEXT,
    processing_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 형용사+명사 후보 테이블 (벡터 DB와 별도 메타데이터)
CREATE TABLE phrase_candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phrase VARCHAR(100) NOT NULL UNIQUE,
    adjective VARCHAR(50) NOT NULL,
    noun VARCHAR(50) NOT NULL,
    frequency_score FLOAT DEFAULT 0, -- 코퍼스에서의 빈도
    semantic_category VARCHAR(100), -- 의미 카테고리 (창의성, 리더십 등)
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 인덱스 생성
CREATE INDEX idx_resumes_user_id ON resumes(user_id);
CREATE INDEX idx_resumes_status ON resumes(status);
CREATE INDEX idx_resumes_created_at ON resumes(created_at);

CREATE INDEX idx_title_recommendations_resume_id ON title_recommendations(resume_id);
CREATE INDEX idx_title_recommendations_created_at ON title_recommendations(created_at);

CREATE INDEX idx_processing_logs_request_id ON processing_logs(request_id);
CREATE INDEX idx_processing_logs_user_id_hash ON processing_logs(user_id_hash);
CREATE INDEX idx_processing_logs_operation ON processing_logs(operation);
CREATE INDEX idx_processing_logs_created_at ON processing_logs(created_at);

CREATE INDEX idx_phrase_candidates_phrase ON phrase_candidates(phrase);
CREATE INDEX idx_phrase_candidates_adjective ON phrase_candidates(adjective);
CREATE INDEX idx_phrase_candidates_noun ON phrase_candidates(noun);
CREATE INDEX idx_phrase_candidates_is_active ON phrase_candidates(is_active);
CREATE INDEX idx_phrase_candidates_semantic_category ON phrase_candidates(semantic_category);

-- 트리거를 위한 함수: updated_at 자동 업데이트
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- updated_at 자동 업데이트 트리거
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_resumes_updated_at BEFORE UPDATE ON resumes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_phrase_candidates_updated_at BEFORE UPDATE ON phrase_candidates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();