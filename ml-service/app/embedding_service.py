"""임베딩 서비스 코어 로직"""

import os
import logging
import numpy as np
from typing import List, Optional, Union
from sentence_transformers import SentenceTransformer
import torch

logger = logging.getLogger(__name__)


class EmbeddingService:
    """한국어 Sentence-BERT 기반 임베딩 서비스"""
    
    def __init__(self, model_name: str = "BM-K/KoSimCSE-bert"):
        """
        임베딩 서비스 초기화
        
        Args:
            model_name: 사용할 모델명 (기본: BM-K/KoSimCSE-bert)
        """
        self.model_name = model_name
        self.model: Optional[SentenceTransformer] = None
        self.embedding_dimension = 768  # KoSimCSE-bert 기본 차원
        self.device = self._get_device()
        
        # ONNX Runtime 사용 여부
        self.use_onnx = os.getenv("ENABLE_ONNX", "false").lower() == "true"
        
        self._load_model()
    
    def _get_device(self) -> str:
        """사용할 디바이스 결정"""
        if torch.cuda.is_available():
            device = "cuda"
            logger.info(f"CUDA 사용 가능. GPU 디바이스: {torch.cuda.get_device_name()}")
        elif hasattr(torch.backends, 'mps') and torch.backends.mps.is_available():
            device = "mps"  # Apple Silicon Mac
            logger.info("MPS 사용 (Apple Silicon)")
        else:
            device = "cpu"
            logger.info("CPU 사용")
        
        return device
    
    def _load_model(self):
        """모델 로드"""
        try:
            logger.info(f"모델 로딩 시작: {self.model_name}")
            
            # 모델 로드 옵션 설정
            model_kwargs = {"device": self.device}
            if self.use_onnx:
                logger.info("ONNX Runtime 사용")
                # ONNX 사용시 추가 설정
                pass
            
            self.model = SentenceTransformer(self.model_name, **model_kwargs)
            
            # 임베딩 차원 확인
            test_embedding = self.model.encode("테스트", convert_to_numpy=True)
            self.embedding_dimension = len(test_embedding)
            
            logger.info(f"모델 로드 완료. 차원: {self.embedding_dimension}")
            
        except Exception as e:
            logger.error(f"모델 로드 실패: {e}")
            raise RuntimeError(f"모델 로드 실패: {e}")
    
    def encode_single(self, text: str) -> List[float]:
        """
        단일 텍스트 임베딩 생성
        
        Args:
            text: 임베딩할 텍스트
            
        Returns:
            768차원 임베딩 벡터
        """
        if not self.model:
            raise RuntimeError("모델이 로드되지 않았습니다")
        
        if not text or not text.strip():
            raise ValueError("비어있는 텍스트는 임베딩할 수 없습니다")
        
        try:
            # 텍스트 전처리
            processed_text = self._preprocess_text(text)
            
            # 임베딩 생성
            embedding = self.model.encode(
                processed_text,
                convert_to_numpy=True,
                show_progress_bar=False,
                normalize_embeddings=True  # 코사인 유사도 최적화
            )
            
            return embedding.tolist()
            
        except Exception as e:
            logger.error(f"단일 텍스트 임베딩 실패: {e}")
            raise RuntimeError(f"임베딩 생성 실패: {e}")
    
    def encode_batch(self, texts: List[str]) -> List[List[float]]:
        """
        배치 텍스트 임베딩 생성
        
        Args:
            texts: 임베딩할 텍스트 목록
            
        Returns:
            각 텍스트의 768차원 임베딩 벡터 목록
        """
        if not self.model:
            raise RuntimeError("모델이 로드되지 않았습니다")
        
        if not texts:
            raise ValueError("빈 텍스트 목록입니다")
        
        # 빈 텍스트 필터링 및 전처리
        processed_texts = []
        valid_indices = []
        
        for i, text in enumerate(texts):
            if text and text.strip():
                processed_texts.append(self._preprocess_text(text))
                valid_indices.append(i)
        
        if not processed_texts:
            raise ValueError("유효한 텍스트가 없습니다")
        
        try:
            # 배치 임베딩 생성
            embeddings = self.model.encode(
                processed_texts,
                convert_to_numpy=True,
                show_progress_bar=False,
                normalize_embeddings=True,
                batch_size=32  # 메모리 효율성을 위한 배치 크기
            )
            
            # 결과 재구성 (빈 텍스트 위치에 None 추가)
            result = []
            valid_idx = 0
            
            for i in range(len(texts)):
                if i in valid_indices:
                    result.append(embeddings[valid_idx].tolist())
                    valid_idx += 1
                else:
                    # 빈 텍스트의 경우 영 벡터 반환
                    result.append([0.0] * self.embedding_dimension)
            
            return result
            
        except Exception as e:
            logger.error(f"배치 텍스트 임베딩 실패: {e}")
            raise RuntimeError(f"배치 임베딩 생성 실패: {e}")
    
    def _preprocess_text(self, text: str) -> str:
        """
        텍스트 전처리
        
        Args:
            text: 원본 텍스트
            
        Returns:
            전처리된 텍스트
        """
        # 기본 정규화
        processed = text.strip()
        
        # 연속된 공백 정리
        import re
        processed = re.sub(r'\s+', ' ', processed)
        
        # 너무 긴 텍스트 자르기 (BERT 토큰 한계 고려)
        max_length = 512  # 토큰 기준이 아닌 문자 기준 근사치
        if len(processed) > max_length:
            processed = processed[:max_length]
            logger.warning(f"텍스트가 너무 길어 {max_length}자로 자름")
        
        return processed
    
    def get_model_info(self) -> dict:
        """모델 정보 반환"""
        return {
            "model_name": self.model_name,
            "embedding_dimension": self.embedding_dimension,
            "device": self.device,
            "use_onnx": self.use_onnx,
            "model_loaded": self.model is not None
        }
    
    def health_check(self) -> bool:
        """헬스체크: 모델이 정상 작동하는지 확인"""
        try:
            if not self.model:
                return False
            
            # 간단한 테스트 임베딩 생성
            test_embedding = self.encode_single("건강 상태 확인")
            return len(test_embedding) == self.embedding_dimension
            
        except Exception as e:
            logger.error(f"헬스체크 실패: {e}")
            return False