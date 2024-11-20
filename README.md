# URL Shortener API

A high-performance URL shortening service built with Go, Gin, and MySQL.

## Features

- URL shortening with random slug generation
- Custom slug support
- Configurable expiration dates
- Rate limiting to prevent abuse
- MySQL persistence with GORM
- RESTful API design
- Comprehensive test coverage

## Tech Stack

- Go 1.23+
- Gin Web Framework
- GORM ORM
- MySQL 8.0+
- Environment-based configuration

## Prerequisites

- Go 1.23 or higher
- MySQL 8.0 or higher
- Make (optional, for using Makefile)

## Installation

1. Clone the repository:
```bash
git clone ssh://git@github.com:MunishMummadi/url-shortener.git
cd url-shortener
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your MySQL credentials
```

3. Install dependencies:
```bash
go mod download
```

4. Run the application:
```bash
go run main.go
```

## API Endpoints

### Create Short URL
```bash
POST /generate/shortlink
Content-Type: application/json

{
    "url": "https://www.google.com",
    "customSlug": "google",    # optional
    "expirationDate": "2024-12-31"    # optional
}
```

### Access Short URL
```bash
GET /{shortLink}
```

### Delete URL
```bash
DELETE /{shortLink}
```

## Testing

Run all tests:
```bash
go test -v ./...
```

Run with coverage:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Performance

- Handles 1000 requests/second on standard hardware
- Average response time < 50ms
- Rate limited to 40 requests/minute per IP

## Security

- Input validation for all endpoints
- Rate limiting per IP address
- SQL injection prevention with GORM
- Custom slug validation
- Expiration date validation

