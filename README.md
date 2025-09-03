# Web Page Analyzer

A simple web application that analyzes web pages to extract useful information like HTML version, headings, links, and login forms. Built with Go backend and React frontend.

## What It Does

- Analyzes web pages for HTML structure and content
- Detects internal/external links and their accessibility
- Identifies login forms and page metadata
- Supports both synchronous and asynchronous processing
- Provides clean API endpoints for integration

## Prerequisites

- Docker and Docker Compose

## Quick Start

1. Clone the repository
2. Start the services:
   ```bash
   docker compose build
   ``` 
   ```bash
   docker compose up -d
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
   - Clean, modern interface

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
- **Validation**: Client-side URL validation and normalization

## Challenges & Solutions

- **Async Processing**: Replaced complex worker services with Go routines for simplicity
- **Frontend Caching**: Implemented proper build process with Docker
- **Database Schema**: Used JSONB for flexible result storage
- **Error Handling**: Comprehensive error messages and logging

## Future Improvements

- Add user authentication and job history
- Implement result export (CSV, JSON)
- Add batch URL processing
- Enhance link accessibility checking
- Add more HTML analysis features

## Architecture Details

For detailed technical architecture, design patterns, and implementation details, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## Configuration

The application can be configured through `config/config.yaml` and environment variables. Key configuration parameters include:

### Analysis Settings
- `analysis.request_timeout` - HTTP request timeout for web page fetching (default: 30s)
- `analysis.max_content_length` - Maximum HTML content size to process (default: 10MB)
- `analysis.cache_ttl` - Cache time-to-live for analysis results (default: 1h)
- `analysis.link_check_timeout` - Timeout for checking link accessibility (default: 5s)
- `analysis.max_links_to_check` - Maximum number of links to check per page (default: 50)
- `analysis.max_concurrent_link_checks` - Concurrent link checks limit (default: 10)
- `analysis.max_html_depth` - Maximum HTML parsing depth (default: 100)
- `analysis.max_url_length` - Maximum URL length allowed (default: 2048)

### Rate Limiting
- `analysis.rate_limit_per_ip` - Requests per IP per window (default: 100)
- `analysis.rate_limit_window` - Rate limiting time window (default: 1m)

### Database & Cache
- `database.*` - PostgreSQL connection settings
- `redis.*` - Redis connection and cache settings

### Server Settings
- `server.port` - HTTP server port (default: 8080)
- `server.read_timeout` - Server read timeout (default: 30s)
- `server.write_timeout` - Server write timeout (default: 30s)

Environment variables can override any config value using the format: `SECTION_KEY` (e.g., `ANALYSIS_REQUEST_TIMEOUT=45s`).

## Development

- Run tests: `go test ./...`
- Run tests with coverage: `go test -cover ./...`
- Build: `go build ./cmd/api`
- Docker build: `docker compose build`

## API Documentation

- Base URL: `http://localhost:8080`
- Version: `/api/v1`
- Endpoints: `/analyze`, `/analysis/:id`, `/analyses`
- Health: `/health`, `/metrics`

## License

MIT License - feel free to use and modify.
