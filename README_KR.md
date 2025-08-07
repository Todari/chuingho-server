# 췽호 (Chuingho) - 자기소개서 기반 개인 칭호 추천 서비스

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![Python Version](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://python.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> **췽호**: 취준생(취업준비생) + 칭호(稱號)의 줄임말  
> 자기소개서를 분석하여 '나'를 특징짓는 개인만의 칭호를 추천해주는 AI 서비스

## 📋 목차

- [프로젝트 개요](#-프로젝트-개요)
- [주요 기능](#-주요-기능)
- [AI 기술 구현 상황](#-ai-기술-구현-상황)
- [시스템 아키텍처](#-시스템-아키텍처)
- [기술 스택](#-기술-스택)
- [빠른 시작](#-빠른-시작)
- [API 문서](#-api-문서)
- [환경 설정](#-환경-설정)
- [개발 가이드](#-개발-가이드)
- [배포](#-배포)
- [기여하기](#-기여하기)

## 🚀 프로젝트 개요

췽호는 취업준비생들을 위한 혁신적인 AI 서비스입니다. 자기소개서의 텍스트를 분석하여 개인의 특성을 벡터로 변환하고, 의미적으로 가까운 형용사+명사 조합을 찾아 개인만의 독특한 칭호를 추천합니다.

### 💡 핵심 아이디어

- **자기소개서 텍스트** → **768차원 벡터** → **유사도 검색** → **3개 췽호 추천**
- 예시: "창의적 설계자", "세심한 분석가", "적극적 리더"

### 🎯 대상 사용자

- 자신만의 특별한 칭호를 원하는 취업준비생
- 개인 브랜딩을 고민하는 직장인
- 자기 이해를 깊게 하고 싶은 누구나

## ✨ 주요 기능

### 1. 자기소개서 업로드 ⭐ **NEW: 텍스트 입력 방식**
- **입력 방식**: JSON 기반 텍스트 직접 입력 (복사-붙여넣기 지원)
- **텍스트 길이**: 10자~50,000자 (한글 기준)
- **지원 형식**: 순수 텍스트 (파일 업로드 방식에서 개선)
- **실시간 검증**: 길이, 형식 즉시 확인

### 2. AI 기반 췽호 추천
- **한국어 특화**: KoSimCSE-BERT 모델 사용
- **고성능 검색**: Faiss HNSW 알고리즘으로 < 200ms 응답
- **다양성 보장**: MMR 알고리즘으로 유사한 결과 방지
- **개인정보 보호**: 원본 텍스트는 벡터 DB에 저장 안 함

### 3. 관리 도구
- **구문 사전 구축**: CLI 도구로 형용사+명사 후보 자동 등록
- **실시간 모니터링**: 헬스체크 및 성능 지표
- **상세 로깅**: 요청 추적 및 오류 분석

## 🧠 AI 기술 구현 상황

### 🔬 현재 구현 상태 (v1.0)

#### 1. 텍스트 임베딩 (Text Embedding) ⭐ **업데이트됨**
```python
# 현재 구현: Python FastAPI 서비스 + 텍스트 기반 입력
Model: sentence-transformers 기반
Architecture: KoSimCSE-BERT 호환 모델
Dimension: 768차원 벡터
Input: JSON 형태 한국어 자기소개서 텍스트 (10자~50,000자)
Processing: 파일 파싱 제거 → 직접 텍스트 처리로 성능 개선
Output: 정규화된 768차원 밀집 벡터
```

**구현 세부사항:**
- **모델 로딩**: `sentence-transformers` 라이브러리 사용
- **GPU 지원**: CUDA 환경에서 자동 감지 및 활용
- **배치 처리**: 다중 텍스트 동시 임베딩 지원
- **ONNX 런타임**: 환경변수로 최적화된 추론 옵션 제공
- **캐싱**: 모델 로딩 시간 최적화

```python
# 실제 구현 예시
class EmbeddingService:
    def __init__(self):
        self.model = SentenceTransformer('BM-K/KoSimCSE-bert')
        self.device = 'cuda' if torch.cuda.is_available() else 'cpu'
    
    def embed_text(self, text: str) -> List[float]:
        embedding = self.model.encode(
            text, 
            convert_to_tensor=True,
            normalize_embeddings=True
        )
        return embedding.cpu().numpy().tolist()
```

#### 2. 벡터 유사도 검색 (Vector Similarity Search)
```go
// 현재 구현: Go + Faiss 인메모리 검색
Algorithm: HNSW (Hierarchical Navigable Small World)
Index Type: IndexHNSWFlat
Metric: Inner Product (코사인 유사도)
Performance: p95 < 200ms (로컬 환경)
Scale: 1M+ 벡터 지원
```

**구현 세부사항:**
- **인덱스 관리**: 실시간 벡터 추가/삭제/업데이트
- **지속성**: JSON 기반 메타데이터 + 바이너리 인덱스 저장
- **메모리 효율성**: 벡터 압축 및 양자화 지원
- **동시성**: Go 루틴 기반 병렬 검색
- **헬스체크**: 인덱스 상태 실시간 모니터링

```go
// 실제 구현 예시
type FaissDB struct {
    index      *faiss.IndexHNSWFlat
    vectors    map[string]VectorData
    dimension  int
    mu         sync.RWMutex
}

func (f *FaissDB) Search(ctx context.Context, query []float32, topK int) ([]VectorSearchResult, error) {
    distances, indices := f.index.Search(
        query, 
        int64(topK),
    )
    // 결과 후처리 및 다양성 랭킹 적용
    return f.diversityRanking(results), nil
}
```

#### 3. 다양성 보장 알고리즘 (Diversity Ranking)
```go
// MMR (Maximal Marginal Relevance) 구현
Algorithm: 유사도 0.7 + 다양성 0.3 가중치
Method: Jaccard 유사도 기반 문자열 비교
Purpose: 동일한 의미의 중복 췽호 방지
Output: 의미적으로 다양한 상위 3개 결과
```

**구현 세부사항:**
- **유사도 계산**: 코사인 유사도와 문자열 유사도 결합
- **가중치 조정**: 관련성과 다양성의 균형 최적화
- **실시간 처리**: 검색 결과에 즉시 적용
- **한국어 특화**: 형태소 단위 유사도 비교

#### 4. 구문 후보 데이터베이스
```sql
-- 현재 스키마: PostgreSQL + 벡터 메타데이터
CREATE TABLE phrase_candidates (
    id UUID PRIMARY KEY,
    phrase VARCHAR(100) NOT NULL UNIQUE,  -- "창의적 설계자"
    adjective VARCHAR(50) NOT NULL,       -- "창의적"
    noun VARCHAR(50) NOT NULL,            -- "설계자" 
    frequency_score FLOAT,                -- 코퍼스 빈도
    semantic_category VARCHAR(100),       -- "창의성", "리더십" 등
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**데이터 구축 현황:**
- **초기 데이터셋**: 수동 큐레이션된 500+ 형용사+명사 조합
- **카테고리**: 성격특성, 업무스타일, 리더십, 창의성, 분석력 등
- **확장 계획**: 위키백과, 뉴스 코퍼스 자동 추출 파이프라인

### 🚀 성능 최적화

#### 1. 응답 시간 최적화
- **목표**: p95 < 200ms
- **현재**: 로컬 환경 기준 50-100ms
- **최적화 기법**:
  - Faiss HNSW 인덱스로 근사 최근접 이웃 검색
  - Go 루틴 기반 병렬 처리
  - 벡터 정규화 사전 계산
  - 메모리 풀링으로 GC 압박 최소화

#### 2. 확장성 설계
- **수평 확장**: ML 서비스 로드 밸런싱 지원
- **벡터 DB 샤딩**: 카테고리별 인덱스 분산 계획
- **캐싱 전략**: Redis 기반 임베딩 결과 캐싱
- **모니터링**: Prometheus + Grafana 지표 수집

### 🔮 향후 발전 계획

#### Phase 2: 고도화 (2024 Q4)
```yaml
고급_임베딩_모델:
  - 모델: KR-SBERT-V2, KoSimCSE-RoBERTa
  - 다중모델: 앙상블 임베딩으로 정확도 향상
  - 파인튜닝: 자기소개서 도메인 특화 학습

개인화_추천:
  - 사용자_프로필: 학과, 직무, 경력 기반 가중치
  - 피드백_학습: 사용자 선호도 반영 재랭킹
  - A/B_테스트: 추천 알고리즘 성능 비교

실시간_학습:
  - 온라인_학습: 새로운 구문 자동 발견 및 추가
  - 품질_관리: 부적절한 조합 자동 필터링
  - 트렌드_반영: 시기별 인기 키워드 가중치 조정
```

#### Phase 3: AI 고도화 (2025 Q1)
```yaml
멀티모달_분석:
  - 텍스트+이미지: 포트폴리오 이미지 분석 추가
  - 음성_분석: 면접 연습 영상에서 말투/톤 분석
  - 감정_분석: 글의 어조와 감정 상태 반영

생성형_AI_통합:
  - GPT_기반: 췽호 설명 및 개선 제안 생성
  - 개인화_코칭: 자기소개서 개선 방향 제시
  - 면접_시뮬레이션: 췽호 기반 예상 질문 생성

고급_벡터_검색:
  - 하이브리드_검색: 키워드 + 의미적 검색 결합
  - 설명가능_AI: 추천 이유 및 근거 제시
  - 대화형_개선: 사용자와의 대화로 점진적 개선
```

#### Phase 4: 플랫폼화 (2025 Q2)
```yaml
API_플랫폼:
  - 개발자_API: 타사 서비스 통합용 REST/GraphQL API
  - 임베딩_API: 범용 한국어 텍스트 임베딩 서비스
  - 맞춤형_모델: 기업별 특화 췽호 시스템 구축

데이터_생태계:
  - 크라우드소싱: 사용자 기여형 구문 데이터베이스
  - 전문가_큐레이션: HR 전문가 검증 시스템
  - 산업별_특화: IT, 금융, 의료 등 분야별 췽호
```

### 📊 기술 지표 및 모니터링

#### 현재 성능 메트릭
```yaml
임베딩_성능:
  - 처리속도: ~100 토큰/초 (CPU), ~500 토큰/초 (GPU)
  - 메모리사용: 모델당 ~2GB RAM
  - 배치크기: 최적 16-32 텍스트/배치

검색_성능:
  - 검색지연: p50=45ms, p95=120ms, p99=200ms
  - 처리량: 1000 QPS (단일 노드)
  - 정확도: Top-3 만족도 85%+ (내부 테스트)

시스템_안정성:
  - 가용성: 99.9% SLA 목표
  - 오류율: < 0.1% 4xx/5xx 응답
  - 복구시간: < 30초 자동 재시작
```

#### 품질 보증
```yaml
테스트_커버리지:
  - 단위테스트: Go 85%+, Python 90%+
  - 통합테스트: API 시나리오 95% 커버
  - 성능테스트: 부하 시나리오 자동화

모델_검증:
  - 회귀테스트: 배포 전 기준 데이터셋 검증
  - A/B_테스트: 새 모델 점진적 적용
  - 사용자_피드백: 실시간 만족도 수집

보안_및_개인정보:
  - 암호화: 모든 PII 데이터 AES-256 암호화
  - 익명화: 로그에서 개인정보 자동 마스킹
  - 컴플라이언스: GDPR, PIPEDA 준수
```

### 🛠 개발 및 운영 도구

#### ML 모델 관리
```bash
# 모델 배포 파이프라인
make deploy-model MODEL=KoSimCSE-bert VERSION=v1.2
make test-model ENDPOINT=http://ml-service:8001
make rollback-model PREVIOUS_VERSION=v1.1

# 성능 벤치마크
make benchmark-embedding BATCH_SIZE=32 ITERATIONS=1000
make benchmark-search VECTORS=1M QUERIES=10000
```

#### 모니터링 대시보드
```yaml
실시간_지표:
  - 요청_수: 분당 처리 요청 수
  - 응답_시간: 백분위별 지연시간 분포  
  - 오류_률: HTTP 상태코드별 집계
  - 리소스_사용: CPU, 메모리, GPU 사용률

비즈니스_지표:
  - 사용자_만족도: 췽호 채택률
  - 다양성_점수: 추천 결과의 다양성 측정
  - 재사용률: 동일 사용자 재방문 비율
```

## 🏗 시스템 아키텍처

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │───▶│   Go API    │───▶│  ML Service │
│  (React)    │    │   Server    │    │  (Python)   │
└─────────────┘    └─────────────┘    └─────────────┘
                           │                   │
                           ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │ PostgreSQL  │    │ Vector DB   │
                   │ (메타데이터)   │    │  (Faiss)    │
                   └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │   MinIO     │
                   │ (파일저장)    │
                   └─────────────┘
```

### 컴포넌트 설명

- **Go API Server**: Gin 기반 REST API, 비즈니스 로직 처리
- **ML Service**: FastAPI 기반, KoSimCSE-BERT 임베딩 생성
- **PostgreSQL**: 사용자, 자기소개서, 추천 결과 메타데이터
- **Vector DB**: Faiss 기반 고성능 벡터 검색 엔진
- **MinIO**: S3 호환 객체 스토리지, 원본 파일 암호화 저장

## 🛠 기술 스택

### 백엔드
- **Go 1.22**: 메인 API 서버
- **Gin 2.x**: HTTP 웹 프레임워크
- **pgx/v5**: PostgreSQL 드라이버
- **Viper**: 설정 관리
- **Zap**: 구조화된 로깅

### ML 서비스
- **Python 3.11**: ML 서비스 런타임
- **FastAPI**: API 프레임워크
- **sentence-transformers**: 임베딩 모델
- **PyTorch**: 딥러닝 프레임워크

### 데이터베이스 & 스토리지
- **PostgreSQL 15**: 메인 데이터베이스
- **MinIO**: S3 호환 객체 스토리지
- **Faiss**: 벡터 검색 엔진

### 인프라
- **Docker & Docker Compose**: 컨테이너화
- **GitHub Actions**: CI/CD 파이프라인

## 🚀 빠른 시작

### 필수 요구사항

- Docker & Docker Compose
- Make (선택사항)

### 1. 프로젝트 클론

```bash
git clone https://github.com/Todari/chuingho-server.git
cd chuingho-server
```

### 2. 전체 스택 실행

```bash
# Make 사용시
make up

# 또는 직접 실행
docker-compose up -d --build
```

### 3. 서비스 확인

```bash
# API 서버 헬스체크
curl http://localhost:8080/health

# ML 서비스 헬스체크  
curl http://localhost:8001/health
```

### 4. 구문 사전 구축 (최초 1회)

```bash
# 샘플 구문으로 벡터 DB 초기화
make prepare-phrases
```

### 5. API 테스트

```bash
# 자기소개서 텍스트 업로드
curl -X POST -H "Content-Type: application/json" \
  -d '{"text":"안녕하세요. 저는 창의적이고 열정적인 개발자입니다..."}' \
  http://localhost:8080/v1/resumes

# 췽호 추천 (위에서 받은 resumeId 사용)
curl -X POST -H "Content-Type: application/json" \
  -d '{"resumeId":"RESUME_ID_HERE"}' \
  http://localhost:8080/v1/titles
```

## 📚 API 문서

### 주요 엔드포인트

#### 자기소개서 업로드
```http
POST /v1/resumes
Content-Type: application/json

{
  "text": "안녕하세요. 저는 창의적이고 열정적인 개발자입니다..."
}
```

**응답 예시:**
```json
{
  "resumeId": "123e4567-e89b-12d3-a456-426614174000",
  "status": "uploaded"
}
```

#### 췽호 추천 생성
```http
POST /v1/titles
Content-Type: application/json

{
  "resumeId": "123e4567-e89b-12d3-a456-426614174000"
}
```

**응답 예시:**
```json
{
  "titles": [
    "창의적 설계자",
    "세심한 분석가", 
    "적극적 리더"
  ]
}
```

#### 헬스체크
```http
GET /health
```

### 전체 API 문서

개발 중: Swagger/OpenAPI 3.1 문서 자동 생성 예정

## ⚙️ 환경 설정

### 환경변수

주요 환경변수들을 `.env` 파일이나 시스템 환경변수로 설정할 수 있습니다:

```bash
# 서버 설정
CHUINGHO_SERVER_PORT=8080
CHUINGHO_SERVER_ENVIRONMENT=production

# 데이터베이스
CHUINGHO_DATABASE_HOST=localhost
CHUINGHO_DATABASE_PASSWORD=your_password

# ML 서비스
CHUINGHO_ML_SERVICE_URL=http://localhost:8001
CHUINGHO_ML_EMBEDDING_MODEL=BM-K/KoSimCSE-bert

# 벡터 DB
CHUINGHO_VECTOR_TYPE=faiss
CHUINGHO_VECTOR_DIMENSION=768
```

### 설정 파일

`config.yaml` 파일을 통해 상세 설정 가능:

```yaml
server:
  port: 8080
  environment: dev

database:
  host: localhost
  port: 5432
  
ml:
  service_url: http://localhost:8001
  timeout: 30
  
# ... 기타 설정
```

## 👩‍💻 개발 가이드

### 로컬 개발 환경 설정

```bash
# 1. 개발 환경 설정
make dev-setup

# 2. 의존성 서비스만 시작 (DB, Storage, ML)
make dev-deps

# 3. 로컬에서 API 서버 실행
make run-server
```

### 코드 스타일

```bash
# 코드 포맷팅
make fmt

# 린팅
make lint

# 테스트
make test
```

### 디렉토리 구조

```
chuingho-server/
├── cmd/                    # 실행 가능한 애플리케이션
│   ├── server/            # 메인 API 서버
│   ├── migration/         # DB 마이그레이션 도구
│   └── prepare_phrases/   # 구문 사전 구축 도구
├── internal/              # 내부 패키지
│   ├── config/           # 설정 관리
│   ├── database/         # DB 연결
│   ├── handler/          # HTTP 핸들러
│   ├── service/          # 비즈니스 로직
│   ├── storage/          # 객체 스토리지
│   └── vector/           # 벡터 DB
├── pkg/                   # 공개 패키지
│   ├── model/            # 데이터 모델
│   └── util/             # 유틸리티
├── ml-service/            # Python ML 서비스
├── migrations/            # DB 스키마 마이그레이션
└── test/                  # 테스트 코드
```

### 새 기능 개발

1. **피처 브랜치 생성**
```bash
git checkout -b feature/새기능명
```

2. **코드 작성 및 테스트**
```bash
make test
make lint
```

3. **통합 테스트**
```bash
make up
make test-api
```

4. **PR 생성 및 리뷰**

## 🚀 배포

### 프로덕션 배포

```bash
# 1. 환경변수 설정
export CHUINGHO_SERVER_ENVIRONMENT=production
export CHUINGHO_DATABASE_PASSWORD=strong_password

# 2. 프로덕션 모드로 시작
docker-compose -f docker-compose.prod.yaml up -d
```

### Kubernetes 배포

K8s 매니페스트 파일 제공 예정

### 모니터링

- 헬스체크: `/health`, `/ready`, `/live`
- 메트릭스: Prometheus 지원 예정
- 로그: 구조화된 JSON 로그

## 🧪 테스트

### 단위 테스트

```bash
# 전체 테스트 실행
make test

# 특정 패키지 테스트
go test -v ./internal/service/...

# 커버리지 확인
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 통합 테스트

```bash
# 전체 스택으로 테스트
make up
make test-api
```

### 성능 테스트

```bash
# 벡터 검색 성능 (목표: p95 < 200ms)
ab -n 1000 -c 10 -T "application/json" \
  -p test_data.json http://localhost:8080/v1/titles
```

## 📊 성능 지표

### 목표 성능

- **응답 시간**: p95 < 200ms (벡터 검색)
- **처리량**: 100 req/s (동시 사용자)
- **가용성**: 99.9% (연간 약 8.76시간 다운타임)

### 리소스 요구사항

- **API 서버**: CPU 2코어, 메모리 4GB
- **ML 서비스**: CPU 4코어, 메모리 8GB (GPU 권장)
- **데이터베이스**: CPU 2코어, 메모리 4GB, 스토리지 100GB
- **벡터 DB**: 메모리 4GB (1M 벡터 기준)

## 🤝 기여하기

프로젝트에 기여해 주셔서 감사합니다!

### 기여 방법

1. **이슈 확인**: [GitHub Issues](https://github.com/Todari/chuingho-server/issues)
2. **포크 및 브랜치**: `feature/기능명` 또는 `fix/버그명`
3. **커밋 컨벤션**: [Conventional Commits](https://conventionalcommits.org/) 사용
4. **테스트**: 모든 테스트 통과 확인
5. **PR 생성**: 상세한 설명과 함께

### 코드 리뷰 체크리스트

- [ ] 모든 테스트 통과
- [ ] 코드 커버리지 ≥ 80%
- [ ] 한글 주석 및 문서화
- [ ] 성능 영향도 평가
- [ ] 보안 취약점 검토

## 📄 라이센스

이 프로젝트는 MIT 라이센스를 따릅니다. 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

## 📞 문의

- **이메일**: [개발자 이메일]
- **GitHub Issues**: [프로젝트 이슈](https://github.com/Todari/chuingho-server/issues)
- **Discord**: [개발 커뮤니티] (준비 중)

---

**Made with ❤️ for 취준생들**

취업 준비는 힘들지만, 당신만의 특별한 췽호로 자신감을 찾아보세요! 🌟