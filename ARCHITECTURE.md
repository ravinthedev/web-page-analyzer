# Web Page Analyzer - Technical Architecture

This document describes the technical implementation details, design patterns, and architectural decisions for the Web Page Analyzer application.

## System Overview

The application follows Clean Architecture principles with clear separation of concerns. We use Go for the backend API and React for the frontend interface. The system processes web page analysis requests either synchronously or asynchronously using Go routines.

## Core Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND ARCHITECTURE                                        │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐  │
│  │   User Input    │    │   URL Validation│    │   API Request   │    │   Results Display│  │
│  │                 │    │                 │    │                 │    │                 │  │
│  │ • URL Form      │    │ • Format Check  │    │ • POST /analyze │    │ • Analysis Data │  │
│  │ • Async Option  │    │ • Normalization │    │ • Sync/Async    │    │ • Job Status    │  │
│  │ • Priority      │    │ • Error Display │    │ • Polling       │    │ • Links/Images  │  │
│  └─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘  │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        │ HTTP Requests
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────────┐
│                                    BACKEND (Go)                                           │
├─────────────────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                        PRESENTATION LAYER                                           │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │    │
│  │  │   Gin Router    │  │   Handlers      │  │   Middleware    │  │     Routes      │ │    │
│  │  │                 │  │                 │  │                 │  │                 │ │    │
│  │  │ • HTTP Server   │  │ • API Endpoints │  │ • Rate Limiting │  │ • Route Config  │ │    │
│  │  │ • Request Flow  │  │ • Request Proc  │  │ • CORS          │  │ • Middleware    │ │    │
│  │  │ • Response      │  │ • Response Gen  │  │ • Logging       │  │ • Error Handling│ │    │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────────────────────┘    │
│                                        │                                                    │
│                                        │ Business Logic                                     │
│                                        ▼                                                    │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                        APPLICATION LAYER                                             │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │    │
│  │  │   Use Cases     │  │   Validation    │  │   Business Logic│  │   Orchestration │ │    │
│  │  │                 │  │                 │  │                 │  │                 │ │    │
│  │  │ • AnalyzeURL    │  │ • Input Valid   │  │ • URL Processing│  │ • Workflow Mgmt │ │    │
│  │  │ • ProcessAsync  │  │ • Business Rules│  │ • HTML Parsing  │  │ • Error Handling│ │    │
│  │  │ • GetAnalysis   │  │ • Constraints   │  │ • Data Analysis │  │ • Async Processing│ │    │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────────────────────┘    │
│                                        │                                                    │
│                                        │ Domain Logic                                       │
│                                        ▼                                                    │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                          DOMAIN LAYER                                                │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │    │
│  │  │   Entities      │  │   Services      │  │   Repositories  │  │   Interfaces    │ │    │
│  │  │                 │  │                 │  │                 │  │                 │ │    │
│  │  │ • Analysis      │  │ • Analyzer      │  │ • Analysis Repo │  │ • Repository    │ │    │
│  │  │ • AnalysisResult│  │ • HTML Parser   │  │ • Cache Repo    │  │ • Service       │ │    │
│  │  │ • LinkAnalysis  │  │ • HTTP Client   │  │ • Queue Repo    │  │ • Use Case      │ │    │
│  │  │ • AnalysisJob   │  │ • URL Validator │  │ • Job Queue    │  │ • Handler       │ │    │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────────────────────┘    │
│                                        │                                                    │
│                                        │ Data Access                                        │
│                                        ▼                                                    │
│  ┌─────────────────────────────────────────────────────────────────────────────────────┐    │
│  │                      INFRASTRUCTURE LAYER                                            │    │
│  │  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐ │    │
│  │  │   PostgreSQL    │  │      Redis      │  │   HTTP Client   │  │   Monitoring    │ │    │
│  │  │                 │  │                 │  │                 │  │                 │ │    │
│  │  │ • Analysis Data │  │ • Result Cache  │  │ • Circuit Breaker│  │ • Prometheus    │ │    │
│  │  │ • Job Queue     │  │ • Job Queue     │  │ • Timeout Mgmt  │  │ • Metrics       │ │    │
│  │  │ • UUID Storage  │  │ • Priority Queue│  │ • HTTP Client   │  │ • Health Checks │ │    │
│  │  └─────────────────┘  └─────────────────┘  └─────────────────┘  └─────────────────┘ │    │
│  └─────────────────────────────────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────────────────────────────────┘
```

### Architecture Flow:
1. **Frontend** → Sends HTTP requests to Backend API
2. **Presentation Layer** → Handles HTTP requests, applies middleware
3. **Application Layer** → Orchestrates business logic and workflows
4. **Domain Layer** → Contains core business entities and rules
5. **Infrastructure Layer** → Manages data persistence and external services

### Key Components:

#### **Frontend Flow:**
- **User Input**: URL form with async option and priority setting
- **URL Validation**: Client-side format checking and normalization
- **API Request**: POST to /api/v1/analyze with sync/async modes
- **Results Display**: Analysis data, job status, and metadata

#### **Backend Layers:**
- **Domain Layer**: Core business entities and interfaces
- **Application Layer**: Business logic, validation, and use case orchestration
- **Infrastructure Layer**: Data persistence and external service integration
- **Presentation Layer**: HTTP handling and API endpoints

#### **Data Flow:**
- **Synchronous**: Direct analysis with immediate response
- **Asynchronous**: Background processing with job tracking and polling
- **Caching**: Redis-based result caching for performance
- **Queue Management**: Redis-based priority job queue for async processing

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Framework**: Gin for HTTP routing
- **Database**: PostgreSQL with connection pooling
- **Cache**: Redis for result caching
- **Logging**: Zap structured logging
- **Configuration**: Viper for config management

### Frontend
- **Framework**: React 18
- **Styling**: CSS3 with responsive design
- **HTTP**: Fetch API for backend communication
- **Validation**: Client-side URL validation

### DevOps
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Docker Compose for local development
- **Monitoring**: Prometheus metrics collection
- **Health Checks**: Built-in endpoint monitoring

## Design Patterns Implemented

### Repository Pattern
- Abstract data access through interfaces
- PostgreSQL implementation for analysis storage
- Redis implementation for caching and queues

### Circuit Breaker Pattern
- HTTP client with failure detection
- Automatic recovery after timeout periods
- Prevents cascading failures

### Middleware Chain
- Request processing pipeline
- Rate limiting per IP address
- Structured logging with correlation IDs
- CORS and security headers

### Strategy Pattern
- Different analysis modes (sync/async)
- Configurable timeout and retry strategies
- Pluggable HTML parsing implementations

## Data Flow

### Synchronous Analysis
1. Client submits URL via POST /api/v1/analyze
2. Backend validates URL and checks cache
3. If not cached, fetches webpage content
4. Parses HTML and extracts metadata
5. Stores result in database and cache
6. Returns analysis result to client

### Asynchronous Analysis
1. Client submits URL with async flag
2. Backend creates analysis record and job
3. Starts Go routine for background processing
4. Returns job ID for status tracking
5. Background process updates analysis status
6. Client polls for completion status

## Database Schema

### Analyses Table
- UUID primary key
- URL and status tracking
- JSONB result storage for flexibility
- Timestamps and user correlation
- Retry count and priority fields

### Caching Strategy
- Redis-based result caching
- Configurable TTL (default: 1 hour)
- Cache key format: `analysis:{url}`
- Automatic cache invalidation

## Performance Considerations

### Connection Pooling
- PostgreSQL connection limits
- Redis connection pooling
- HTTP client connection reuse
- Graceful connection cleanup

### Rate Limiting
- Per-IP request limiting
- Configurable window and count
- Sliding window implementation
- Automatic cleanup of expired entries

### Timeout Handling
- Request timeout middleware
- Configurable analysis timeouts
- Graceful degradation on failures
- Circuit breaker for external requests

## Security Implementation

### Input Validation
- URL format validation (scheme, host checking)
- Content length limits with MaxBytesReader
- Basic request size validation
- SQL injection prevention through parameterized queries

### Authentication
- User ID header support (X-User-ID)
- Anonymous user handling
- Correlation ID tracking
- Request logging and monitoring

### CORS Configuration
- Frontend origin allowance
- Method and header restrictions
- Preflight request handling
- Basic security headers

### Security Gaps & Improvements
- **Could Add**: XSS protection and HTML sanitization
- **Could Add**: Advanced security headers (CSP, HSTS, etc.)
- **Could Add**: Input content sanitization beyond URL format
- **Could Add**: Rate limiting per user (currently only per IP)
- **Could Add**: Request signature validation
- **Could Add**: API key authentication

## Monitoring and Observability

### Prometheus Metrics
- HTTP request duration and count
- Analysis job metrics
- Cache hit/miss rates
- Database connection status

### Structured Logging
- JSON format for parsing
- Correlation ID tracking
- User and request context
- Error stack traces

### Health Checks
- Database connectivity
- Redis availability
- Service readiness
- External dependency status

## Error Handling Strategy

### Graceful Degradation
- Partial result returns
- Fallback error messages
- Timeout handling
- Circuit breaker implementation

### Error Propagation
- Structured error types
- Context preservation
- Stack trace logging
- Client-friendly messages

## Testing Strategy

### Unit Tests
- Mock implementations
- No real HTTP requests
- Fast execution times

### Integration Tests
- Database connectivity
- Redis operations
- API endpoint validation
- End-to-end workflows

### Test Coverage
- Domain entities
- Application layer: Basic structure
- Presentation layer: HTTP handling
- Middleware: Core functionality

## Deployment Architecture

### Development Environment
- Docker Compose for services
- Local port mapping
- Hot reload for development
- Environment-specific configs

### Production Considerations
- Multi-container deployment
- Health check monitoring
- Resource limits and scaling
- Backup and recovery

## Scalability Features

### Horizontal Scaling
- Stateless API design
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

### Technical Enhancements
- GraphQL API support
- WebSocket real-time updates
- Advanced caching strategies
- Microservice decomposition

### Feature Additions
- User authentication
- Job history tracking
- Batch processing
- Result export options

### Performance Optimization
- Connection pooling improvements
- Query optimization
- Background job optimization
- Memory usage optimization

## Configuration Management

### Environment Variables
- Database connection strings
- Redis configuration
- Logging levels
- Feature flags

### Configuration Files
- YAML-based configuration
- Environment-specific overrides
- Default value management
- Validation and error handling

## Development Workflow

### Code Organization
- Interface definitions
- Dependency management
- Testing strategy

### Build Process
- Multi-stage Docker builds
- Go module management
- Frontend build optimization
- Asset compilation

### Deployment Pipeline
- Docker image building
- Service orchestration
- Health check validation
- Rollback procedures

This architecture provides Architecture foundation for a scalable web page analysis system while maintaining simplicity and maintainability.