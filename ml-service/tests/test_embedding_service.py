"""임베딩 서비스 테스트"""

from unittest.mock import Mock, patch

import numpy as np
import pytest
from app.embedding_service import EmbeddingService


class TestEmbeddingService:
    """임베딩 서비스 테스트 클래스"""

    @pytest.fixture
    def mock_model(self):
        """모킹된 SentenceTransformer 모델"""
        mock = Mock()
        mock.encode.return_value = np.array([0.1, 0.2, 0.3, 0.4])
        return mock

    @pytest.fixture
    def embedding_service(self, mock_model):
        """테스트용 임베딩 서비스"""
        with patch('app.embedding_service.SentenceTransformer') as mock_st:
            mock_st.return_value = mock_model
            service = EmbeddingService("test-model")
            service.embedding_dimension = 4  # 테스트용 작은 차원
            return service

    def test_encode_single_success(self, embedding_service):
        """단일 텍스트 임베딩 성공 테스트"""
        text = "테스트 텍스트입니다"
        result = embedding_service.encode_single(text)
        
        assert isinstance(result, list)
        assert len(result) == embedding_service.embedding_dimension
        assert all(isinstance(x, float) for x in result)

    def test_encode_single_empty_text(self, embedding_service):
        """빈 텍스트 임베딩 실패 테스트"""
        with pytest.raises(ValueError, match="비어있는 텍스트는 임베딩할 수 없습니다"):
            embedding_service.encode_single("")

    def test_encode_single_whitespace_only(self, embedding_service):
        """공백만 있는 텍스트 임베딩 실패 테스트"""
        with pytest.raises(ValueError, match="비어있는 텍스트는 임베딩할 수 없습니다"):
            embedding_service.encode_single("   \t\n   ")

    def test_encode_batch_success(self, embedding_service, mock_model):
        """배치 텍스트 임베딩 성공 테스트"""
        # 배치 결과를 위한 모킹 설정
        mock_model.encode.return_value = np.array([
            [0.1, 0.2, 0.3, 0.4],
            [0.5, 0.6, 0.7, 0.8],
            [0.9, 1.0, 1.1, 1.2]
        ])
        
        texts = ["첫 번째 텍스트", "두 번째 텍스트", "세 번째 텍스트"]
        results = embedding_service.encode_batch(texts)
        
        assert len(results) == 3
        assert all(isinstance(result, list) for result in results)
        assert all(len(result) == embedding_service.embedding_dimension for result in results)

    def test_encode_batch_with_empty_texts(self, embedding_service, mock_model):
        """빈 텍스트가 포함된 배치 임베딩 테스트"""
        # 유효한 텍스트만을 위한 모킹
        mock_model.encode.return_value = np.array([
            [0.1, 0.2, 0.3, 0.4],
            [0.5, 0.6, 0.7, 0.8]
        ])
        
        texts = ["첫 번째 텍스트", "", "세 번째 텍스트", "   "]
        results = embedding_service.encode_batch(texts)
        
        assert len(results) == 4
        # 빈 텍스트는 영 벡터로 처리
        assert results[1] == [0.0] * embedding_service.embedding_dimension
        assert results[3] == [0.0] * embedding_service.embedding_dimension

    def test_encode_batch_empty_list(self, embedding_service):
        """빈 텍스트 목록 임베딩 실패 테스트"""
        with pytest.raises(ValueError, match="빈 텍스트 목록입니다"):
            embedding_service.encode_batch([])

    def test_preprocess_text(self, embedding_service):
        """텍스트 전처리 테스트"""
        # private 메서드 테스트
        text = "  이것은   연속된    공백이   있는  텍스트입니다  "
        processed = embedding_service._preprocess_text(text)
        
        assert processed == "이것은 연속된 공백이 있는 텍스트입니다"

    def test_preprocess_text_long(self, embedding_service):
        """긴 텍스트 전처리 테스트"""
        long_text = "가" * 1000  # 512자보다 긴 텍스트
        processed = embedding_service._preprocess_text(long_text)
        
        assert len(processed) <= 512

    def test_get_model_info(self, embedding_service):
        """모델 정보 조회 테스트"""
        info = embedding_service.get_model_info()
        
        assert isinstance(info, dict)
        assert "model_name" in info
        assert "embedding_dimension" in info
        assert "device" in info
        assert "use_onnx" in info
        assert "model_loaded" in info
        
        assert info["model_loaded"] is True
        assert info["embedding_dimension"] == embedding_service.embedding_dimension

    def test_health_check_success(self, embedding_service):
        """헬스체크 성공 테스트"""
        result = embedding_service.health_check()
        assert result is True

    def test_health_check_no_model(self):
        """모델이 없을 때 헬스체크 실패 테스트"""
        service = EmbeddingService("test-model")
        service.model = None
        
        result = service.health_check()
        assert result is False

    @patch('app.embedding_service.torch')
    def test_get_device_cuda(self, mock_torch, embedding_service):
        """CUDA 디바이스 선택 테스트"""
        mock_torch.cuda.is_available.return_value = True
        mock_torch.cuda.get_device_name.return_value = "NVIDIA GeForce GTX 1080"
        
        device = embedding_service._get_device()
        assert device == "cuda"

    @patch('app.embedding_service.torch')
    def test_get_device_cpu(self, mock_torch, embedding_service):
        """CPU 디바이스 선택 테스트"""
        mock_torch.cuda.is_available.return_value = False
        mock_torch.backends.mps.is_available.return_value = False
        
        device = embedding_service._get_device()
        assert device == "cpu"

    def test_korean_text_encoding(self, embedding_service):
        """한글 텍스트 임베딩 테스트"""
        korean_text = "안녕하세요. 저는 창의적인 개발자입니다. 새로운 기술에 관심이 많고 문제 해결을 좋아합니다."
        result = embedding_service.encode_single(korean_text)
        
        assert isinstance(result, list)
        assert len(result) == embedding_service.embedding_dimension
        assert all(isinstance(x, float) for x in result)

    def test_mixed_language_encoding(self, embedding_service):
        """한영 혼합 텍스트 임베딩 테스트"""
        mixed_text = "Hello 안녕하세요 Python programming 프로그래밍"
        result = embedding_service.encode_single(mixed_text)
        
        assert isinstance(result, list)
        assert len(result) == embedding_service.embedding_dimension

    def test_special_characters_encoding(self, embedding_service):
        """특수문자 포함 텍스트 임베딩 테스트"""
        special_text = "이메일: test@example.com, 전화: 010-1234-5678, 특수문자: !@#$%^&*()"
        result = embedding_service.encode_single(special_text)
        
        assert isinstance(result, list)
        assert len(result) == embedding_service.embedding_dimension


class TestEmbeddingServiceIntegration:
    """임베딩 서비스 통합 테스트"""

    @pytest.fixture
    def real_service(self):
        """실제 모델을 사용한 서비스 (통합 테스트용)"""
        # 실제 환경에서는 작은 모델 사용 또는 스킵
        pytest.skip("실제 모델 로딩이 필요한 통합 테스트")
        return EmbeddingService("sentence-transformers/all-MiniLM-L6-v2")

    def test_real_model_consistency(self, real_service):
        """실제 모델의 일관성 테스트"""
        text = "테스트용 문장입니다"
        
        result1 = real_service.encode_single(text)
        result2 = real_service.encode_single(text)
        
        # 같은 입력에 대해 같은 결과를 반환해야 함
        assert result1 == result2

    def test_real_model_similarity(self, real_service):
        """실제 모델의 유사도 테스트"""
        text1 = "창의적인 개발자"
        text2 = "혁신적인 프로그래머"
        text3 = "맛있는 음식"
        
        vec1 = np.array(real_service.encode_single(text1))
        vec2 = np.array(real_service.encode_single(text2))
        vec3 = np.array(real_service.encode_single(text3))
        
        # 코사인 유사도 계산
        def cosine_similarity(a, b):
            return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))
        
        sim_12 = cosine_similarity(vec1, vec2)
        sim_13 = cosine_similarity(vec1, vec3)
        
        # 관련 있는 문장끼리 더 유사해야 함
        assert sim_12 > sim_13