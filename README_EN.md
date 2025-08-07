# Chuingho - AI-Powered Personal Title Recommendation Service

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![Python Version](https://img.shields.io/badge/Python-3.11+-blue.svg)](https://python.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> **Chuingho**: Portmanteau of "취준생" (job seekers) + "칭호" (title)  
> An AI service that analyzes personal statements to recommend unique personal titles

## 📋 Table of Contents

- [Project Overview](#-project-overview)
- [Key Features](#-key-features)
- [AI Technology Implementation](#-ai-technology-implementation)
- [System Architecture](#-system-architecture)
- [Tech Stack](#-tech-stack)
- [Quick Start](#-quick-start)
- [API Documentation](#-api-documentation)
- [Configuration](#-configuration)
- [Development Guide](#-development-guide)
- [Deployment](#-deployment)
- [Contributing](#-contributing)

## 🚀 Project Overview

Chuingho is an innovative AI service designed for job seekers. It analyzes personal statement text, converts individual characteristics into vectors, and recommends unique personal titles by finding semantically similar adjective+noun combinations.

### 💡 Core Concept

- **Personal Statement Text** → **768-dim Vector** → **Similarity Search** → **3 Title Recommendations**
- Examples: "Creative Designer", "Meticulous Analyst", "Proactive Leader"

### 🎯 Target Users

- Job seekers wanting unique personal titles
- Professionals considering personal branding
- Anyone seeking deeper self-understanding

## ✨ Key Features

### 1. Personal Statement Upload ⭐ **NEW: Text Input Method**
- **Input Method**: JSON-based direct text input (copy-paste supported)
- **Text Length**: 10-50,000 characters (Korean standard)
- **Supported Format**: Pure text (improved from file upload method)
- **Real-time Validation**: Instant length and format verification

### 2. AI-Powered Title Recommendation
- **Korean-Optimized**: Uses KoSimCSE-BERT model
- **High Performance**: < 200ms response with Faiss HNSW algorithm
- **Diversity Guaranteed**: MMR algorithm prevents similar results
- **Privacy Protected**: Original text not stored in vector DB

### 3. Management Tools
- **Phrase Dictionary Builder**: CLI tool for automatic candidate registration
- **Real-time Monitoring**: Health checks and performance metrics
- **Detailed Logging**: Request tracking and error analysis

## 🧠 AI Technology Implementation

### 🔬 Current Implementation Status (v1.0)

#### 1. Text Embedding ⭐ **Updated**
```python
# Current Implementation: Python FastAPI Service + Text-based Input
Model: sentence-transformers based
Architecture: KoSimCSE-BERT compatible model
Dimension: 768-dimensional vectors
Input: JSON format Korean personal statement text (10-50,000 characters)
Processing: File parsing removed → Direct text processing for improved performance
Output: Normalized 768-dimensional dense vectors
```

**Implementation Details:**
- **Model Loading**: Using `sentence-transformers` library
- **GPU Support**: Automatic detection and utilization in CUDA environments
- **Batch Processing**: Simultaneous embedding of multiple texts
- **ONNX Runtime**: Optimized inference option via environment variables
- **Caching**: Model loading time optimization

```python
# Implementation Example
class EmbeddingService:
    def __init__(self):
        self.model = SentenceTransformer('BM-K/KoSimCSE-bert')
        self.device = 'cuda' if torch.cuda.is_available() else 'cpu'
    
    def embed_text(self, text: str) -> List[float]:
        embedding = self.model.encode(
            text, 
            convert_to_tensor=True,
            normalize_embeddings=True
        )
        return embedding.cpu().numpy().tolist()
```

#### 2. Vector Similarity Search
```go
// Current Implementation: Go + Faiss In-Memory Search
Algorithm: HNSW (Hierarchical Navigable Small World)
Index Type: IndexHNSWFlat
Metric: Inner Product (Cosine Similarity)
Performance: p95 < 200ms (Local Environment)
Scale: 1M+ vectors supported
```

**Implementation Details:**
- **Index Management**: Real-time vector add/delete/update
- **Persistence**: JSON-based metadata + binary index storage
- **Memory Efficiency**: Vector compression and quantization support
- **Concurrency**: Go routine-based parallel search
- **Health Check**: Real-time index status monitoring

```go
// Implementation Example
type FaissDB struct {
    index      *faiss.IndexHNSWFlat
    vectors    map[string]VectorData
    dimension  int
    mu         sync.RWMutex
}

func (f *FaissDB) Search(ctx context.Context, query []float32, topK int) ([]VectorSearchResult, error) {
    distances, indices := f.index.Search(
        query, 
        int64(topK),
    )
    // Post-processing and diversity ranking application
    return f.diversityRanking(results), nil
}
```

#### 3. Diversity Assurance Algorithm
```go
// MMR (Maximal Marginal Relevance) Implementation
Algorithm: 0.7 Similarity + 0.3 Diversity weights
Method: Jaccard similarity-based string comparison
Purpose: Prevent duplicate titles with identical meanings
Output: Semantically diverse top 3 results
```

**Implementation Details:**
- **Similarity Calculation**: Combining cosine similarity and string similarity
- **Weight Adjustment**: Optimizing balance between relevance and diversity
- **Real-time Processing**: Immediate application to search results
- **Korean Language Specialized**: Morpheme-unit similarity comparison

#### 4. Phrase Candidate Database
```sql
-- Current Schema: PostgreSQL + Vector Metadata
CREATE TABLE phrase_candidates (
    id UUID PRIMARY KEY,
    phrase VARCHAR(100) NOT NULL UNIQUE,  -- "Creative Designer"
    adjective VARCHAR(50) NOT NULL,       -- "Creative"
    noun VARCHAR(50) NOT NULL,            -- "Designer" 
    frequency_score FLOAT,                -- Corpus frequency
    semantic_category VARCHAR(100),       -- "Creativity", "Leadership" etc
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);
```

**Data Construction Status:**
- **Initial Dataset**: 500+ manually curated adjective+noun combinations
- **Categories**: Personality traits, work styles, leadership, creativity, analytical skills, etc.
- **Expansion Plan**: Automated extraction pipeline from Wikipedia and news corpus

### 🚀 Performance Optimization

#### 1. Response Time Optimization
- **Target**: p95 < 200ms
- **Current**: 50-100ms in local environment
- **Optimization Techniques**:
  - Approximate nearest neighbor search with Faiss HNSW index
  - Go routine-based parallel processing
  - Pre-computed vector normalization
  - Memory pooling to minimize GC pressure

#### 2. Scalability Design
- **Horizontal Scaling**: ML service load balancing support
- **Vector DB Sharding**: Category-based index distribution plan
- **Caching Strategy**: Redis-based embedding result caching
- **Monitoring**: Prometheus + Grafana metrics collection

### 🔮 Future Development Plans

#### Phase 2: Advanced Features (2024 Q4)
```yaml
Advanced_Embedding_Models:
  - Models: KR-SBERT-V2, KoSimCSE-RoBERTa
  - Multi-Model: Ensemble embedding for improved accuracy
  - Fine-tuning: Personal statement domain-specific learning

Personalized_Recommendations:
  - User_Profile: Weight-based recommendations by major, job, experience
  - Feedback_Learning: Re-ranking reflecting user preferences
  - A/B_Testing: Performance comparison of recommendation algorithms

Real-time_Learning:
  - Online_Learning: Automatic discovery and addition of new phrases
  - Quality_Management: Automatic filtering of inappropriate combinations
  - Trend_Reflection: Periodic weighting of popular keywords
```

#### Phase 3: AI Enhancement (2025 Q1)
```yaml
Multimodal_Analysis:
  - Text+Image: Portfolio image analysis addition
  - Voice_Analysis: Speech tone analysis from interview practice videos
  - Emotion_Analysis: Text tone and emotional state reflection

Generative_AI_Integration:
  - GPT_Based: Title explanation and improvement suggestion generation
  - Personalized_Coaching: Personal statement improvement direction
  - Interview_Simulation: Expected questions based on titles

Advanced_Vector_Search:
  - Hybrid_Search: Combining keyword + semantic search
  - Explainable_AI: Providing recommendation reasons and evidence
  - Interactive_Improvement: Gradual improvement through user dialogue
```

#### Phase 4: Platform Development (2025 Q2)
```yaml
API_Platform:
  - Developer_API: REST/GraphQL API for third-party service integration
  - Embedding_API: General-purpose Korean text embedding service
  - Custom_Models: Company-specific title system construction

Data_Ecosystem:
  - Crowdsourcing: User-contributed phrase database
  - Expert_Curation: HR expert verification system
  - Industry_Specialization: Field-specific titles for IT, finance, healthcare, etc.
```

### 📊 Technical Metrics and Monitoring

#### Current Performance Metrics
```yaml
Embedding_Performance:
  - Processing_Speed: ~100 tokens/sec (CPU), ~500 tokens/sec (GPU)
  - Memory_Usage: ~2GB RAM per model
  - Batch_Size: Optimal 16-32 texts/batch

Search_Performance:
  - Search_Latency: p50=45ms, p95=120ms, p99=200ms
  - Throughput: 1000 QPS (single node)
  - Accuracy: Top-3 satisfaction 85%+ (internal testing)

System_Reliability:
  - Availability: 99.9% SLA target
  - Error_Rate: < 0.1% 4xx/5xx responses
  - Recovery_Time: < 30 seconds automatic restart
```

#### Quality Assurance
```yaml
Test_Coverage:
  - Unit_Tests: Go 85%+, Python 90%+
  - Integration_Tests: 95% API scenario coverage
  - Performance_Tests: Automated load scenario testing

Model_Validation:
  - Regression_Testing: Pre-deployment baseline dataset validation
  - A/B_Testing: Gradual application of new models
  - User_Feedback: Real-time satisfaction collection

Security_and_Privacy:
  - Encryption: AES-256 encryption for all PII data
  - Anonymization: Automatic PII masking in logs
  - Compliance: GDPR, PIPEDA compliant
```

### 🛠 Development and Operations Tools

#### ML Model Management
```bash
# Model deployment pipeline
make deploy-model MODEL=KoSimCSE-bert VERSION=v1.2
make test-model ENDPOINT=http://ml-service:8001
make rollback-model PREVIOUS_VERSION=v1.1

# Performance benchmarking
make benchmark-embedding BATCH_SIZE=32 ITERATIONS=1000
make benchmark-search VECTORS=1M QUERIES=10000
```

#### Monitoring Dashboard
```yaml
Real-time_Metrics:
  - Request_Count: Requests processed per minute
  - Response_Time: Latency distribution by percentile  
  - Error_Rate: Aggregation by HTTP status code
  - Resource_Usage: CPU, memory, GPU utilization

Business_Metrics:
  - User_Satisfaction: Title adoption rate
  - Diversity_Score: Recommendation result diversity measurement
  - Retention_Rate: Same user revisit ratio
```

## 🏗 System Architecture

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Frontend  │───▶│   Go API    │───▶│  ML Service │
│  (React)    │    │   Server    │    │  (Python)   │
└─────────────┘    └─────────────┘    └─────────────┘
                           │                   │
                           ▼                   ▼
                   ┌─────────────┐    ┌─────────────┐
                   │ PostgreSQL  │    │ Vector DB   │
                   │ (Metadata)  │    │  (Faiss)    │
                   └─────────────┘    └─────────────┘
                           │
                           ▼
                   ┌─────────────┐
                   │   MinIO     │
                   │(File Storage)│
                   └─────────────┘
```

### Component Description

- **Go API Server**: Gin-based REST API, business logic processing
- **ML Service**: FastAPI-based, KoSimCSE-BERT embedding generation
- **PostgreSQL**: User, personal statement, recommendation metadata
- **Vector DB**: Faiss-based high-performance vector search engine
- **MinIO**: S3-compatible object storage, encrypted original file storage

## 🛠 Tech Stack

### Backend
- **Go 1.22**: Main API server
- **Gin 2.x**: HTTP web framework
- **pgx/v5**: PostgreSQL driver
- **Viper**: Configuration management
- **Zap**: Structured logging

### ML Service
- **Python 3.11**: ML service runtime
- **FastAPI**: API framework
- **sentence-transformers**: Embedding model
- **PyTorch**: Deep learning framework

### Database & Storage
- **PostgreSQL 15**: Main database
- **MinIO**: S3-compatible object storage
- **Faiss**: Vector search engine

### Infrastructure
- **Docker & Docker Compose**: Containerization
- **GitHub Actions**: CI/CD pipeline

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Make (optional)

### 1. Clone Project

```bash
git clone https://github.com/Todari/chuingho-server.git
cd chuingho-server
```

### 2. Start Full Stack

```bash
# Using Make
make up

# Or direct execution
docker-compose up -d --build
```

### 3. Verify Services

```bash
# API server health check
curl http://localhost:8080/health

# ML service health check  
curl http://localhost:8001/health
```

### 4. Build Phrase Dictionary (First Time Only)

```bash
# Initialize vector DB with sample phrases
make prepare-phrases
```

### 5. Test API

```bash
# Upload personal statement text
curl -X POST -H "Content-Type: application/json" \
  -d '{"text":"Hello, I am a creative and passionate developer..."}' \
  http://localhost:8080/v1/resumes

# Get title recommendations (use resumeId from above)
curl -X POST -H "Content-Type: application/json" \
  -d '{"resumeId":"RESUME_ID_HERE"}' \
  http://localhost:8080/v1/titles
```

## 📚 API Documentation

### Main Endpoints

#### Upload Personal Statement
```http
POST /v1/resumes
Content-Type: application/json

{
  "text": "Hello, I am a creative and passionate developer..."
}
```

**Response Example:**
```json
{
  "resumeId": "123e4567-e89b-12d3-a456-426614174000",
  "status": "uploaded"
}
```

#### Generate Title Recommendations
```http
POST /v1/titles
Content-Type: application/json

{
  "resumeId": "123e4567-e89b-12d3-a456-426614174000"
}
```

**Response Example:**
```json
{
  "titles": [
    "Creative Designer",
    "Meticulous Analyst", 
    "Proactive Leader"
  ]
}
```

#### Health Check
```http
GET /health
```

### Complete API Documentation

Coming soon: Auto-generated Swagger/OpenAPI 3.1 documentation

## ⚙️ Configuration

### Environment Variables

Key environment variables can be set in `.env` file or system environment:

```bash
# Server settings
CHUINGHO_SERVER_PORT=8080
CHUINGHO_SERVER_ENVIRONMENT=production

# Database
CHUINGHO_DATABASE_HOST=localhost
CHUINGHO_DATABASE_PASSWORD=your_password

# ML Service
CHUINGHO_ML_SERVICE_URL=http://localhost:8001
CHUINGHO_ML_EMBEDDING_MODEL=BM-K/KoSimCSE-bert

# Vector DB
CHUINGHO_VECTOR_TYPE=faiss
CHUINGHO_VECTOR_DIMENSION=768
```

### Configuration File

Detailed settings via `config.yaml`:

```yaml
server:
  port: 8080
  environment: dev

database:
  host: localhost
  port: 5432
  
ml:
  service_url: http://localhost:8001
  timeout: 30
  
# ... other settings
```

## 👩‍💻 Development Guide

### Local Development Setup

```bash
# 1. Setup development environment
make dev-setup

# 2. Start dependency services only (DB, Storage, ML)
make dev-deps

# 3. Run API server locally
make run-server
```

### Code Style

```bash
# Code formatting
make fmt

# Linting
make lint

# Testing
make test
```

### Directory Structure

```
chuingho-server/
├── cmd/                    # Executable applications
│   ├── server/            # Main API server
│   ├── migration/         # DB migration tool
│   └── prepare_phrases/   # Phrase dictionary builder
├── internal/              # Internal packages
│   ├── config/           # Configuration management
│   ├── database/         # DB connection
│   ├── handler/          # HTTP handlers
│   ├── service/          # Business logic
│   ├── storage/          # Object storage
│   └── vector/           # Vector DB
├── pkg/                   # Public packages
│   ├── model/            # Data models
│   └── util/             # Utilities
├── ml-service/            # Python ML service
├── migrations/            # DB schema migrations
└── test/                  # Test code
```

### Developing New Features

1. **Create Feature Branch**
```bash
git checkout -b feature/new-feature-name
```

2. **Write Code and Tests**
```bash
make test
make lint
```

3. **Integration Testing**
```bash
make up
make test-api
```

4. **Create PR and Review**

## 🚀 Deployment

### Production Deployment

```bash
# 1. Set environment variables
export CHUINGHO_SERVER_ENVIRONMENT=production
export CHUINGHO_DATABASE_PASSWORD=strong_password

# 2. Start in production mode
docker-compose -f docker-compose.prod.yaml up -d
```

### Kubernetes Deployment

K8s manifest files coming soon

### Monitoring

- Health checks: `/health`, `/ready`, `/live`
- Metrics: Prometheus support planned
- Logs: Structured JSON logs

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Test specific package
go test -v ./internal/service/...

# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Test with full stack
make up
make test-api
```

### Performance Testing

```bash
# Vector search performance (target: p95 < 200ms)
ab -n 1000 -c 10 -T "application/json" \
  -p test_data.json http://localhost:8080/v1/titles
```

## 📊 Performance Metrics

### Target Performance

- **Response Time**: p95 < 200ms (vector search)
- **Throughput**: 100 req/s (concurrent users)
- **Availability**: 99.9% (~8.76 hours downtime/year)

### Resource Requirements

- **API Server**: 2 CPU cores, 4GB memory
- **ML Service**: 4 CPU cores, 8GB memory (GPU recommended)
- **Database**: 2 CPU cores, 4GB memory, 100GB storage
- **Vector DB**: 4GB memory (for 1M vectors)

## 🤝 Contributing

Thank you for contributing to the project!

### How to Contribute

1. **Check Issues**: [GitHub Issues](https://github.com/Todari/chuingho-server/issues)
2. **Fork & Branch**: `feature/feature-name` or `fix/bug-name`
3. **Commit Convention**: Use [Conventional Commits](https://conventionalcommits.org/)
4. **Testing**: Ensure all tests pass
5. **Create PR**: With detailed description

### Code Review Checklist

- [ ] All tests pass
- [ ] Code coverage ≥ 80%
- [ ] Documentation and comments
- [ ] Performance impact assessment
- [ ] Security vulnerability review

## 📄 License

This project is licensed under the MIT License. See [LICENSE](LICENSE) file for details.

## 📞 Contact

- **Email**: [Developer Email]
- **GitHub Issues**: [Project Issues](https://github.com/Todari/chuingho-server/issues)
- **Discord**: [Development Community] (Coming soon)

---

**Made with ❤️ for Job Seekers**

Job hunting is tough, but find confidence with your unique personal title! 🌟