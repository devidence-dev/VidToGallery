# Multi-stage build optimized for Raspberry Pi 5 (ARM64)
FROM --platform=$BUILDPLATFORM golang:1.24.3-bookworm AS builder

# Build arguments for cross-compilation with defaults
ARG TARGETPLATFORM=linux/arm64
ARG BUILDPLATFORM=linux/amd64
ARG TARGETOS=linux
ARG TARGETARCH=arm64

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with module cache
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations for ARM64
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS:-linux} \
    GOARCH=${TARGETARCH:-arm64} \
    go build -ldflags="-w -s -extldflags '-static'" \
    -a -installsuffix cgo \
    -o bin/server ./cmd/server

# Final stage - minimal runtime image
FROM python:3.11.13-alpine3.22

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    wget

# Install yt-dlp with specific version for stability
RUN pip3 install --no-cache-dir yt-dlp==2024.12.13

# Create non-root user for security
RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/bin/server .

# Set proper ownership and permissions
RUN chown appuser:appgroup /app/server && \
    chmod +x /app/server

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Add metadata labels
LABEL maintainer="vidtogallery" \
      version="1.0" \
      description="VidToGallery - Video downloader service optimized for Raspberry Pi 5"

# Optimized health check for ARM64/RPi5
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider --timeout=5 http://localhost:8080/health || exit 1

# Run the application
CMD ["./server"]
