"""FastAPI 메인 애플리케이션 테스트"""

import pytest
from httpx import AsyncClient
from unittest.mock import Mock, patch
from app.main import app
from app.models import EmbeddingRequest, BatchEmbeddingRequest


@pytest.fixture
def mock_embedding_service():
    """모킹된 임베딩 서비스"""
    mock = Mock()
    mock.encode_single.return_value = [0.1, 0.2, 0.3, 0.4]
    mock.encode_batch.return_value = {
        "창의적 설계자": [0.1, 0.2, 0.3, 0.4],
        "세심한 분석가": [0.5, 0.6, 0.7, 0.8]
    }
    mock.get_model_info.return_value = {
        "model_name": "test-model",
        "embedding_dimension": 4,
        "device": "cpu",
        "use_onnx": False,
        "model_loaded": True
    }
    mock.health_check.return_value = True
    return mock


@pytest.fixture
async def client():
    """테스트 클라이언트"""
    async with AsyncClient(app=app, base_url="http://test") as ac:
        yield ac


class TestMainApp:
    """메인 애플리케이션 테스트"""

    async def test_root_endpoint(self, client):
        """루트 엔드포인트 테스트"""
        response = await client.get("/")
        
        assert response.status_code == 200
        data = response.json()
        assert "service" in data
        assert "췽호 ML 임베딩 서비스" in data["service"]
        assert "endpoints" in data

    @patch('app.main.embedding_service')
    async def test_health_endpoint_success(self, mock_service, client):
        """헬스체크 엔드포인트 성공 테스트"""
        mock_service.get_model_info.return_value = {
            "model_name": "test-model",
            "embedding_dimension": 768,
            "device": "cpu",
            "model_loaded": True
        }
        mock_service.health_check.return_value = True
        
        response = await client.get("/health")
        
        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["model_loaded"] is True
        assert data["embedding_dimension"] == 768

    @patch('app.main.embedding_service')
    async def test_health_endpoint_unhealthy(self, mock_service, client):
        """헬스체크 엔드포인트 실패 테스트"""
        mock_service.get_model_info.return_value = {
            "model_name": "test-model",
            "embedding_dimension": 768,
            "device": "cpu",
            "model_loaded": True
        }
        mock_service.health_check.return_value = False
        
        response = await client.get("/health")
        
        assert response.status_code == 503

    @patch('app.main.embedding_service')
    async def test_health_endpoint_no_service(self, mock_service, client):
        """임베딩 서비스가 없을 때 헬스체크 테스트"""
        # embedding_service를 None으로 설정
        with patch('app.main.embedding_service', None):
            response = await client.get("/health")
            
            assert response.status_code == 503

    @patch('app.main.embedding_service')
    async def test_embed_endpoint_success(self, mock_service, client):
        """단일 임베딩 엔드포인트 성공 테스트"""
        mock_service.encode_single.return_value = [0.1, 0.2, 0.3, 0.4]
        
        request_data = {"text": "테스트 텍스트입니다"}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "vector" in data
        assert data["vector"] == [0.1, 0.2, 0.3, 0.4]
        
        # 서비스가 올바른 인자로 호출되었는지 확인
        mock_service.encode_single.assert_called_once_with("테스트 텍스트입니다")

    @patch('app.main.embedding_service')
    async def test_embed_endpoint_invalid_request(self, mock_service, client):
        """잘못된 요청으로 임베딩 엔드포인트 테스트"""
        # 텍스트 필드 누락
        request_data = {}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 422  # Validation error

    @patch('app.main.embedding_service')
    async def test_embed_endpoint_service_error(self, mock_service, client):
        """임베딩 서비스 에러 테스트"""
        mock_service.encode_single.side_effect = ValueError("테스트 에러")
        
        request_data = {"text": "테스트 텍스트"}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 400

    @patch('app.main.embedding_service')
    async def test_embed_endpoint_no_service(self, mock_service, client):
        """임베딩 서비스가 없을 때 임베딩 엔드포인트 테스트"""
        with patch('app.main.embedding_service', None):
            request_data = {"text": "테스트 텍스트"}
            response = await client.post("/embed", json=request_data)
            
            assert response.status_code == 503

    @patch('app.main.embedding_service')
    async def test_embed_phrases_endpoint_success(self, mock_service, client):
        """배치 임베딩 엔드포인트 성공 테스트"""
        mock_service.encode_batch.return_value = {
            "창의적 설계자": [0.1, 0.2, 0.3, 0.4],
            "세심한 분석가": [0.5, 0.6, 0.7, 0.8]
        }
        
        request_data = {"phrases": ["창의적 설계자", "세심한 분석가"]}
        response = await client.post("/embed/phrases", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "results" in data
        assert len(data["results"]) == 2
        
        # 결과 구조 확인
        for result in data["results"]:
            assert "phrase" in result
            assert "vector" in result
            assert len(result["vector"]) == 4

    @patch('app.main.embedding_service')
    async def test_embed_phrases_endpoint_empty_list(self, mock_service, client):
        """빈 구문 목록으로 배치 임베딩 테스트"""
        request_data = {"phrases": []}
        response = await client.post("/embed/phrases", json=request_data)
        
        assert response.status_code == 422  # Validation error

    @patch('app.main.embedding_service')
    async def test_embed_phrases_endpoint_too_many(self, mock_service, client):
        """너무 많은 구문으로 배치 임베딩 테스트"""
        # 1000개 초과
        phrases = [f"구문 {i}" for i in range(1001)]
        request_data = {"phrases": phrases}
        response = await client.post("/embed/phrases", json=request_data)
        
        assert response.status_code == 422  # Validation error

    async def test_cors_headers(self, client):
        """CORS 헤더 테스트"""
        response = await client.options("/embed")
        
        assert response.status_code == 204
        assert "access-control-allow-origin" in response.headers
        assert "access-control-allow-methods" in response.headers

    async def test_process_time_header(self, client, mock_embedding_service):
        """처리 시간 헤더 테스트"""
        with patch('app.main.embedding_service', mock_embedding_service):
            response = await client.get("/health")
            
            assert "x-process-time" in response.headers
            # 처리 시간이 기록되었는지 확인
            process_time = response.headers["x-process-time"]
            assert process_time is not None

    @patch('app.main.embedding_service')
    async def test_korean_text_embedding(self, mock_service, client):
        """한글 텍스트 임베딩 테스트"""
        mock_service.encode_single.return_value = [0.1, 0.2, 0.3, 0.4]
        
        korean_text = "안녕하세요. 저는 창의적인 개발자입니다."
        request_data = {"text": korean_text}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 200
        data = response.json()
        assert "vector" in data
        
        # 한글 텍스트가 올바르게 전달되었는지 확인
        mock_service.encode_single.assert_called_once_with(korean_text)

    @patch('app.main.embedding_service')
    async def test_long_text_embedding(self, mock_service, client):
        """긴 텍스트 임베딩 테스트"""
        mock_service.encode_single.return_value = [0.1, 0.2, 0.3, 0.4]
        
        # 매우 긴 텍스트 (10000자)
        long_text = "가" * 10000
        request_data = {"text": long_text}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 422  # Validation error (max_length=10000)

    async def test_request_validation(self, client):
        """요청 검증 테스트"""
        # 잘못된 JSON
        response = await client.post("/embed", data="invalid json")
        assert response.status_code == 422

        # 필수 필드 누락
        response = await client.post("/embed", json={})
        assert response.status_code == 422

        # 잘못된 데이터 타입
        response = await client.post("/embed", json={"text": 123})
        assert response.status_code == 422


class TestExceptionHandling:
    """예외 처리 테스트"""

    @patch('app.main.embedding_service')
    async def test_internal_server_error(self, mock_service, client):
        """내부 서버 에러 처리 테스트"""
        mock_service.encode_single.side_effect = Exception("예상치 못한 에러")
        
        request_data = {"text": "테스트 텍스트"}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 500
        data = response.json()
        assert "error" in data
        assert "내부 서버 오류가 발생했습니다" in data["error"]

    @patch('app.main.embedding_service')
    async def test_value_error_handling(self, mock_service, client):
        """ValueError 처리 테스트"""
        mock_service.encode_single.side_effect = ValueError("잘못된 입력")
        
        request_data = {"text": "테스트 텍스트"}
        response = await client.post("/embed", json=request_data)
        
        assert response.status_code == 400
        data = response.json()
        assert "error" in data