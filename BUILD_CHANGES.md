# Build Process Changes with New React UI

This document outlines the changes to the build and run process after integrating the new React UI.

## Summary of Changes

The application now uses a modern React-based UI (from the `frontend` directory) instead of the previous static HTML/JavaScript files. The frontend is embedded into the Go binary during build, creating a single executable deployment.

## New Prerequisites

**Additional Requirements:**
- **Node.js 18+** and **npm** (for building the React UI)
- Previously only required Go 1.24+

## Build Process

### Using Make (Recommended)

The Makefile has been updated to handle the React UI build automatically:

```bash
# Install all dependencies (Go + React UI)
make deps

# Build everything (React UI + Go binary)
make build

# Build only the React UI
make build-ui

# Run the application
make run

# Clean build artifacts (including React build)
make clean
```

### Manual Build Process

If you prefer to build manually:

```bash
# 1. Install Go dependencies
go mod tidy

# 2. Install React UI dependencies
cd frontend
npm install

# 3. Build React UI (outputs to frontend/dist/)
npm run build
cd ..

# 4. Build Go binary (embeds frontend/dist/)
cd backend
go build -o ../open-telemorph-prime .
```

## Development Workflow

### Backend Development

For Go backend development, the workflow remains the same:

```bash
# Run the Go server (serves the embedded React UI)
cd backend
go run main.go -config ../config.yaml
```

**Note:** Make sure to build the React UI first (`make build-ui` or `cd frontend && npm run build`) before building the Go binary, as the frontend is embedded during the Go build.

### Frontend Development

For React UI development with hot reload:

```bash
# Start Vite dev server (runs on port 3000)
cd frontend
npm run dev
```

**Note:** The Vite dev server runs independently and proxies API calls. For full integration testing, you'll need to:
1. Run the Go server on port 8080
2. Configure Vite to proxy API requests to `http://localhost:8080/api/v1/*`

## Docker Build

The Dockerfile has been updated to build the React UI as part of the Docker image:

```bash
# Build Docker image (includes React UI build)
docker build -t open-telemorph-prime .

# Or using Make
make docker-build
```

The Docker build process:
1. Installs Node.js and npm in the builder stage
2. Builds the React UI (`frontend/` → `frontend/dist/`)
3. Builds the Go binary (embeds `frontend/dist/` via `go:embed`)
4. Creates a single executable with embedded frontend

## File Structure Changes

### New Files/Directories
- `frontend/` - React UI source code
- `backend/` - Go backend source code
- `frontend/dist/` - Built React app (generated, embedded in Go binary)

### Deprecated (but still present)
- `web/*.html` - Old HTML templates (no longer served)
- `web/static/` - Old static JavaScript/CSS (no longer served)

The Go server now serves:
- `/assets/*` - React app assets from embedded filesystem
- `/*` - React SPA (all routes serve embedded `index.html`)

## Running the Application

### Production Build

```bash
# Build everything
make build

# Run
./open-telemorph-prime
```

The application will serve the React UI at `http://localhost:8080` (or your configured port).

### Development Mode

**Option 1: Full Stack (Recommended for integration testing)**
```bash
# Terminal 1: Build UI once
make build-ui

# Terminal 1: Run Go server
cd backend && go run main.go -config ../config.yaml

# Terminal 2: React dev server (for UI hot reload)
cd frontend && npm run dev
```

**Option 2: Backend Only**
```bash
# Build UI once
make build-ui

# Run Go server (serves embedded UI)
cd backend && go run main.go -config ../config.yaml
```

## Troubleshooting

### "404 Not Found" for UI routes
- Make sure you've built the React UI: `make build-ui`
- Check that `frontend/dist/index.html` exists
- Rebuild the Go binary after building the frontend: `make build`

### API calls failing
- The React app makes API calls to `/api/v1/*` endpoints
- These are served by the Go server on the same port (8080 by default)
- No CORS configuration needed (same origin)

### Docker build fails
- Ensure Node.js 18+ is available in the build environment
- Check that `frontend/package.json` is present
- Verify npm can install dependencies

### UI not updating after changes
- Rebuild the React UI: `make build-ui`
- Rebuild the Go binary to embed the new frontend: `make build`
- Or use the Vite dev server for hot reload: `cd frontend && npm run dev`

## Migration Notes

- The old HTML templates in `web/*.html` are no longer used but kept for reference
- All UI routes (`/dashboard`, `/metrics`, etc.) are now handled by React Router
- API endpoints remain unchanged at `/api/v1/*`
- The Go server now serves a Single Page Application (SPA) instead of multiple HTML pages

