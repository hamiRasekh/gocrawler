# 🕷️ Professional Go Web Crawler

A powerful and professional web crawling system built with Go, designed for scraping embroidery design data from websites. This project uses Clean Architecture and provides advanced features such as headless browser support, rotating proxy management, and worker pool processing.

## 📋 Table of Contents

- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Configuration](#configuration)
- [Security](#security)
- [API Documentation](#api-documentation)
- [Usage Examples](#usage-examples)
- [Development](#development)
- [Contributing](#contributing)

## ✨ Features

### Core Capabilities

- ✅ **Headless Browser**: Uses Chromedp for crawling JavaScript-rendered web pages
- ✅ **Rotating Proxy**: Automatic proxy management and rotation with health checking
- ✅ **Browser Fingerprinting**: Simulates real browsers with advanced fingerprinting
- ✅ **REST API**: Complete API for managing tasks, proxies, and results
- ✅ **PostgreSQL**: Data storage in PostgreSQL with automatic migrations
- ✅ **Docker Support**: Full Docker and Docker Compose support
- ✅ **Rate Limiting**: Request rate control to prevent blocking
- ✅ **Retry Mechanism**: Automatic retry with exponential backoff
- ✅ **Worker Pool**: Concurrent processing with worker pool for high performance
- ✅ **Embroidery API Crawler**: Specialized crawler for embroiderydesigns.com API
- ✅ **WebSocket Support**: Real-time log streaming via WebSocket
- ✅ **JWT Authentication**: Secure token-based authentication
- ✅ **Rate Limiting**: Protection against brute force attacks

### Advanced Features

- 🔄 **Auto-retry**: Automatic retry on errors
- 🎭 **Stealth Mode**: Advanced techniques to hide crawler identity
- 📊 **Real-time Monitoring**: Monitor task and proxy status
- 🗄️ **Product Management**: Store and manage scraped products
- 🔍 **Advanced Filtering**: Advanced product filtering by brand, catalog, etc.
- 🔐 **Security Hardened**: No hardcoded credentials, configurable CORS, rate limiting

## 🏗️ Architecture

The project is designed with Clean Architecture and Dependency Injection:

```
gocrawler/
├── cmd/
│   └── crawler/              # Application entry point
│       └── main.go           # Main application file
│
├── internal/
│   ├── api/                  # API layer and HTTP handlers
│   │   ├── handlers.go       # REST API handlers
│   │   ├── middleware.go     # Middleware (CORS, Logger, Recovery, Rate Limiting)
│   │   ├── routes.go         # Route definitions
│   │   └── server.go         # HTTP server
│   │
│   ├── auth/                 # Authentication
│   │   ├── jwt.go           # JWT token generation and validation
│   │   └── refresh.go       # Refresh token management
│   │
│   ├── browser/              # Headless browser management
│   │   ├── fingerprint.go   # Browser fingerprint generation
│   │   ├── launcher.go      # Browser launcher
│   │   ├── manager.go       # Browser manager
│   │   └── stealth.go       # Stealth techniques
│   │
│   ├── config/               # Configuration management
│   │   └── config.go        # Configuration loading and management
│   │
│   ├── crawler/              # Crawler core
│   │   ├── api_crawler.go          # REST API crawler
│   │   ├── browser_crawler.go      # Browser crawler
│   │   ├── embroidery_api_crawler.go  # Specialized embroidery crawler
│   │   ├── web_crawler.go           # General web crawler
│   │   ├── worker_pool.go          # Worker pool for parallel processing
│   │   └── interface.go            # Crawler interfaces
│   │
│   ├── fingerprint/          # Browser fingerprinting
│   │   ├── headers.go       # Realistic header generation
│   │   └── profile.go       # Browser profile
│   │
│   ├── proxy/                # Proxy management
│   │   ├── health_checker.go # Proxy health checking
│   │   ├── manager.go       # Proxy manager
│   │   └── pool.go          # Proxy pool management
│   │
│   ├── service/              # Business logic layer
│   │   ├── crawler_service.go  # Crawler service
│   │   └── task_service.go     # Task service
│   │
│   ├── storage/              # Database layer
│   │   ├── models.go        # Data models
│   │   ├── postgres.go      # PostgreSQL connection
│   │   └── repository.go    # Repository pattern
│   │
│   ├── utils/                # Utility functions
│   │   ├── logger.go        # Logger
│   │   ├── rate_limiter.go  # Rate limiter
│   │   └── retry.go         # Retry mechanism
│   │
│   └── websocket/            # WebSocket support
│       ├── handler.go       # WebSocket handler
│       └── hub.go           # WebSocket hub
│
├── migrations/               # Database migrations
│   ├── 001_initial.up.sql
│   ├── 001_initial.down.sql
│   └── ...
│
├── docker/                   # Docker files
│   ├── Dockerfile           # Main Dockerfile
│   ├── Dockerfile.arvan     # Arvan Cloud Dockerfile
│   ├── entrypoint.sh        # Startup script
│   └── daemon.json.example  # Docker daemon settings
│
├── frontend/                 # React frontend
│   └── ...
│
├── docker-compose.yml        # Docker Compose configuration
├── go.mod                    # Go modules
├── Makefile                  # Make commands
└── README.md                 # This file
```

## 🚀 Installation

### Prerequisites

- Go 1.21 or higher
- Docker & Docker Compose
- PostgreSQL 15+ (or use Docker Compose)

### Method 1: Docker Compose (Recommended)

1. **Clone the repository:**
```bash
git clone <repository-url>
cd gocrawler
```

2. **Create environment file:**
```bash
cp .env.example .env
# Edit .env and set required values (see Configuration section)
```

3. **Create Docker network (if it doesn't exist):**
```bash
docker network create production_network
```

4. **Start backend services:**
```bash
docker compose up -d
```

5. **Start frontend (optional):**
```bash
docker compose -f docker-compose.frontend.yml up -d
```

6. **Check status:**
```bash
curl http://localhost:8009/api/v1/health
```

### Method 2: Local Development

1. **Install dependencies:**
```bash
go mod download
```

2. **Set up PostgreSQL:**
```bash
# Create database
psql -U postgres -c "CREATE DATABASE crawler_db;"
```

3. **Run migrations:**
```bash
# Install golang-migrate first
migrate -path migrations -database "postgres://crawler:password@localhost:5432/crawler_db?sslmode=disable" up
```

4. **Set environment variables:**
```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=crawler
export DB_PASSWORD=your_password
export DB_NAME=crawler_db
export JWT_SECRET=your_jwt_secret
# ... (see Configuration section for all variables)
```

5. **Run the application:**
```bash
go run cmd/crawler/main.go
```

## ⚙️ Configuration

### Environment Variables

Create a `.env` file in the project root (use `.env.example` as a template):

#### Required Variables

```bash
# Database (REQUIRED)
DB_PASSWORD=your_secure_password_here
POSTGRES_PASSWORD=your_secure_password_here

# Authentication (REQUIRED)
JWT_SECRET=your_jwt_secret_key_here_minimum_32_characters
```

#### Server Configuration

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8009
API_PREFIX=/api/v1
CORS_ORIGIN=*  # Set to your frontend domain in production
```

#### Database Configuration

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=crawler
DB_PASSWORD=your_secure_password_here
DB_NAME=crawler_db
DB_SSLMODE=disable  # Use 'require' in production
```

#### Authentication & Security

```bash
JWT_SECRET=your_jwt_secret_key_here_minimum_32_characters
JWT_EXPIRATION=24h
REFRESH_TOKEN_EXPIRATION=168h
ADMIN_TOKEN_LIFETIME=8760h

# Rate limiting for authentication endpoints
AUTH_RATE_LIMIT_REQUESTS=5
AUTH_RATE_LIMIT_WINDOW=15m
```

#### Crawler Configuration

```bash
MAX_WORKERS=10
RATE_LIMIT_PER_SECOND=5
REQUEST_TIMEOUT=30s
RETRY_MAX_ATTEMPTS=3
RETRY_BACKOFF_MULTIPLIER=2
```

#### Browser Configuration

```bash
HEADLESS=true
BROWSER_TIMEOUT=60s
USER_DATA_DIR=/tmp/browser-data
```

#### Proxy Configuration

```bash
PROXY_ENABLED=true
PROXY_HEALTH_CHECK_INTERVAL=5m
PROXY_MAX_FAILURES=3
```

#### Embroidery API Configuration (Optional)

```bash
EMBROIDERY_BASE_URL=https://www.embroiderydesigns.com/es/prdsrch
EMBROIDERY_AUTH_TOKEN=
EMBROIDERY_COOKIES=
EMBROIDERY_PAGE_SIZE=120
EMBROIDERY_CHECK_INTERVAL=6h
```

### Generating Secrets

Generate a secure JWT secret:
```bash
openssl rand -base64 32
```

## 🔐 Security

### Security Features

- ✅ **No Hardcoded Credentials**: All secrets must be provided via environment variables
- ✅ **Configurable CORS**: Set `CORS_ORIGIN` to your frontend domain (not `*` in production)
- ✅ **Rate Limiting**: Authentication endpoints are rate-limited to prevent brute force attacks
- ✅ **JWT Authentication**: Secure token-based authentication with refresh tokens
- ✅ **WebSocket Authentication**: WebSocket connections require valid JWT tokens
- ✅ **Input Validation**: All inputs are validated to prevent injection attacks
- ✅ **Error Handling**: Error messages don't leak sensitive information
- ✅ **SQL Injection Protection**: All queries use parameterized statements

### Security Best Practices

1. **Never commit `.env` files** to version control
2. **Use strong, unique passwords** for production
3. **Generate JWT_SECRET** using: `openssl rand -base64 32`
4. **Set CORS_ORIGIN** to your frontend domain in production (not `*`)
5. **Use SSL/TLS** in production (set `DB_SSLMODE=require`)
6. **Regularly rotate** secrets and passwords
7. **Keep dependencies updated** for security patches

### WebSocket Security

WebSocket connections require authentication via:
- Query parameter: `ws://host/ws/logs?token=YOUR_JWT_TOKEN`
- Authorization header: `Authorization: Bearer YOUR_JWT_TOKEN`

## 📡 API Documentation

### Base URL

```
http://localhost:8009/api/v1
```

### Swagger UI

Interactive API documentation is available at:
```
http://localhost:8009/swagger
```

### Authentication

Most endpoints require JWT authentication. Include the token in the Authorization header:

```
Authorization: Bearer YOUR_JWT_TOKEN
```

### Authentication Endpoints

#### Register Admin

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "admin",
  "password": "SecurePassword123"
}
```

#### Login

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "SecurePassword123"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "refresh_token_string",
  "expires_in": 86400,
  "refresh_expires_at": "2024-01-08T12:00:00Z",
  "user": {
    "id": 1,
    "username": "admin"
  }
}
```

#### Refresh Token

```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your_refresh_token"
}
```

#### Generate Admin API Token

```http
POST /api/v1/auth/admin-token
Content-Type: application/json

{
  "username": "admin",
  "password": "SecurePassword123",
  "token_name": "my-api-token"
}
```

### Task Management

#### Create Task

```http
POST /api/v1/tasks
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "name": "Example Task",
  "url": "https://example.com/api/data",
  "type": "api",
  "config": {
    "crawler_type": "embroidery_api"
  }
}
```

#### List Tasks

```http
GET /api/v1/tasks?limit=10&offset=0
Authorization: Bearer YOUR_TOKEN
```

#### Start Task

```http
POST /api/v1/tasks/{id}/start
Authorization: Bearer YOUR_TOKEN
```

#### Get Task Status

```http
GET /api/v1/tasks/{id}/status
Authorization: Bearer YOUR_TOKEN
```

### Product Management

#### List Products

```http
GET /api/v1/products?limit=20&offset=0&brand=ABC&in_stock=true
Authorization: Bearer YOUR_TOKEN
```

#### Get Product

```http
GET /api/v1/products/{id}
Authorization: Bearer YOUR_TOKEN
```

#### Start Embroidery Crawl

```http
POST /api/v1/products/crawl
Authorization: Bearer YOUR_TOKEN
```

### Proxy Management

#### List Proxies

```http
GET /api/v1/proxies
Authorization: Bearer YOUR_TOKEN
```

#### Add Proxy

```http
POST /api/v1/proxies
Authorization: Bearer YOUR_TOKEN
Content-Type: application/json

{
  "host": "proxy.example.com",
  "port": 8080,
  "type": "http",
  "username": "user",
  "password": "pass"
}
```

## 💡 Usage Examples

### Example 1: Create and Start a Crawling Task

```bash
# Create task
curl -X POST http://localhost:8009/api/v1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Embroidery Products Crawl",
    "url": "https://www.embroiderydesigns.com/es/prdsrch",
    "type": "api",
    "config": {
      "crawler_type": "embroidery_api"
    }
  }'

# Start task (replace {task_id} with actual ID)
curl -X POST http://localhost:8009/api/v1/tasks/{task_id}/start \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Example 2: Quick Embroidery Crawl

```bash
curl -X POST http://localhost:8009/api/v1/products/crawl \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Example 3: Connect to WebSocket for Real-time Logs

```javascript
const token = 'YOUR_JWT_TOKEN';
const ws = new WebSocket(`ws://localhost:8009/ws/logs?token=${token}&task_id=1`);

ws.onmessage = (event) => {
  const log = JSON.parse(event.data);
  console.log(`[${log.level}] ${log.message}`);
};
```

## 🔧 Development

### Building

```bash
go build -o crawler ./cmd/crawler
```

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

### Linting

```bash
golangci-lint run
```

### Building Docker Image

```bash
docker build -f docker/Dockerfile -t crawler:latest .
```

## 📊 Database Schema

### Main Tables

- **tasks**: Crawling tasks
- **crawl_results**: Crawling results
- **proxies**: Proxy configurations
- **products**: Scraped products
- **users**: User accounts
- **api_tokens**: API tokens
- **refresh_tokens**: Refresh tokens
- **crawler_settings**: Application settings

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Code Style

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for exported functions
- Write tests for new features
- Ensure all tests pass before submitting

## 📝 License

MIT License - see LICENSE file for details

## 🐛 Troubleshooting

### Common Issues

1. **Database Connection Error**
   - Verify PostgreSQL is running
   - Check database credentials in `.env`
   - Ensure database exists

2. **Proxy Errors**
   - Verify proxy configurations
   - Check proxy health status
   - Ensure proxy credentials are correct

3. **Browser Errors**
   - Verify Chromium is installed
   - Check browser path configuration
   - Ensure sufficient system resources

4. **Authentication Errors**
   - Verify JWT_SECRET is set
   - Check token expiration
   - Ensure token is included in requests

For more troubleshooting help, see `TROUBLESHOOTING.md`.

## 📞 Support

For questions and issues, please open an issue in the repository.

---

**Note**: This project is designed for educational and research purposes. Please respect the terms of service and robots.txt of target websites.
