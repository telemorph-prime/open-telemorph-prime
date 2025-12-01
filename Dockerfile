# Build stage
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Install dependencies for building
RUN apk add --no-cache git nodejs npm

# Build React UI first
COPY frontend/ ./frontend/
WORKDIR /app/frontend
RUN npm install && npm run build

# Back to app root
WORKDIR /app

# Copy Go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ ./backend/

# Copy frontend dist into backend directory for embedding (needed for go:embed)
RUN cp -r /app/frontend/dist /app/backend/dist

# Copy config
COPY config.yaml ./

# Build the application (frontend/dist is embedded via go:embed)
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ../open-telemorph-prime .

# Final stage
FROM alpine:latest

# Install ca-certificates, sqlite, and wget for health checks
RUN apk --no-cache add ca-certificates sqlite wget

WORKDIR /app

# Copy the binary from builder stage (frontend is embedded)
COPY --from=builder /app/open-telemorph-prime .

# Copy default config
COPY --from=builder /app/config.yaml .

# Create data directory
RUN mkdir -p data

# Expose ports
EXPOSE 8080 4317 4318

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./open-telemorph-prime"]


