FROM node:22.4-bookworm-slim

# Dev command to run Svelte in development mode with hot reload:
# npm run dev -- --host 0.0.0.0 --port 5173

WORKDIR /workspace

# Install development dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    netcat-openbsd \
    && rm -rf /var/lib/apt/lists/*

# Keep container alive for development
CMD ["sleep", "infinity"]
