"""FastAPI ML 임베딩 서비스 메인 애플리케이션"""

import os
import logging
import time
from contextlib import asynccontextmanager
from typing import List

from fastapi import FastAPI, HTTPException, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse

from .models import (
    EmbeddingRequest, EmbeddingResponse,
    BatchEmbeddingRequest, BatchEmbeddingResponse, PhraseEmbedding,
    HealthResponse, ErrorResponse
)
from .embedding_service import EmbeddingService

# 로깅 설정
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
logger = logging.getLogger(__name__)

# 전역 임베딩 서비스 인스턴스
embedding_service: EmbeddingService = None


@asynccontextmanager
async def lifespan(app: FastAPI):
    """애플리케이션 생명주기 관리"""
    global embedding_service
    
    # 시작시 모델 로드
    logger.info("ML 서비스 시작 - 모델 로딩 중...")
    model_name = os.getenv("EMBEDDING_MODEL", "BM-K/KoSimCSE-bert")
    
    try:
        embedding_service = EmbeddingService(model_name=model_name)
        logger.info("모델 로드 완료")
    except Exception as e:
        logger.error(f"모델 로드 실패: {e}")
        raise
    
    yield
    
    # 종료시 정리
    logger.info("ML 서비스 종료")


# FastAPI 앱 생성
app = FastAPI(
    title="췽호 ML 임베딩 서비스",
    description="한국어 자기소개서 텍스트를 벡터로 변환하는 ML 서비스",
    version="1.0.0",
    lifespan=lifespan
)

# CORS 설정
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # 개발환경, 운영환경에서는 특정 도메인만 허용
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.middleware("http")
async def add_process_time_header(request: Request, call_next):
    """응답 시간 측정 미들웨어"""
    start_time = time.time()
    response = await call_next(request)
    process_time = time.time() - start_time
    response.headers["X-Process-Time"] = str(process_time)
    return response


@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    """전역 예외 처리"""
    logger.error(f"전역 예외 발생: {exc}", exc_info=True)
    
    if isinstance(exc, HTTPException):
        return JSONResponse(
            status_code=exc.status_code,
            content=ErrorResponse(
                error=exc.detail,
                code=f"HTTP_{exc.status_code}"
            ).dict()
        )
    
    return JSONResponse(
        status_code=500,
        content=ErrorResponse(
            error="내부 서버 오류가 발생했습니다",
            code="INTERNAL_ERROR",
            details=str(exc)
        ).dict()
    )


@app.get("/health", response_model=HealthResponse)
async def health_check():
    """헬스체크 엔드포인트"""
    if not embedding_service:
        raise HTTPException(status_code=503, detail="임베딩 서비스가 초기화되지 않았습니다")
    
    model_info = embedding_service.get_model_info()
    is_healthy = embedding_service.health_check()
    
    if not is_healthy:
        raise HTTPException(status_code=503, detail="임베딩 서비스 상태 이상")
    
    return HealthResponse(
        status="healthy",
        model_loaded=model_info["model_loaded"],
        model_name=model_info["model_name"],
        embedding_dimension=model_info["embedding_dimension"]
    )


@app.post("/embed", response_model=EmbeddingResponse)
async def embed_text(request: EmbeddingRequest):
    """
    단일 텍스트 임베딩 생성
    
    자기소개서 텍스트를 768차원 벡터로 변환합니다.
    """
    if not embedding_service:
        raise HTTPException(status_code=503, detail="임베딩 서비스가 초기화되지 않았습니다")
    
    try:
        logger.info(f"단일 텍스트 임베딩 요청 (길이: {len(request.text)})")
        
        vector = embedding_service.encode_single(request.text)
        
        logger.info("단일 텍스트 임베딩 완료")
        return EmbeddingResponse(vector=vector)
        
    except ValueError as e:
        logger.warning(f"잘못된 입력: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"임베딩 생성 실패: {e}")
        raise HTTPException(status_code=500, detail="임베딩 생성 중 오류가 발생했습니다")


@app.post("/embed/phrases", response_model=BatchEmbeddingResponse)
async def embed_phrases(request: BatchEmbeddingRequest):
    """
    배치 구문 임베딩 생성
    
    여러 형용사+명사 구문을 한번에 벡터로 변환합니다.
    주로 후보 사전 구축시 사용됩니다.
    """
    if not embedding_service:
        raise HTTPException(status_code=503, detail="임베딩 서비스가 초기화되지 않았습니다")
    
    try:
        logger.info(f"배치 임베딩 요청 (구문 수: {len(request.phrases)})")
        
        vectors = embedding_service.encode_batch(request.phrases)
        
        results = [
            PhraseEmbedding(phrase=phrase, vector=vector)
            for phrase, vector in zip(request.phrases, vectors)
        ]
        
        logger.info(f"배치 임베딩 완료 (처리된 구문 수: {len(results)})")
        return BatchEmbeddingResponse(results=results)
        
    except ValueError as e:
        logger.warning(f"잘못된 입력: {e}")
        raise HTTPException(status_code=400, detail=str(e))
    except Exception as e:
        logger.error(f"배치 임베딩 생성 실패: {e}")
        raise HTTPException(status_code=500, detail="배치 임베딩 생성 중 오류가 발생했습니다")


@app.get("/")
async def root():
    """루트 엔드포인트"""
    return {
        "service": "췽호 ML 임베딩 서비스",
        "version": "1.0.0",
        "status": "running",
        "endpoints": {
            "health": "/health",
            "embed": "/embed",
            "batch_embed": "/embed/phrases",
            "docs": "/docs"
        }
    }


if __name__ == "__main__":
    import uvicorn
    
    port = int(os.getenv("PORT", "8001"))
    host = os.getenv("HOST", "0.0.0.0")
    
    uvicorn.run(
        "app.main:app",
        host=host,
        port=port,
        reload=False,  # 운영환경에서는 False
        log_level="info"
    )