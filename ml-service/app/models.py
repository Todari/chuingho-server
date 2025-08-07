"""ML 서비스 데이터 모델"""

from typing import List, Optional
from pydantic import BaseModel, Field


class EmbeddingRequest(BaseModel):
    """단일 텍스트 임베딩 요청"""
    text: str = Field(..., min_length=1, max_length=10000, description="임베딩할 텍스트")


class EmbeddingResponse(BaseModel):
    """단일 텍스트 임베딩 응답"""
    vector: List[float] = Field(..., description="768차원 임베딩 벡터")


class BatchEmbeddingRequest(BaseModel):
    """배치 텍스트 임베딩 요청"""
    phrases: List[str] = Field(..., min_items=1, max_items=1000, description="임베딩할 텍스트 목록")


class PhraseEmbedding(BaseModel):
    """구문과 벡터 쌍"""
    phrase: str = Field(..., description="원본 구문")
    vector: List[float] = Field(..., description="768차원 임베딩 벡터")


class BatchEmbeddingResponse(BaseModel):
    """배치 텍스트 임베딩 응답"""
    results: List[PhraseEmbedding] = Field(..., description="구문별 임베딩 결과")


class HealthResponse(BaseModel):
    """헬스체크 응답"""
    status: str = Field(..., description="서비스 상태")
    model_loaded: bool = Field(..., description="모델 로드 상태")
    model_name: str = Field(..., description="사용 중인 모델명")
    embedding_dimension: int = Field(..., description="임베딩 차원")


class ErrorResponse(BaseModel):
    """에러 응답"""
    error: str = Field(..., description="에러 메시지")
    code: Optional[str] = Field(None, description="에러 코드")
    details: Optional[str] = Field(None, description="상세 정보")