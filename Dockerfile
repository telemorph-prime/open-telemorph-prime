# Build stage
ARG GO_VERSION=1.24
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (only necessary files)
COPY main.go ./
COPY internal/ ./internal/
COPY web/ ./web/
COPY config.yaml ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o open-telemorph-prime .

# Final stage
FROM alpine:latest

# Install ca-certificates, sqlite, and wget for health checks
RUN apk --no-cache add ca-certificates sqlite wget

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/open-telemorph-prime .

# Copy web assets
COPY --from=builder /app/web ./web

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






