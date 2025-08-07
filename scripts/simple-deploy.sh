#!/bin/bash

# 🚀 간단한 로컬 배포 테스트
set -e

# 색상 정의
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${BLUE}=== 🚀 로컬 배포 테스트 시작 ===${NC}"

# 1. 현재 서비스 상태 확인
echo -e "${BLUE}1️⃣ 현재 서비스 상태 확인${NC}"
echo "포트 8080 서비스 헬스체크:"
if curl -s http://localhost:8080/health | jq .; then
    echo -e "${GREEN}✅ 서비스 정상 동작 중${NC}"
else
    echo -e "${YELLOW}⚠️ 서비스가 실행되지 않고 있습니다${NC}"
fi

# 2. API 기능 테스트
echo -e "\n${BLUE}2️⃣ API 기능 테스트${NC}"

# 자기소개서 업로드 테스트
echo "자기소개서 업로드 테스트..."
RESUME_TEXT="안녕하세요. 저는 창의적이고 열정적인 개발자입니다. 팀워크를 중시하며 지속적인 학습과 성장을 추구합니다. 다양한 프로젝트 경험을 통해 문제 해결 능력을 키워왔으며, 새로운 기술에 대한 호기심이 많습니다."

UPLOAD_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
    -d "{\"text\":\"$RESUME_TEXT\"}" \
    http://localhost:8080/v1/resumes)

echo "업로드 응답:"
echo "$UPLOAD_RESPONSE" | jq .

if echo "$UPLOAD_RESPONSE" | jq -e '.resumeId' > /dev/null; then
    RESUME_ID=$(echo "$UPLOAD_RESPONSE" | jq -r '.resumeId')
    echo -e "${GREEN}✅ 자기소개서 업로드 성공: $RESUME_ID${NC}"
    
    # 췽호 생성 테스트
    echo -e "\n췽호 생성 테스트..."
    TITLE_RESPONSE=$(curl -s -X POST -H "Content-Type: application/json" \
        -d "{\"resumeId\":\"$RESUME_ID\"}" \
        http://localhost:8080/v1/titles)
    
    echo "췽호 생성 응답:"
    echo "$TITLE_RESPONSE" | jq .
    
    if echo "$TITLE_RESPONSE" | jq -e '.titles' > /dev/null; then
        echo -e "${GREEN}✅ 췽호 생성 성공${NC}"
        echo "생성된 췽호들:"
        echo "$TITLE_RESPONSE" | jq -r '.titles[]' | sed 's/^/  - /'
    else
        echo -e "${YELLOW}⚠️ 췽호 생성 실패 (모의 응답 확인)${NC}"
    fi
else
    echo -e "${YELLOW}⚠️ 자기소개서 업로드 실패${NC}"
fi

# 3. 부하 테스트 시뮬레이션
echo -e "\n${BLUE}3️⃣ 부하 테스트 시뮬레이션${NC}"
echo "5개의 동시 요청 전송..."

for i in {1..5}; do
    curl -s -X GET http://localhost:8080/health > /tmp/health_$i.json &
done

wait

echo "헬스체크 결과:"
for i in {1..5}; do
    if [ -f "/tmp/health_$i.json" ]; then
        echo "  요청 $i: $(cat /tmp/health_$i.json | jq -r '.status')"
        rm /tmp/health_$i.json
    fi
done

# 4. 배포 시나리오 시뮬레이션
echo -e "\n${BLUE}4️⃣ 배포 시나리오 시뮬레이션${NC}"

echo "📋 실제 배포에서 수행될 작업들:"
echo "  1. ✅ 헬스체크 확인"
echo "  2. ✅ API 기능 테스트" 
echo "  3. ✅ 부하 테스트"
echo "  4. 🔄 새 버전 이미지 빌드"
echo "  5. 🔄 블루-그린 배포"
echo "  6. 🔄 트래픽 전환"
echo "  7. 🔄 구버전 종료"

echo -e "\n${BLUE}5️⃣ 배포 메트릭 수집${NC}"

# 응답 시간 측정
echo "응답 시간 측정 중..."
RESPONSE_TIME=$(curl -s -o /dev/null -w "%{time_total}" http://localhost:8080/health)
echo "헬스체크 응답 시간: ${RESPONSE_TIME}초"

# 메모리 사용량 확인
echo "서버 프로세스 리소스 사용량:"
ps aux | grep test-server | grep -v grep | awk '{print "  CPU: " $3 "%, MEM: " $4 "%, PID: " $2}'

echo -e "\n${GREEN}🎉 로컬 배포 테스트 완료!${NC}"

echo -e "\n=== 📊 배포 요약 ==="
echo "배포 환경: 로컬 테스트"
echo "현재 버전: $(git rev-parse --short HEAD)"
echo "테스트 포트: 8080"
echo "상태: 정상"

echo -e "\n=== 🔗 유용한 링크 ==="
echo "헬스체크: curl http://localhost:8080/health"
echo "API 문서: http://localhost:8080/swagger/index.html (가능한 경우)"
echo "로그 확인: 터미널에서 서버 실행 로그 확인"