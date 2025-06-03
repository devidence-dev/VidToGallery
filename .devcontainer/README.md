# VidToGallery Development Environment

This project uses separate devcontainers for frontend and backend development with Redis.

## Setup Instructions

### 1. Frontend Development (Svelte 5)
1. Open the project in VS Code
2. Choose "Reopen in Container" and select "Frontend"
3. VS Code will build and start the frontend container
4. Once inside the container, run:
   ```bash
   npm run dev -- --host 0.0.0.0 --port 5173
   ```

### 2. Backend Development (Go + Redis)
1. Open a new VS Code window
2. Open the same project folder
3. Choose "Reopen in Container" and select "Backend"
4. VS Code will build and start the backend container with Redis
5. Once inside the container, run:
   ```bash
   air
   ```

**Important**: Start the backend devcontainer first to create the shared network.

## Services Available

- **Frontend**: http://localhost:5173 (Svelte dev server)
- **Backend**: http://localhost:8080 (Go Fiber API)
- **Redis**: localhost:6379 (accessible from backend container)

## Network Communication

- **Frontend → Backend**: Uses `http://localhost:8080` from browser
- **Container → Container**: Uses shared `vidtogallery-network`
- **Environment Variable**: `VITE_API_URL=http://localhost:8080` (frontend)

## Development Commands

### Frontend Container
- Start dev server: `npm run dev -- --host 0.0.0.0 --port 5173`
- Build: `npm run build`
- Preview: `npm run preview`

### Backend Container
- Start with hot reload: `air`
- Run without hot reload: `go run cmd/main.go`
- Build: `go build -o bin/vidtogallery cmd/main.go`
- Test: `go test ./...`

### Redis Commands (from backend container)
- Connect to Redis: `redis-cli -h redis`
- Monitor Redis: `redis-cli -h redis monitor`

## Container Features

### Frontend Container
- Node.js 20
- Git
- VS Code extensions for Svelte, Prettier, TailwindCSS
- Hot reload support

### Backend Container
- Go 1.21
- Git
- Redis CLI tools
- Air for hot reload
- VS Code Go extension

## Project Structure Expected
```
/workspace
├── frontend/
│   ├── src/
│   ├── package.json
│   └── ...
├── backend/
│   ├── cmd/
│   ├── pkg/
│   ├── go.mod
│   └── ...
└── .devcontainer/
```

## Tips
- Both containers use `sleep infinity` so they stay alive until you manually start the services
- **Start backend first** to create the shared Docker network
- Frontend changes will auto-reload when using `npm run dev`
- Backend changes will auto-reload when using `air`
- Redis data persists between container restarts
- You can run both devcontainers simultaneously in separate VS Code windows
- Frontend can communicate with backend via shared network
