# Web Page Analyzer

A simple web application that analyzes web pages to extract useful information like HTML version, headings, links, and login forms. Built with Go backend and React frontend.

## What It Does

- Analyzes web pages for HTML structure and content
- Detects internal/external links and their accessibility
- Identifies login forms and page metadata
- Supports both synchronous and asynchronous processing
- Provides clean API endpoints for integration

## Prerequisites

- Go 1.21+
- Node.js 16+
- Docker and Docker Compose
- PostgreSQL 15+
- Redis 7+

## Quick Start

1. Clone the repository
2. Start the services:
   ```bash
   docker-compose up -d
   ```
3. Access the application:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Health check: http://localhost:8080/health

## Project Structure

- `cmd/api/` - Main Go API server
- `internal/` - Core application logic
- `frontend/` - React web interface
- `config/` - Configuration files
- `migrations/` - Database schema

## Main Features

- **URL Analysis**: Submit URLs for detailed webpage analysis
- **Async Processing**: Handle multiple URLs in background
- **Caching**: Redis-based result caching for performance
- **Rate Limiting**: Built-in protection against abuse
- **Health Monitoring**: Prometheus metrics and health checks

## External Dependencies

- PostgreSQL for data storage
- Redis for caching and job queues
- Prometheus for metrics collection

## Setup Instructions

1. **Database Setup**:
   - PostgreSQL runs on port 5432
   - Redis runs on port 6379
   - Migrations run automatically on startup

2. **Backend Setup**:
   - Configuration in `config/config.yaml`
   - Environment variables override config
   - Logs in structured JSON format

3. **Frontend Setup**:
   - React app with environment config
   - URL validation and error handling
   - Responsive design for mobile/desktop

## Usage

1. Open the web interface
2. Enter a URL to analyze
3. Choose sync or async processing
4. View results with detailed breakdown
5. Track job status for async requests

## Key Technologies

- **Backend**: Go, Gin, PostgreSQL, Redis
- **Frontend**: React, CSS3, Fetch API
- **DevOps**: Docker, Prometheus, health checks
- **Validation**: Client-side URL validation, server-side sanitization

## Challenges & Solutions

- **Async Processing**: Replaced complex worker services with Go routines for simplicity
- **Frontend Caching**: Implemented proper build process and cache busting
- **Database Schema**: Used JSONB for flexible result storage
- **Error Handling**: Comprehensive error messages and logging

## Improvements

- Add user authentication and job history
- Implement result export (CSV, JSON)
- Add batch URL processing
- Enhance link accessibility checking
- Add more HTML analysis features

## Architecture Details

For detailed technical architecture, design patterns, and implementation details, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Development

- Run tests: `./test-coverage.sh`
- Build: `go build ./cmd/api`
- Frontend dev: `cd frontend && npm start`

## API Documentation

- Base URL: `http://localhost:8080`
- Version: `/api/v1`
- Endpoints: `/analyze`, `/analysis/:id`, `/analyses`
- Health: `/health`, `/metrics`

## License

MIT License - feel free to use and modify.
