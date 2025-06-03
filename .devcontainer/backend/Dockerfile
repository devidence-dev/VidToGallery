FROM golang:1.21-bullseye

# Dev command to run Go Fiber with live reload:
# go install github.com/cosmtrek/air@latest && air

WORKDIR /workspace

# Install development dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    redis-tools \
    && rm -rf /var/lib/apt/lists/*

# Install Air for hot reload in development
RUN go install github.com/cosmtrek/air@latest

# Keep container alive for development
CMD ["sleep", "infinity"]
