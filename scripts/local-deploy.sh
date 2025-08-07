#!/bin/bash

# 🚀 로컬 배포 시뮬레이션 스크립트
set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 로그 함수
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 배포 설정
APP_NAME="chuingho-server"
CURRENT_VERSION=$(git rev-parse --short HEAD)
DEPLOY_DIR="/tmp/chuingho-deploy"
OLD_PORT=${OLD_PORT:-8080}
NEW_PORT=${NEW_PORT:-8081}

log_info "🚀 로컬 배포 시뮬레이션 시작"
log_info "📦 버전: $CURRENT_VERSION"
log_info "🔄 롤링 배포: $OLD_PORT -> $NEW_PORT"

# 1. 사전 검사
log_info "1️⃣ 사전 검사 수행 중..."

# Go 바이너리 빌드 확인
if [ ! -f "bin/test-server" ]; then
    log_warning "바이너리가 없습니다. 빌드를 시작합니다..."
    make build-test
fi

# 현재 서비스 헬스체크
log_info "현재 서비스 상태 확인 중..."
if curl -s -f http://localhost:$OLD_PORT/health > /dev/null; then
    log_success "현재 서비스가 정상 동작 중입니다"
    CURRENT_STATUS=$(curl -s http://localhost:$OLD_PORT/health | jq -r '.status')
    log_info "현재 상태: $CURRENT_STATUS"
else
    log_warning "현재 서비스가 실행되지 않고 있습니다"
fi

# 2. 새 버전 배포 준비
log_info "2️⃣ 새 버전 배포 준비 중..."

# 배포 디렉토리 생성
mkdir -p $DEPLOY_DIR
cp bin/test-server $DEPLOY_DIR/

# 환경별 설정 파일 준비
cat > $DEPLOY_DIR/config.yaml << EOF
server:
  port: $NEW_PORT
  environment: staging
database:
  host: localhost
  port: 5432
  username: chuingho
  password: chuingho_password
  dbname: chuingho_staging
logging:
  level: debug
EOF

log_success "새 버전 준비 완료"

# 3. 블루-그린 배포 시뮬레이션
log_info "3️⃣ 블루-그린 배포 시뮬레이션..."

# 새 포트에서 서비스 시작 (시뮬레이션)
log_info "새 포트($NEW_PORT)에서 서비스 시작 중..."

# 포트 사용 중인지 확인
if lsof -i :$NEW_PORT > /dev/null 2>&1; then
    log_warning "포트 $NEW_PORT가 이미 사용 중입니다"
    EXISTING_PID=$(lsof -t -i :$NEW_PORT)
    log_info "기존 프로세스 종료 중: PID $EXISTING_PID"
    kill -9 $EXISTING_PID 2>/dev/null || true
    sleep 2
fi

# 새 서비스 시작 (백그라운드)
log_info "새 서비스 시작 중..."
cd $DEPLOY_DIR
PORT=$NEW_PORT ./test-server > /tmp/new-service.log 2>&1 &
NEW_PID=$!
cd - > /dev/null

log_info "새 서비스 PID: $NEW_PID"

# 4. 헬스체크 대기
log_info "4️⃣ 새 서비스 헬스체크 대기 중..."

MAX_RETRIES=30
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s -f http://localhost:$NEW_PORT/health > /dev/null; then
        log_success "새 서비스가 정상적으로 시작되었습니다!"
        NEW_STATUS=$(curl -s http://localhost:$NEW_PORT/health | jq -r '.status')
        log_info "새 서비스 상태: $NEW_STATUS"
        break
    fi
    
    log_info "헬스체크 재시도 중... ($((RETRY_COUNT + 1))/$MAX_RETRIES)"
    sleep 2
    RETRY_COUNT=$((RETRY_COUNT + 1))
done

if [ $RETRY_COUNT -eq $MAX_RETRIES ]; then
    log_error "새 서비스 시작에 실패했습니다"
    log_info "로그 확인: tail -f /tmp/new-service.log"
    kill -9 $NEW_PID 2>/dev/null || true
    exit 1
fi

# 5. 스모크 테스트
log_info "5️⃣ 스모크 테스트 수행 중..."

# API 테스트
RESUME_TEXT="안녕하세요. 저는 창의적이고 열정적인 개발자입니다. 팀워크를 중시하며 지속적인 학습과 성장을 추구합니다."

log_info "자기소개서 업로드 테스트..."
UPLOAD_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"text\":\"$RESUME_TEXT\"}" \
    http://localhost:$NEW_PORT/v1/resumes)

if echo "$UPLOAD_RESPONSE" | jq -e '.resumeId' > /dev/null; then
    RESUME_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.resumeId')
    log_success "자기소개서 업로드 성공: $RESUME_ID"
    
    log_info "췽호 생성 테스트..."
    TITLE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
        -d "{\"resumeId\":\"$RESUME_ID\"}" \
        http://localhost:$NEW_PORT/v1/titles)
    
    if echo "$TITLE_RESPONSE" | jq -e '.titles' > /dev/null; then
        TITLES=$(echo "$TITLE_RESPONSE" | jq -r '.titles[]' | tr '\n' ', ')
        log_success "췽호 생성 성공: $TITLES"
    else
        log_error "췽호 생성 실패"
        echo "$TITLE_RESPONSE"
    fi
else
    log_error "자기소개서 업로드 실패"
    echo "$UPLOAD_RESPONSE"
fi

# 6. 트래픽 전환 시뮬레이션
log_info "6️⃣ 트래픽 전환 시뮬레이션..."

log_info "로드 밸런서 설정 업데이트 중... (시뮬레이션)"
log_info "기존 서비스($OLD_PORT) -> 새 서비스($NEW_PORT)로 트래픽 전환"

# 실제로는 nginx, haproxy 등의 설정을 업데이트
sleep 2
log_success "트래픽 전환 완료"

# 7. 모니터링 및 검증
log_info "7️⃣ 배포 후 모니터링..."

log_info "에러율 모니터링 중..."
for i in {1..5}; do
    HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:$NEW_PORT/health)
    if [ "$HTTP_STATUS" = "200" ]; then
        log_success "헬스체크 $i/5: OK"
    else
        log_error "헬스체크 $i/5: FAIL (HTTP $HTTP_STATUS)"
    fi
    sleep 1
done

# 8. 정리
log_info "8️⃣ 배포 완료 및 정리..."

log_info "새 서비스 정보:"
log_info "  - PID: $NEW_PID"
log_info "  - Port: $NEW_PORT"
log_info "  - Version: $CURRENT_VERSION"
log_info "  - Log: /tmp/new-service.log"

log_success "🎉 로컬 배포 시뮬레이션 완료!"

# 서비스 상태 요약
echo ""
echo "=== 서비스 상태 요약 ==="
echo "기존 서비스 (포트 $OLD_PORT):"
if curl -s -f http://localhost:$OLD_PORT/health > /dev/null; then
    echo "  상태: ✅ 정상"
else
    echo "  상태: ❌ 중단"
fi

echo "새 서비스 (포트 $NEW_PORT):"
if curl -s -f http://localhost:$NEW_PORT/health > /dev/null; then
    echo "  상태: ✅ 정상"
else
    echo "  상태: ❌ 중단"
fi

echo ""
echo "=== 다음 단계 ==="
echo "1. 새 서비스 로그 확인: tail -f /tmp/new-service.log"
echo "2. 새 서비스 중지: kill $NEW_PID"
echo "3. 기존 서비스로 롤백: curl http://localhost:$OLD_PORT/health"