FROM docker.arvancloud.ir/golang:1.21 AS builder

WORKDIR /app

RUN apt-get update && apt-get install -y git gcc && rm -rf /var/lib/apt/lists/*

# Copy go mod files first
COPY go.mod ./

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org
ENV GOPRIVATE=

RUN go mod download || (GOPROXY=https://goproxy.io,direct go mod download) || (GOPROXY=direct go mod download)
RUN go mod verify

# Copy source code
COPY . .

# Generate go.sum if missing
RUN go mod tidy || true

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crawler ./cmd/crawler

FROM docker.arvancloud.ir/debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y \
    ca-certificates \
    tzdata \
    chromium \
    chromium-driver \
    postgresql-client \
    && rm -rf /var/lib/apt/lists/*

ENV CHROMIUM_PATH=/usr/bin/chromium
ENV CHROME_PATH=/usr/bin/chromium

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/crawler .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Copy documentation assets (Swagger UI, examples, etc.)
COPY --from=builder /app/docs ./docs

# Copy entrypoint script
COPY docker/entrypoint.sh /root/entrypoint.sh
RUN chmod +x /root/entrypoint.sh

# Expose port
EXPOSE 8009

# Use entrypoint script to run migrations and start app
ENTRYPOINT ["/root/entrypoint.sh"]

