# 췽호 프로젝트 Makefile

.PHONY: help build test clean up down restart logs migration prepare-phrases

# 기본 타겟
help:
	@echo "췽호 프로젝트 빌드 도구"
	@echo ""
	@echo "사용 가능한 명령어:"
	@echo "  build           - Go 바이너리 빌드"
	@echo "  test            - 테스트 실행"
	@echo "  clean           - 빌드 아티팩트 정리"
	@echo "  up              - Docker Compose로 전체 스택 시작"
	@echo "  down            - Docker Compose 스택 중지"
	@echo "  restart         - 스택 재시작"
	@echo "  logs            - 로그 출력"
	@echo "  migration       - 데이터베이스 마이그레이션 실행"
	@echo "  prepare-phrases - 구문 후보 사전 구축"
	@echo "  test-api        - API 테스트 실행"

# 바이너리 빌드
build:
	@echo "Go 바이너리 빌드 중..."
	go build -o bin/server ./cmd/server
	go build -o bin/migration ./cmd/migration
	go build -o bin/prepare_phrases ./cmd/prepare_phrases
	@echo "빌드 완료!"

# 테스트 실행
test:
	@echo "테스트 실행 중..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "테스트 완료! 커버리지 리포트: coverage.html"

# 정리
clean:
	@echo "빌드 아티팩트 정리 중..."
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker system prune -f
	@echo "정리 완료!"

# Docker Compose 전체 스택 시작
up:
	@echo "전체 스택 시작 중..."
	docker-compose up -d --build
	@echo "스택 시작 완료!"
	@echo "API 서버: http://localhost:8080"
	@echo "MinIO 콘솔: http://localhost:9001 (minioadmin/minioadmin123)"

# Docker Compose 스택 중지
down:
	@echo "스택 중지 중..."
	docker-compose down
	@echo "스택 중지 완료!"

# 스택 재시작
restart: down up

# 로그 출력
logs:
	docker-compose logs -f

# 특정 서비스 로그
logs-api:
	docker-compose logs -f api-server

logs-ml:
	docker-compose logs -f ml-service

# 데이터베이스 마이그레이션
migration:
	@echo "데이터베이스 마이그레이션 실행 중..."
	docker-compose run --rm migration
	@echo "마이그레이션 완료!"

# 구문 후보 사전 구축
prepare-phrases:
	@echo "구문 후보 사전 구축 중..."
	docker-compose --profile tools run --rm phrase-builder
	@echo "구문 사전 구축 완료!"

# API 테스트
test-api:
	@echo "API 테스트 실행 중..."
	@echo "자기소개서 업로드 테스트..."
	curl -F "file=@sample_resume.txt" http://localhost:8080/v1/resumes
	@echo ""
	@echo "헬스체크 테스트..."
	curl -s http://localhost:8080/health | jq .
	@echo ""

# 개발 환경 설정
dev-setup:
	@echo "개발 환경 설정 중..."
	go mod tidy
	go mod download
	@echo "개발 환경 설정 완료!"

# 로컬 실행 (의존성 서비스만)
dev-deps:
	@echo "개발용 의존성 서비스 시작 중..."
	docker-compose up -d postgres minio ml-service
	@echo "의존성 서비스 시작 완료!"

# 바이너리 실행 (로컬)
run-server: build
	@echo "로컬에서 서버 실행 중..."
	./bin/server

run-migration: build
	@echo "로컬에서 마이그레이션 실행 중..."
	./bin/migration

run-prepare-phrases: build
	@echo "로컬에서 구문 준비 실행 중..."
	./bin/prepare_phrases

# Docker 이미지 빌드만
docker-build:
	@echo "Docker 이미지 빌드 중..."
	docker build -t chuingho-server .
	docker build -t chuingho-ml-service ./ml-service
	@echo "Docker 이미지 빌드 완료!"

# 포맷팅 및 린팅
fmt:
	@echo "코드 포맷팅 중..."
	go fmt ./...
	@echo "포맷팅 완료!"

lint:
	@echo "린팅 중..."
	golangci-lint run
	@echo "린팅 완료!"

# 전체 검사
check: fmt lint test
	@echo "전체 검사 완료!"