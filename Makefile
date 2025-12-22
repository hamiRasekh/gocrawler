.PHONY: build run test clean docker-build docker-up docker-down migrate-up migrate-down

# Build the application
build:
	go build -o bin/crawler ./cmd/crawler

# Run the application
run:
	go run ./cmd/crawler

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f crawler

# Docker commands
docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f crawler

# Database migrations (requires golang-migrate)
migrate-up:
	migrate -path migrations -database "postgres://crawler:crawler_password@localhost:5432/crawler_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://crawler:crawler_password@localhost:5432/crawler_db?sslmode=disable" down

# Install dependencies
deps:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

