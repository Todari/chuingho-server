# syntax=docker/dockerfile:1.7
# Go 애플리케이션 Dockerfile

# 빌드 스테이지
FROM golang:1.23-alpine AS builder

# 빌드 도구 설치
RUN apk add --no-cache git ca-certificates tzdata build-base

# 작업 디렉토리 설정
WORKDIR /build

# Go 모듈 파일 복사 및 의존성 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 최소 복사 (불필요한 대용량 제외)
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY migrations/ ./migrations/
COPY data/ ./data/
COPY docker-entrypoint.sh ./
COPY config.yaml ./
## 선택 샘플/데이터 파일은 빌드에 필수 아님 (리포에 없을 수 있으므로 복사하지 않음)

# 멀티아키텍처 빌드를 위한 타겟 변수 수신 (buildx가 주입)
ARG TARGETOS
ARG TARGETARCH

# 바이너리 빌드 (디렉터리 진입 후 빌드하여 경로 문제 방지)
WORKDIR /build/cmd/server
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/server .

WORKDIR /build/cmd/migration
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/migration .

WORKDIR /build/cmd/prepare_phrases
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/prepare_phrases .

# 다시 루트로 복귀
WORKDIR /build

# 런타임 스테이지
# 런타임 스테이지는 distroless로도 가능하지만, 여기서는 Alpine 유지
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
COPY --from=builder /build/config.yaml /app/config.yaml

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

# 엔트리포인트 스크립트 (권한/소유자 지정하여 복사)
USER root
COPY --from=builder --chown=appuser:appuser --chmod=0755 /build/docker-entrypoint.sh /app/docker-entrypoint.sh
USER appuser

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["server"]