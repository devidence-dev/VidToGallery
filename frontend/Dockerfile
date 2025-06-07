# Multi-stage build for production-ready Vue.js app on Raspberry Pi 5
# Stage 1: Build stage
FROM node:22.16.0-bookworm-slim AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install all dependencies (including devDependencies for build)
RUN npm ci && \
    npm cache clean --force

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Production stage with Nginx
FROM nginx:stable-alpine3.21 AS production

# Install security updates
RUN apk update && apk upgrade && \
    apk add --no-cache \
    ca-certificates \
    tzdata && \
    rm -rf /var/cache/apk/*

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# Copy built application from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Set proper ownership and permissions for static files only
RUN chown -R appuser:appgroup /usr/share/nginx/html

# Create nginx cache directories (will be overridden by tmpfs but needed for mkdir)
RUN mkdir -p /var/cache/nginx/client_temp \
             /var/cache/nginx/proxy_temp \
             /var/cache/nginx/fastcgi_temp \
             /var/cache/nginx/uwsgi_temp \
             /var/cache/nginx/scgi_temp

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/health || exit 1

# Expose port
EXPOSE 8000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
