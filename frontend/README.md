# Web Page Analyzer Frontend

## Features

### URL Validation
- Accepts various URL formats:
  - Full URLs: `https://example.com`, `http://github.com`
  - Domain names: `example.com`, `svgtopng.com`, `github.com`
  - Subdomains: `subdomain.example.com`
  - International domains: `test.co.uk`

### Processing Modes
- **Synchronous**: Immediate analysis results
- **Asynchronous**: Background processing with job tracking

### Configuration
Environment variables in `.env`:
- `REACT_APP_API_BASE_URL`: Backend API URL (default: http://localhost:8080)
- `REACT_APP_API_VERSION`: API version (default: v1)
- `REACT_APP_POLLING_INTERVAL`: Polling interval in ms (default: 2000)
- `REACT_APP_MAX_POLLING_ATTEMPTS`: Max polling attempts (default: 150)

## Development

### Local Development
```bash
npm install
npm start
```

### Docker Development
```bash
docker compose build frontend
docker compose up frontend
```

### Environment Setup
Copy `.env.example` to `.env` and configure:
```bash
cp .env.example .env
# Edit .env with your configuration
```

## URL Validation Examples

✅ **Valid Inputs:**
- `svgtopng.com` → Normalized to `https://svgtopng.com`
- `http://example.com` → Kept as `http://example.com`
- `https://github.com` → Kept as `https://github.com`
- `subdomain.example.com` → Normalized to `https://subdomain.example.com`

❌ **Invalid Inputs:**
- `invalid` → Error: "Please enter a valid domain name (e.g., example.com)"
- `http://` → Error: "Please enter a valid URL format"
- `domain` → Error: "Please enter a valid domain name (e.g., example.com)"
