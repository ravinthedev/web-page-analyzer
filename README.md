# Web Page Analyzer

A Go application for analyzing web pages and extracting useful information.

## What We Added

Clean Architecture Structure
- Domain layer with entities, repositories, and services interfaces
- Application layer with use cases and DTOs  
- Presentation layer with handlers and routes
- Infrastructure layer structure ready for implementation

Basic Entities
- Analysis entity with status tracking
- AnalysisResult with extracted page information
- LinkAnalysis and ImageAnalysis for page elements

Repository Interface
- AnalysisRepository interface for data operations

Service Interfaces  
- AnalyzerService for analysis operations
- HTMLParser for parsing HTML content
- HTTPClient for fetching web pages

Use Cases
- AnalysisUseCase with basic business logic

DTOs
- Request and response data transfer objects

HTTP Handlers
- REST API endpoints for analysis
- Health check endpoint

Routes
- API routing configuration

Docker Configuration
- Multi-stage Dockerfile
- Docker Compose files for development and production
- Dependencies only compose file

Database Schema
- PostgreSQL migration for analyses table

Monitoring
- Prometheus configuration

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
├── docker-compose.yml         # Development environment
├── docker-compose.deps.yml    # Dependencies only
├── docker-compose.prod.yml    # Production environment
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
