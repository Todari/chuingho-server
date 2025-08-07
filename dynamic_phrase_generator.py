# 동적 형용사+명사 조합 생성기 (개념 증명)

import numpy as np
from typing import List, Tuple
import asyncio
import aiohttp

class DynamicPhraseGenerator:
    def __init__(self):
        # 형용사 풀 (예시)
        self.adjectives = [
            "아름다운", "빠른", "따뜻한", "차가운", "밝은", "어두운", 
            "부드러운", "단단한", "높은", "깊은", "넓은", "작은",
            "큰", "새로운", "오래된", "젊은", "조용한", "시끄러운",
            "달콤한", "쓴", "향기로운", "투명한", "불투명한", "반짝이는",
            "흐린", "맑은", "뜨거운", "시원한", "무거운", "가벼운"
        ]
        
        # 명사 풀 (예시)
        self.nouns = [
            "바람", "별", "바다", "꿈", "빛", "소리", "향기", "미소",
            "눈물", "웃음", "마음", "생각", "기억", "시간", "공간",
            "하늘", "구름", "달", "태양", "꽃", "나무", "잎", "뿌리",
            "물", "불", "흙", "공기", "산", "강", "호수", "길",
            "집", "문", "창", "거울", "그림", "책", "음악", "춤",
            "사랑", "희망", "꿈", "용기", "지혜", "평화", "자유"
        ]
    
    async def generate_semantic_combinations(self, 
                                           resume_text: str, 
                                           top_k: int = 3) -> List[str]:
        """
        자기소개서 텍스트를 기반으로 의미적으로 유사한 조합 생성
        """
        # 1단계: 자기소개서 임베딩 계산
        resume_embedding = await self.get_embedding(resume_text)
        
        # 2단계: 관련성 높은 형용사/명사 필터링
        relevant_adj = await self.filter_relevant_words(
            self.adjectives, resume_embedding, top_k=10
        )
        relevant_nouns = await self.filter_relevant_words(
            self.nouns, resume_embedding, top_k=20
        )
        
        # 3단계: 조합 생성 및 유사도 계산
        combinations = []
        for adj in relevant_adj:
            for noun in relevant_nouns:
                phrase = f"{adj} {noun}"
                combinations.append(phrase)
        
        # 4단계: 배치로 임베딩 계산 (성능 최적화)
        phrase_embeddings = await self.batch_get_embeddings(combinations)
        
        # 5단계: 유사도 계산 및 정렬
        similarities = []
        for i, phrase_emb in enumerate(phrase_embeddings):
            similarity = self.cosine_similarity(resume_embedding, phrase_emb)
            similarities.append((combinations[i], similarity))
        
        # 6단계: 상위 결과 반환
        similarities.sort(key=lambda x: x[1], reverse=True)
        return [phrase for phrase, _ in similarities[:top_k]]
    
    async def filter_relevant_words(self, 
                                   word_pool: List[str], 
                                   target_embedding: np.ndarray,
                                   top_k: int) -> List[str]:
        """
        타겟 임베딩과 유사한 단어들만 필터링
        """
        word_embeddings = await self.batch_get_embeddings(word_pool)
        similarities = []
        
        for i, word_emb in enumerate(word_embeddings):
            similarity = self.cosine_similarity(target_embedding, word_emb)
            similarities.append((word_pool[i], similarity))
        
        similarities.sort(key=lambda x: x[1], reverse=True)
        return [word for word, _ in similarities[:top_k]]
    
    async def get_embedding(self, text: str) -> np.ndarray:
        """
        ML 서비스에서 임베딩 벡터 가져오기
        """
        async with aiohttp.ClientSession() as session:
            async with session.post(
                'http://localhost:8001/embed',
                json={'text': text}
            ) as response:
                result = await response.json()
                return np.array(result['embedding'])
    
    async def batch_get_embeddings(self, texts: List[str]) -> List[np.ndarray]:
        """
        배치로 임베딩 계산 (성능 최적화)
        """
        async with aiohttp.ClientSession() as session:
            async with session.post(
                'http://localhost:8001/embed_batch',
                json={'texts': texts}
            ) as response:
                result = await response.json()
                return [np.array(emb) for emb in result['embeddings']]
    
    def cosine_similarity(self, a: np.ndarray, b: np.ndarray) -> float:
        """
        코사인 유사도 계산
        """
        return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

# 사용 예시
async def main():
    generator = DynamicPhraseGenerator()
    
    resume_text = """
    안녕하세요. 저는 창의적이고 열정적인 개발자입니다. 
    새로운 기술을 배우는 것을 좋아하며, 팀워크를 중시합니다.
    항상 밝은 에너지로 주변 사람들에게 긍정적인 영향을 주려고 노력합니다.
    """
    
    results = await generator.generate_semantic_combinations(resume_text)
    
    print("생성된 칭호들:")
    for phrase in results:
        print(f"  - {phrase}")

# 예상 결과:
# - 밝은 에너지
# - 새로운 꿈  
# - 창의적인 빛

if __name__ == "__main__":
    asyncio.run(main())