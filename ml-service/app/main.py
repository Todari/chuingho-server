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


@app.post("/generate_dynamic_combinations")
async def generate_dynamic_combinations(request: dict):
    """
    동적 형용사+명사 조합 생성 및 임베딩
    자기소개서 텍스트를 기반으로 의미적으로 관련성 높은 조합들을 실시간 생성
    """
    if not embedding_service:
        raise HTTPException(status_code=503, detail="임베딩 서비스가 초기화되지 않았습니다")
    
    try:
        resume_text = request.get("resume_text", "")
        top_k = request.get("top_k", 3)
        adj_filter_count = request.get("adj_filter_count", 20)
        noun_filter_count = request.get("noun_filter_count", 30)
        
        logger.info(f"동적 조합 생성 요청: resume_length={len(resume_text)}, top_k={top_k}")
        start_time = time.time()
        
        # 1단계: 자기소개서 임베딩 생성
        resume_vector = embedding_service.encode([resume_text])[0]
        
        # 2단계: 형용사/명사 풀 로드
        adjectives = load_word_pool("adjectives")
        nouns = load_word_pool("nouns")
        
        # 3단계: 의미적 유사도 기반 필터링
        relevant_adjectives = filter_relevant_words(
            adjectives, resume_vector, adj_filter_count
        )
        relevant_nouns = filter_relevant_words(
            nouns, resume_vector, noun_filter_count
        )
        
        # 4단계: 조합 생성
        combinations = []
        for adj in relevant_adjectives:
            for noun in relevant_nouns:
                combinations.append(f"{adj} {noun}")
        
        logger.info(f"생성된 조합 수: {len(combinations)}")
        
        # 5단계: 조합들의 임베딩 배치 생성
        combination_vectors = embedding_service.encode_batch(combinations)
        
        # 6단계: 자기소개서와 조합들 간 유사도 계산
        similarities = []
        for i, combination_vector in enumerate(combination_vectors):
            similarity = cosine_similarity(resume_vector, combination_vector)
            similarities.append({
                "phrase": combinations[i],
                "similarity": float(similarity)
            })
        
        # 7단계: 유사도 기준 정렬 및 다양성 고려 선택
        similarities.sort(key=lambda x: x["similarity"], reverse=True)

        # 상위 유사 조합 5개 (MMR 적용 전)
        top_similar = similarities[:5]

        # MMR 알고리즘으로 다양성 고려
        final_results = select_diverse_combinations(similarities, top_k)
        
        processing_time = time.time() - start_time
        
        logger.info(f"동적 조합 생성 완료: {len(final_results)}개, {processing_time:.2f}초")
        
        return {
            "combinations": [result["phrase"] for result in final_results],
            "details": final_results,
            "top_similar": top_similar,
            "processing_time": processing_time,
            "total_generated": len(combinations),
            "filtered_adjectives": len(relevant_adjectives),
            "filtered_nouns": len(relevant_nouns)
        }
        
    except Exception as e:
        logger.error(f"동적 조합 생성 오류: {str(e)}")
        raise HTTPException(status_code=500, detail="동적 조합 생성 중 오류가 발생했습니다")


def load_word_pool(word_type: str) -> List[str]:
    """형용사/명사 풀 로드"""
    import os
    
    file_path = f"../data/{word_type}.txt"
    if not os.path.exists(file_path):
        # 기본 단어들 반환
        if word_type == "adjectives":
            return ["아름다운", "따뜻한", "밝은", "새로운", "창의적인", "독창적인", "혁신적인", 
                   "차가운", "깊은", "높은", "부드러운", "강한", "빠른", "느린", "큰", "작은"]
        else:  # nouns
            return ["바람", "별", "꿈", "빛", "마음", "생각", "미래", "희망", "에너지", "열정",
                   "바다", "하늘", "구름", "꽃", "나무", "물", "불", "길", "집", "문"]
    
    words = []
    with open(file_path, 'r', encoding='utf-8') as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#'):
                words.append(line)
    return words


def filter_relevant_words(words: List[str], target_vector: List[float], top_k: int) -> List[str]:
    """타겟 벡터와 유사한 단어들만 필터링"""
    if len(words) <= top_k:
        return words
    
    # 배치로 임베딩 생성
    word_vectors = embedding_service.encode_batch(words)
    
    # 유사도 계산
    similarities = []
    for i, word_vector in enumerate(word_vectors):
        similarity = cosine_similarity(target_vector, word_vector)
        similarities.append((words[i], similarity))
    
    # 상위 k개 선택
    similarities.sort(key=lambda x: x[1], reverse=True)
    return [word for word, _ in similarities[:top_k]]


def cosine_similarity(a: List[float], b: List[float]) -> float:
    """코사인 유사도 계산"""
    import numpy as np
    a = np.array(a)
    b = np.array(b)
    return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))


def select_diverse_combinations(similarities: List[dict], top_k: int) -> List[dict]:
    """MMR 알고리즘으로 다양성 고려한 조합 선택"""
    if len(similarities) <= top_k:
        return similarities
    
    selected = []
    remaining = similarities.copy()
    
    # 첫 번째는 가장 유사도 높은 것 선택
    selected.append(remaining.pop(0))
    
    # 나머지는 유사도와 다양성 고려
    while len(selected) < top_k and remaining:
        best_score = -1
        best_idx = 0
        
        for i, candidate in enumerate(remaining):
            # 유사도 점수 (70%)
            relevance_score = candidate["similarity"] * 0.7
            
            # 다양성 점수 (30%)
            diversity_score = calculate_diversity_score(candidate, selected) * 0.3
            
            total_score = relevance_score + diversity_score
            
            if total_score > best_score:
                best_score = total_score
                best_idx = i
        
        selected.append(remaining.pop(best_idx))
    
    return selected


def calculate_diversity_score(candidate: dict, selected: List[dict]) -> float:
    """다양성 점수 계산"""
    if not selected:
        return 1.0
    
    min_similarity = 1.0
    candidate_words = candidate["phrase"].split()
    
    for sel in selected:
        selected_words = sel["phrase"].split()
        # Jaccard 유사도 계산
        intersection = len(set(candidate_words) & set(selected_words))
        union = len(set(candidate_words) | set(selected_words))
        jaccard_sim = intersection / union if union > 0 else 0
        
        if jaccard_sim < min_similarity:
            min_similarity = jaccard_sim
    
    return 1.0 - min_similarity  # 유사도가 낮을수록 다양성 높음


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