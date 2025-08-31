# Web Page Analyzer

A Go application for analyzing web pages and extracting useful information.

## What We Added

Complete Backend Implementation
- Clean Architecture with Domain, Application, Infrastructure, and Presentation layers
- PostgreSQL database with full CRUD operations
- Redis caching and job queue management
- Circuit breaker HTTP client for external requests
- Prometheus monitoring and metrics collection
- Structured logging with Zap
- Rate limiting and middleware chain
- Configuration management with Viper

Domain Layer
- Analysis entities with full status tracking
- Repository interfaces for data access
- Service interfaces for business logic

Infrastructure Layer
- PostgreSQL repository implementation
- Redis cache and job queue repositories
- Circuit breaker HTTP client
- Monitoring metrics collection

Application Layer
- Use cases with business logic
- DTOs for API communication
- Job processing and retry logic

Presentation Layer
- REST API endpoints
- Middleware for logging, rate limiting, CORS
- Error handling and validation
- Request/response processing

## Project Structure

```
web-page-analyzer/
├── cmd/api/                    # API server entry point
├── config/                     # Configuration files
├── internal/                   # Internal application code
│   ├── domain/                # Domain layer
│   ├── application/           # Application layer
│   ├── infrastructure/        # Infrastructure layer
│   └── presentation/          # Presentation layer
├── migrations/                # Database migrations
├── monitoring/                # Prometheus configuration
├── pkg/                       # Shared packages
│   ├── config/               # Configuration management
│   ├── logger/               # Structured logging
│   └── errors/               # Error handling
├── docker-compose.yml         # Development environment
└── Dockerfile                 # Multi-stage build
```

## Running

```bash
go mod tidy
go run cmd/api/main.go
```

## API Endpoints

- GET / - API information
- GET /health - Health check  
- POST /api/v1/analyze - Analyze a web page
- GET /api/v1/analyze/:id - Get analysis results by ID

## Features

- Web page analysis with HTML parsing
- Database storage with PostgreSQL
- Redis caching for performance
- Job queue for async processing
- Circuit breaker for external requests
- Rate limiting and security
- Monitoring and metrics
- Structured logging
- Configuration management
