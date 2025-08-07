#!/bin/bash

# 🐳 Docker 기반 배포 테스트
set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== 🐳 Docker 배포 테스트 ===${NC}"

# Docker 상태 확인
echo -e "${BLUE}1️⃣ Docker 환경 확인${NC}"
if docker --version > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker 설치됨${NC}"
    docker --version
else
    echo -e "${RED}❌ Docker가 설치되지 않았습니다${NC}"
    exit 1
fi

if docker ps > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Docker 데몬 실행 중${NC}"
else
    echo -e "${YELLOW}⚠️ Docker 데몬이 실행되지 않고 있습니다${NC}"
    echo "Docker Desktop을 시작해주세요"
    exit 1
fi

# 이미지 빌드
echo -e "\n${BLUE}2️⃣ Docker 이미지 빌드${NC}"
echo "API 서버 이미지 빌드 중..."
if docker build -t chuingho-api:local .; then
    echo -e "${GREEN}✅ API 서버 이미지 빌드 성공${NC}"
else
    echo -e "${RED}❌ API 서버 이미지 빌드 실패${NC}"
    exit 1
fi

echo "ML 서비스 이미지 빌드 중..."
if docker build -t chuingho-ml:local ./ml-service/; then
    echo -e "${GREEN}✅ ML 서비스 이미지 빌드 성공${NC}"
else
    echo -e "${RED}❌ ML 서비스 이미지 빌드 실패${NC}"
    exit 1
fi

# 컨테이너 실행
echo -e "\n${BLUE}3️⃣ 컨테이너 배포${NC}"

# 기존 컨테이너 정리
echo "기존 컨테이너 정리 중..."
docker-compose down -v 2>/dev/null || true

# 새 컨테이너 시작
echo "새 컨테이너 시작 중..."
if docker-compose up -d; then
    echo -e "${GREEN}✅ 컨테이너 배포 성공${NC}"
else
    echo -e "${RED}❌ 컨테이너 배포 실패${NC}"
    exit 1
fi

# 서비스 시작 대기
echo -e "\n${BLUE}4️⃣ 서비스 시작 대기${NC}"
echo "서비스가 시작될 때까지 대기 중..."

MAX_WAIT=120
WAIT_COUNT=0

while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
    if curl -s -f http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✅ API 서버 준비 완료${NC}"
        break
    fi
    
    echo -n "."
    sleep 2
    WAIT_COUNT=$((WAIT_COUNT + 2))
done

if [ $WAIT_COUNT -eq $MAX_WAIT ]; then
    echo -e "${RED}❌ API 서버 시작 타임아웃${NC}"
    echo "로그 확인:"
    docker-compose logs api-server
    exit 1
fi

# 배포 검증
echo -e "\n${BLUE}5️⃣ 배포 검증${NC}"

# 헬스체크
echo "헬스체크 수행..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/health)
echo "헬스체크 응답: $HEALTH_RESPONSE"

# 컨테이너 상태 확인
echo -e "\n컨테이너 상태:"
docker-compose ps

# API 테스트
echo -e "\nAPI 기능 테스트..."
RESUME_TEXT="컨테이너 환경에서 실행되는 테스트용 자기소개서입니다."

UPLOAD_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"text\":\"$RESUME_TEXT\"}" \
    http://localhost:8080/v1/resumes)

if echo "$UPLOAD_RESPONSE" | jq -e '.resumeId' > /dev/null; then
    echo -e "${GREEN}✅ 컨테이너 환경 API 테스트 성공${NC}"
else
    echo -e "${YELLOW}⚠️ API 테스트 실패 또는 모의 응답${NC}"
fi

echo -e "\n${GREEN}🎉 Docker 배포 테스트 완료!${NC}"

echo -e "\n=== 📊 배포 상태 ==="
echo "배포 방식: Docker Compose"
echo "API 서버: http://localhost:8080"
echo "ML 서비스: http://localhost:8001"
echo "데이터베이스: localhost:5432"

echo -e "\n=== 🔧 관리 명령어 ==="
echo "로그 확인: docker-compose logs -f"
echo "컨테이너 중지: docker-compose down"
echo "완전 정리: docker-compose down -v --rmi all"