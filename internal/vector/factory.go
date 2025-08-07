package vector

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/config"
)

// NewVectorDB 설정에 따라 적절한 벡터 DB 클라이언트 생성
func NewVectorDB(ctx context.Context, cfg config.VectorConfig, logger *zap.Logger) (VectorDB, error) {
	switch strings.ToLower(cfg.Type) {
	case "faiss":
		db := NewFaissDB(cfg, logger)
		if err := db.Initialize(ctx); err != nil {
			return nil, fmt.Errorf("Faiss DB 초기화 실패: %w", err)
		}
		return db, nil
		
	case "chroma":
		// TODO: Chroma DB 구현 (필요시)
		return nil, fmt.Errorf("Chroma DB는 아직 구현되지 않았습니다")
		
	default:
		return nil, fmt.Errorf("지원하지 않는 벡터 DB 타입: %s", cfg.Type)
	}
}