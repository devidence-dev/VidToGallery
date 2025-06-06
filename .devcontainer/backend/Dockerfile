FROM golang:1.24.3-bookworm

# Dev command to run Go Fiber with live reload:
# go install github.com/air-verse/air@latest && air

WORKDIR /workspace

# Install development dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    redis-tools \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# Install yt-dlp
RUN pip3 install --break-system-packages yt-dlp

# Install Air for hot reload in development
RUN go install github.com/air-verse/air@latest

# Keep container alive for development
CMD ["sleep", "infinity"]
