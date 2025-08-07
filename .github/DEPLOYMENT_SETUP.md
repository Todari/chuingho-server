# 🚀 배포 환경 설정 가이드

## 📋 필수 GitHub Secrets 설정

### 개발 환경 (Development)
```bash
# 데이터베이스 설정
DEV_DB_HOST=dev-postgres.chuingho.com
DEV_DB_USERNAME=chuingho_dev
DEV_DB_PASSWORD=your_dev_password

# 외부 서비스 (선택사항)
DEV_STORAGE_ACCESS_KEY=your_dev_minio_key
DEV_STORAGE_SECRET_KEY=your_dev_minio_secret
```

### 프로덕션 환경 (Production)
```bash
# 데이터베이스 설정
PROD_DB_HOST=prod-postgres.chuingho.com
PROD_DB_USERNAME=chuingho_prod
PROD_DB_PASSWORD=your_prod_password

# Kubernetes 클러스터 접근
KUBE_CONFIG_DATA=base64_encoded_kubeconfig

# 알림 설정
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
```

### 공통 설정
```bash
# GitHub Container Registry 접근 (자동 설정됨)
GITHUB_TOKEN=자동으로_설정됨

# 코드 품질 분석 (선택사항)
SONAR_TOKEN=your_sonarcloud_token
CODECOV_TOKEN=your_codecov_token
```

## 🛠 배포 방식별 설정

### 1. Docker Compose (개발 환경)
```yaml
# 현재 구성: docker-compose.dev.yml 자동 생성
# 필요한 작업:
# 1. 개발 서버에 Docker, Docker Compose 설치
# 2. 배포 사용자 계정 및 SSH 키 설정
# 3. 환경 변수 파일 준비
```

### 2. Kubernetes (프로덕션 환경)
```yaml
# 현재 구성: k8s-deployment.yml 자동 생성
# 필요한 작업:
# 1. Kubernetes 클러스터 준비 (EKS, GKE, AKS 등)
# 2. kubectl 설치 및 클러스터 연결
# 3. Secret 리소스 생성
# 4. 네임스페이스 및 RBAC 설정
```

## 🔐 보안 설정

### Secret 생성 명령어 (Kubernetes)
```bash
# 데이터베이스 시크릿
kubectl create secret generic db-secret \
  --from-literal=host=your-db-host \
  --from-literal=username=your-username \
  --from-literal=password=your-password \
  --namespace=production

# ML 모델 관련 시크릿 (필요시)
kubectl create secret generic ml-secret \
  --from-literal=model-api-key=your-model-key \
  --namespace=production
```

### 네임스페이스 생성
```bash
kubectl create namespace production
kubectl create namespace development
```

## 📊 모니터링 설정

### 필수 헬스체크 엔드포인트
- **API Server**: `https://api.chuingho.com/health`
- **ML Service**: `https://api.chuingho.com/ml/health`
- **Database**: 내부 헬스체크

### 로그 수집 (권장)
```yaml
# Filebeat, Fluentd 등을 사용한 로그 수집
# Elasticsearch, CloudWatch, DataDog 등으로 전송
```

## 🔄 배포 플로우

### CI → CD 순서 보장
1. **CI Pipeline** (모든 브랜치)
   - 코드 품질 검사
   - 단위 테스트 실행
   - 통합 테스트 실행
   - 보안 스캔

2. **CD Pipeline** (main 브랜치/태그)
   - CI 성공 확인 후 시작
   - Docker 이미지 빌드
   - 개발 환경 배포 (main 브랜치)
   - 프로덕션 배포 (릴리즈 태그)

### 수동 승인 단계 (권장)
```yaml
# GitHub Environment Protection Rules 설정
# 프로덕션 배포 전 관리자 승인 필요
```

## 🚨 주의사항

### 1. 현재 주석 처리된 배포 명령어들
```bash
# 실제 배포 시 주석 해제 필요:
# ssh deploy@dev-server "docker-compose -f docker-compose.dev.yml up -d"
# kubectl apply -f k8s-deployment.yml
```

### 2. 환경별 설정 파일
- 개발: `config/development.yml`
- 프로덕션: `config/production.yml`
- 각 환경에 맞는 설정값 준비 필요

### 3. 데이터베이스 마이그레이션
```bash
# 배포 전 자동 마이그레이션 실행
migrate -path migrations -database "postgres://..." up
```

## 📞 지원 및 문의

배포 설정 중 문제가 발생하면:
1. GitHub Issues에 문의
2. 개발팀 Slack 채널 #deployments
3. 긴급상황: 온콜 담당자 연락

---
*마지막 업데이트: 2025-01-07*