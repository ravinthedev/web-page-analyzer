# Web Page Analyzer - Architecture

This document covers the technical design and implementation details of the web page analyzer.

## Overview

The system is built with Go backend and React frontend, following clean architecture patterns. It analyzes web pages both synchronously and asynchronously, storing results in PostgreSQL with Redis caching.

## System Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                                FRONTEND                                         │
│                              (React App)                                        │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐          │
│  │ URL Input   │   │ Validation  │   │ API Client  │   │ Results UI  │          │
│  │ Form        │   │ & Error     │   │ (Fetch)     │   │ Display     │          │
│  └─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘          │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ HTTP Requests
                                        │ (JSON API)
                                        ▼
┌────────────────────────────────────────────────────────────────────────────────┐
│                               BACKEND API                                      │
│                                (Go + Gin)                                      │
├────────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
│  ┌─────────────────────────────────────────────────────────────────────────────┤
│  │                        HTTP LAYER                                           │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                        │
│  │  │ Gin Router  │   │ Middleware  │   │ Handlers    │                        │
│  │  │ & Routes    │   │ (CORS, Rate │   │ (API Logic) │                        │
│  │  │             │   │ Limiting)   │   │             │                        │
│  │  └─────────────┘   └─────────────┘   └─────────────┘                        │
│  └─────────────────────────────────────────────────────────────────────────────┤
│                                        │                                       │
│                                        ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┤
│  │                    APPLICATION LAYER                                        │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                        │
│  │  │ Use Cases   │   │ Business    │   │ Async Job   │                        │
│  │  │ (Analyze,   │   │ Logic &     │   │ Processing  │                        │
│  │  │ GetResult)  │   │ Validation  │   │ (Goroutines)│                        │
│  │  └─────────────┘   └─────────────┘   └─────────────┘                        │
│  └─────────────────────────────────────────────────────────────────────────────┤
│                                        │                                       │
│                                        ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┤
│  │                      DOMAIN LAYER                                           │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                        │
│  │  │ Analyzer    │   │ HTML Parser │   │ Entities    │                        │
│  │  │ Service     │   │ (Link Check,│   │ (Analysis,  │                        │
│  │  │             │   │ Content)    │   │ Results)    │                        │
│  │  └─────────────┘   └─────────────┘   └─────────────┘                        │
│  └─────────────────────────────────────────────────────────────────────────────┤
│                                        │                                       │
│                                        ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────────────────┤
│  │                  INFRASTRUCTURE LAYER                                       │
│  │  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐                        │
│  │  │ PostgreSQL  │   │ Redis Cache │   │ HTTP Client │                        │
│  │  │ (Analysis   │   │ (Results,   │   │ (External   │                        │
│  │  │ Storage)    │   │ 1hr TTL)    │   │ Web Pages)  │                        │
│  │  └─────────────┘   └─────────────┘   └─────────────┘                        │
│  └─────────────────────────────────────────────────────────────────────────────┤
│                                                                                │
└────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              MONITORING                                         │
│  ┌─────────────┐   ┌─────────────┐   ┌─────────────┐   ┌─────────────┐          │
│  │ Prometheus  │   │ Health      │   │ Structured  │   │ Request     │          │
│  │ Metrics     │   │ Checks      │   │ Logging     │   │ Tracing     │          │
│  └─────────────┘   └─────────────┘   └─────────────┘   └─────────────┘          │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Architecture Layers

The application follows clean architecture with four main layers:

### Frontend (React)
- **URL Input Form**: Client-side validation with protocol normalization
- **Processing Modes**: Toggle between synchronous and asynchronous processing
- **Job Management**: Real-time job status tracking with polling
- **Results Display**: 
  - HTML version detection
  - Page title extraction
  - Headings distribution (H1-H6 counts)
  - Link analysis (internal/external/inaccessible counts)
  - Login form detection
  - Performance metrics (load time, content size, status code)
  - External hosts listing
- **Error Handling**: Comprehensive error display with correlation IDs
- **API Communication**: Fetch-based HTTP client with configurable endpoints

### Backend Layers

**Presentation Layer**
- Gin HTTP router with versioned API endpoints (`/api/v1/`)
- Comprehensive middleware stack:
  - CORS with configurable origins
  - Rate limiting (100 req/min per IP)
  - Request correlation ID tracking
  - Structured logging with Zap
  - Error handling and panic recovery
  - Request size limits
  - Authentication middleware (X-User-ID header)
- Health check endpoints (`/health`, `/health/live`, `/health/ready`)
- Prometheus metrics endpoint (`/metrics`)

**Application Layer**
- Use cases: `AnalyzeURL`, `SubmitAnalysisJob`, `ProcessAnalysisAsync`, `GetAnalysis`, `ListAnalyses`
- Business logic orchestration with context propagation
- Input validation and comprehensive error handling
- Cache-first strategy with TTL management
- Async job processing with goroutines

**Domain Layer**
- Core entities: `Analysis`, `AnalysisResult`, `AnalysisJob`, `LinkAnalysis`
- Services: `AnalyzerService` with HTML parsing, link checking, performance metrics
- Repository interfaces: `AnalysisRepository`, `CacheRepository`
- Status management: pending, processing, completed, failed

**Infrastructure Layer**
- PostgreSQL with JSONB storage for flexible result data
- Redis distributed caching with 1-hour TTL
- HTTP client with configurable timeouts and connection pooling
- Prometheus metrics collection (request duration, cache hits/misses, connection counts)
- Database migration system

### Data Flow

**Synchronous Processing:**
1. Frontend sends URL to `/api/v1/analyze` or `/api/analyze`
2. Handler validates JSON input and extracts correlation ID
3. Use case checks Redis cache first for existing results
4. If cache miss, checks PostgreSQL for recent analysis (within TTL)
5. Creates new analysis record with status tracking
6. Analyzer service fetches URL, parses HTML, checks links
7. Results stored in PostgreSQL, cached in Redis with TTL
8. Response includes analysis ID, status, and complete results

**Asynchronous Processing:**
1. Frontend sends URL with `async: true` flag
2. Handler creates analysis record and returns job ID immediately
3. Spawns goroutine for background processing
4. Background process follows same analysis flow
5. Frontend polls `/api/v1/analysis/{id}` for completion
6. Job status updates: pending → processing → completed/failed

## Technology Stack

### Backend
- **Language**: Go 
- **Framework**: Gin for HTTP routing
- **Database**: PostgreSQL with connection pooling
- **Cache**: Redis for result caching
- **Logging**: Zap structured logging
- **Configuration**: Viper for config management

### Frontend
- **Framework**: React 18
- **Styling**: CSS3
- **HTTP**: Fetch API for backend communication
- **Validation**: Client-side URL validation

### DevOps
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Docker Compose for local development
- **Monitoring**: Prometheus metrics collection
- **Health Checks**: Built-in endpoint monitoring

## Key Design Patterns

**Repository Pattern**
- Abstracts data access through interfaces
- PostgreSQL for analysis storage
- Redis for caching

**Middleware Chain**
- **Error Handling**: Panic recovery with structured logging
- **CORS**: Configurable origins with credential support
- **Correlation ID**: UUID-based request tracking (X-Correlation-ID header)
- **Authentication**: User identification via X-User-ID header
- **Logging**: Structured JSON logs with request/response details
- **Rate Limiting**: Sliding window algorithm (100 req/min per IP)
- **Request Size Limits**: Configurable maximum request body size

**Dependency Injection**
- Clean separation of concerns
- Testable components
- Configuration-driven setup

## Database Schema

### Analyses Table
- **Primary Key**: UUID for distributed system compatibility
- **URL Tracking**: Original URL with normalization
- **Status Management**: pending, processing, completed, failed states
- **Result Storage**: JSONB for flexible schema evolution
- **Audit Fields**: created_at, updated_at timestamps
- **User Correlation**: user_id and correlation_id for request tracking
- **Error Handling**: Dedicated error message field
- **Performance**: Indexes on URL, status, user_id, created_at

### Caching Strategy
- Redis result caching with 1-hour TTL
- Cache key format: `analysis:{url}`
- Avoids duplicate analysis work
- Cache hit/miss metrics tracked

## Performance Features

**Connection Management**
- PostgreSQL connection pooling
- HTTP client connection reuse
- Configurable timeouts and limits

**Rate Limiting**
- **Algorithm**: Sliding window with automatic cleanup
- **Limits**: 100 requests per minute per IP (configurable)
- **Storage**: In-memory visitor tracking with mutex protection
- **Cleanup**: Background goroutine removes expired entries
- **Response**: HTTP 429 with retry information
- **Monitoring**: Rate limit violations logged and tracked

**Timeout Configuration**
- HTTP request timeouts (30s default)
- Server read/write timeouts (30s default)
- Link checking timeouts (5s default)
- Analysis processing limits

## Security
- URL format validation
- Request size limits  
- CORS configuration
- Rate limiting per IP
- SQL injection prevention

## Monitoring

**Health Checks**
- Database connectivity
- Redis availability
- Service readiness

**Logging**
- Structured JSON logs
- Request correlation IDs
- Error tracking

**Metrics**
- Prometheus integration with custom metrics:
  - `http_request_duration_seconds` - Request latency histogram
  - `http_requests_total` - Total request counter by method/endpoint/status
  - `cache_hits_total` / `cache_misses_total` - Cache performance
  - `database_connections_active` - PostgreSQL connection pool
  - `redis_connections_active` - Redis connection monitoring
  - `queue_length` - Job queue metrics

## Async Processing

The current implementation uses simple Go routines for background processing:

**Implementation Details:**
1. **Job Submission**: `SubmitAnalysisJob` creates analysis record and returns job/analysis IDs
2. **Background Processing**: `ProcessAnalysisAsync` spawns goroutine with 5-minute timeout
3. **Context Management**: Proper context propagation with correlation/user IDs
4. **Cache Integration**: Checks Redis cache before performing analysis
5. **Status Updates**: Database updates with completion status and results
6. **Error Handling**: Failed jobs marked with error messages

**Frontend Integration:**
- Job submission returns immediately with pending status
- Polling mechanism checks job status every few seconds
- Real-time UI updates show job progress
- Completed jobs display full analysis results

**Current Limitations:**
- **Persistence**: Jobs lost on server restart (no persistent queue)
- **Scalability**: Single-instance only, no distributed processing
- **Priority**: No job prioritization or queue management
- **Recovery**: No retry mechanism for failed jobs
- **Monitoring**: Limited job lifecycle tracking

This approach works well for the current scope but could be enhanced with persistent job queues for production use.

## Configuration

The system uses YAML configuration with environment variable overrides:

- Analysis timeouts and limits
- Database and Redis settings
- Rate limiting parameters
- Logging levels

See `config/config.yaml` for all available options.

## Deployment

**Development:**
- Docker Compose setup
- Local port mapping
- Hot reload support

**Production Ready:**
- Multi-container deployment
- Health check endpoints
- Prometheus metrics collection
- Resource limits and monitoring

This architecture provides a solid foundation for web page analysis while keeping things simple and maintainable.

## Scalability Features

### Horizontal Scaling
- Shared database and cache
- Load balancer ready
- Container orchestration support

### Async Processing
- Go routine-based jobs
- Non-blocking operations
- Queue management
- Background processing

### Caching Layers
- L1: In-memory caching
- L2: Redis distributed cache
- Cache invalidation strategies
- Performance optimization

## Future Improvements

### Async Processing Enhancements
- **Persistent Job Queue**: Replace simple goroutines with Redis-based job queue
- **Worker Pool Management**: Implement dedicated worker processes for job processing
- **Job Recovery**: Add job persistence to survive server restarts
- **Retry Logic**: Implement exponential backoff for failed jobs
- **Job Monitoring**: Add job status tracking and progress indicators

### Technical Enhancements
- GraphQL API support
- WebSocket real-time updates
- Advanced caching strategies
- Microservice decomposition
- Circuit breaker pattern for external requests

### Feature Additions
- User authentication and authorization
- Job history tracking and analytics
- Batch processing for multiple URLs
- Result export options (CSV, JSON, PDF)
- Real-time job status updates via WebSocket

### Performance Optimization
- Connection pooling improvements
- Query optimization and database indexing
- Background job optimization with priority queues
- Memory usage optimization and garbage collection tuning
- Horizontal scaling with load balancers

### Current Limitations
- **Async Processing**: Currently uses simple goroutines, not production-ready
- **Job Persistence**: Jobs are lost on server restart
- **Queue Management**: No priority handling or job distribution
- **Scalability**: Single-instance design, not horizontally scalable

This architecture provides a solid foundation for web page analysis while keeping things simple and maintainable.