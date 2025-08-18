# Dockerfile for Go Sarama Schema Registry Producer
# Multi-stage build for optimized final image size
# Note: Uses Debian/glibc instead of Alpine/musl for librdkafka compatibility

# Stage 1: Build stage
FROM golang:1.24 AS builder

# Install build dependencies including librdkafka-dev for CGO
RUN apt-get update && apt-get install -y \
    git \
    ca-certificates \
    gcc \
    libc6-dev \
    librdkafka-dev \
    pkg-config \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files first (for better caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build working producer binaries with CGO enabled for confluent-kafka-go
RUN CGO_ENABLED=1 GOOS=linux go build -a -o bin/continuous_producer ./cmd/continuous_producer
RUN CGO_ENABLED=1 GOOS=linux go build -a -o bin/producer ./cmd/producer

# Note: robust_producer and enhanced_producer currently have missing dependencies
# TODO: Complete the errorhandling, logging, and metrics packages to enable these builds
# RUN CGO_ENABLED=1 GOOS=linux go build -a -o bin/robust_producer ./cmd/robust_producer
# RUN CGO_ENABLED=1 GOOS=linux go build -a -o bin/enhanced_producer ./cmd/enhanced_producer

# Stage 2: Final runtime stage
FROM debian:bookworm-slim

# Install runtime dependencies including librdkafka
RUN apt-get update && apt-get install -y \
    ca-certificates \
    librdkafka1 \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN groupadd --system appgroup && useradd --system --gid appgroup appuser

# Set working directory
WORKDIR /app

# Copy built binaries from builder stage
COPY --from=builder /app/bin/ ./bin/

# Copy any necessary config files (if they exist)
COPY --from=builder /app/.env.example ./.env.example

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose any necessary ports (none needed for producers, but good practice)
# EXPOSE 8080

# Environment variables with default values
# Required at runtime (no defaults provided):
# - KAFKA_BOOTSTRAP_SERVERS
# - KAFKA_API_KEY  
# - KAFKA_API_SECRET
# - KAFKA_TOPIC
# - SCHEMA_REGISTRY_URL
# - SCHEMA_REGISTRY_API_KEY
# - SCHEMA_REGISTRY_API_SECRET

# Optional environment variables with defaults:
ENV LOG_LEVEL=info
ENV MESSAGE_INTERVAL=1s
ENV KAFKA_TIMEOUT=30s
ENV KAFKA_RETRY_MAX=3
ENV KAFKA_REQUIRED_ACKS=1
ENV KAFKA_COMPRESSION=snappy
ENV KAFKA_BATCH_SIZE=100
ENV ENABLE_METRICS=false

# Health check (optional, for monitoring)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep -f "continuous_producer|producer" || exit 1

# Default command runs the continuous producer
# Can be overridden at runtime with docker run
CMD ["./bin/continuous_producer"]

# Usage examples:
# Build: docker build -t go-sarama-producer .
# Run continuous producer: docker run --env-file .env go-sarama-producer
# Run basic producer: docker run --env-file .env go-sarama-producer ./bin/producer
# Note: robust_producer and enhanced_producer are temporarily disabled due to missing dependencies
