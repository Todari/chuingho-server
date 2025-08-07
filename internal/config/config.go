package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 애플리케이션 전체 설정 구조체
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Storage  StorageConfig  `mapstructure:"storage"`
	ML       MLConfig       `mapstructure:"ml"`
	Vector   VectorConfig   `mapstructure:"vector"`
	Log      LogConfig      `mapstructure:"log"`
}

// ServerConfig 서버 관련 설정
type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Host         string `mapstructure:"host"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
	Environment  string `mapstructure:"environment"` // dev, staging, prod
}

// DatabaseConfig PostgreSQL 데이터베이스 설정
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	MaxConns int    `mapstructure:"max_conns"`
	MinConns int    `mapstructure:"min_conns"`
}

// StorageConfig S3 호환 객체 스토리지 설정
type StorageConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	BucketName      string `mapstructure:"bucket_name"`
	Region          string `mapstructure:"region"`
	UseSSL          bool   `mapstructure:"use_ssl"`
	PresignedExpiry int    `mapstructure:"presigned_expiry"` // seconds
}

// MLConfig ML 서비스 설정
type MLConfig struct {
	ServiceURL     string `mapstructure:"service_url"`
	Timeout        int    `mapstructure:"timeout"`
	RetryCount     int    `mapstructure:"retry_count"`
	EmbeddingModel string `mapstructure:"embedding_model"`
}

// VectorConfig 벡터 DB 설정
type VectorConfig struct {
	Type       string `mapstructure:"type"`        // faiss, chroma
	Host       string `mapstructure:"host"`        // chroma 사용시
	Port       int    `mapstructure:"port"`        // chroma 사용시
	IndexPath  string `mapstructure:"index_path"`  // faiss 사용시
	Dimension  int    `mapstructure:"dimension"`   // 768 for KoSentenceBERT
	MetricType string `mapstructure:"metric_type"` // IP, L2
}

// LogConfig 로그 설정
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"` // json, console
	OutputPath string `mapstructure:"output_path"`
}

// LoadConfig 설정 파일을 읽고 Config 구조체를 반환
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("/etc/chuingho")

	// 환경변수 설정
	viper.AutomaticEnv()
	viper.SetEnvPrefix("CHUINGHO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 기본값 설정
	setDefaultValues()

	// 설정 파일 읽기
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 설정 파일이 없으면 환경변수와 기본값만 사용
			fmt.Printf("설정 파일을 찾을 수 없습니다. 환경변수와 기본값을 사용합니다.\n")
		} else {
			return nil, fmt.Errorf("설정 파일 읽기 실패: %w", err)
		}
	} else {
		fmt.Printf("설정 파일 로드됨: %s\n", viper.ConfigFileUsed())
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("설정 파싱 실패: %w", err)
	}

	return &config, nil
}

// setDefaultValues 기본값 설정
func setDefaultValues() {
	// 서버 기본값
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.read_timeout", 10)
	viper.SetDefault("server.write_timeout", 10)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("server.environment", "dev")

	// 데이터베이스 기본값
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.username", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "chuingho")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.min_conns", 5)

	// 스토리지 기본값
	viper.SetDefault("storage.endpoint", "localhost:9000")
	viper.SetDefault("storage.access_key_id", "minioadmin")
	viper.SetDefault("storage.secret_access_key", "minioadmin")
	viper.SetDefault("storage.bucket_name", "chuingho-resumes")
	viper.SetDefault("storage.region", "us-east-1")
	viper.SetDefault("storage.use_ssl", false)
	viper.SetDefault("storage.presigned_expiry", 3600)

	// ML 서비스 기본값
	viper.SetDefault("ml.service_url", "http://localhost:8001")
	viper.SetDefault("ml.timeout", 30)
	viper.SetDefault("ml.retry_count", 3)
	viper.SetDefault("ml.embedding_model", "BM-K/KoSimCSE-bert")

	// 벡터 DB 기본값
	viper.SetDefault("vector.type", "faiss")
	viper.SetDefault("vector.host", "localhost")
	viper.SetDefault("vector.port", 8000)
	viper.SetDefault("vector.index_path", "./faiss_index")
	viper.SetDefault("vector.dimension", 768)
	viper.SetDefault("vector.metric_type", "IP")

	// 로그 기본값
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "console")
	viper.SetDefault("log.output_path", "stdout")
}

// GetDSN PostgreSQL 연결 문자열 생성
func (db DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.Username, db.Password, db.DBName, db.SSLMode)
}

// IsProduction 프로덕션 환경인지 확인
func (s ServerConfig) IsProduction() bool {
	return s.Environment == "prod" || s.Environment == "production"
}

// GetLogLevel zap 로그 레벨 반환
func (l LogConfig) GetLogLevel() zap.AtomicLevel {
	switch strings.ToLower(l.Level) {
	case "debug":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn", "warning":
		return zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case "fatal":
		return zap.NewAtomicLevelAt(zap.FatalLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}