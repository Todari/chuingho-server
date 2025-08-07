# Go 애플리케이션 Dockerfile

# 빌드 스테이지
FROM golang:1.22-alpine AS builder

# 빌드 도구 설치
RUN apk add --no-cache git ca-certificates tzdata

# 작업 디렉토리 설정
WORKDIR /build

# Go 모듈 파일 복사 및 의존성 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 복사
COPY . .

# 바이너리 빌드
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o server ./cmd/server

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o migration ./cmd/migration

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o prepare_phrases ./cmd/prepare_phrases

# 런타임 스테이지
FROM alpine:latest

# 시스템 패키지 업데이트 및 CA 인증서 설치
RUN apk --no-cache add ca-certificates tzdata && \
    update-ca-certificates

# 비특권 사용자 생성
RUN adduser -D -g '' appuser

# 작업 디렉토리 생성
WORKDIR /app

# 빌드된 바이너리 복사
COPY --from=builder /build/server /app/
COPY --from=builder /build/migration /app/
COPY --from=builder /build/prepare_phrases /app/

# 설정 파일 복사 (선택적)
COPY --from=builder /build/config.yaml /app/config.yaml* 
COPY --from=builder /build/phrases_corpus.txt /app/
COPY --from=builder /build/sample_resume.txt /app/

# 마이그레이션 디렉토리 복사
COPY --from=builder /build/migrations /app/migrations

# 사용자 권한 설정
RUN chown -R appuser:appuser /app
USER appuser

# 환경변수 설정
ENV GIN_MODE=release
ENV CHUINGHO_SERVER_ENVIRONMENT=production

# 포트 노출
EXPOSE 8080

# 헬스체크
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 엔트리포인트 스크립트
COPY --from=builder /build/docker-entrypoint.sh /app/
RUN chmod +x /app/docker-entrypoint.sh

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["server"]