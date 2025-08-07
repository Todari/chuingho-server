package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Todari/chuingho-server/internal/config"
)

// DB 데이터베이스 연결 풀 래퍼
type DB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// New 새로운 데이터베이스 연결 생성
func New(ctx context.Context, cfg config.DatabaseConfig, logger *zap.Logger) (*DB, error) {
	// 연결 풀 설정
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 설정 파싱 실패: %w", err)
	}

	// 연결 풀 설정 조정
	poolConfig.MaxConns = int32(cfg.MaxConns)
	poolConfig.MinConns = int32(cfg.MinConns)
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = time.Minute * 30
	poolConfig.HealthCheckPeriod = time.Minute

	// 연결 풀 생성
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 풀 생성 실패: %w", err)
	}

	// 연결 테스트
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("데이터베이스 연결 테스트 실패: %w", err)
	}

	logger.Info("데이터베이스 연결 성공",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
		zap.Int("max_conns", cfg.MaxConns),
		zap.Int("min_conns", cfg.MinConns))

	return &DB{
		Pool:   pool,
		logger: logger,
	}, nil
}

// Close 데이터베이스 연결 종료
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("데이터베이스 연결 종료")
	}
}

// HealthCheck 데이터베이스 상태 확인
func (db *DB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return db.Pool.Ping(ctx)
}

// GetStats 연결 풀 통계 정보 반환
func (db *DB) GetStats() *pgxpool.Stat {
	return db.Pool.Stat()
}