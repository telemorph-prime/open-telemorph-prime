# Open-Telemorph-Prime

A simplified, single-binary observability platform designed for home users and developers. Open-Telemorph-Prime eliminates the complexity of enterprise observability platforms while maintaining core functionality for ingesting, storing, and querying OpenTelemetry signals (traces, metrics, logs).

## 🚀 Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/open-telemorph-prime.git
cd open-telemorph-prime

# Start the service
docker-compose up -d

# Open your browser
open http://localhost:8080
```

#### Choosing Go Version for Docker Builds

Open-Telemorph-Prime requires Go ≥1.24. To maintain compatibility, you can specify which Go version to use when building the Docker image:

In your `Dockerfile`:
```dockerfile
ARG GO_VERSION=1.25
FROM golang:${GO_VERSION}-alpine AS builder
```

To build with a specific Go version (e.g., Go 1.25):
```bash
docker build --build-arg GO_VERSION=1.25 -t open-telemorph-prime .
```

If you need to support older Go versions (where compatible with your go.mod), modify the `GO_VERSION` arg accordingly.

### Option 2: Direct Binary

```bash
# Download the latest release
wget https://github.com/your-org/open-telemorph-prime/releases/latest/download/open-telemorph-prime-linux-amd64
chmod +x open-telemorph-prime-linux-amd64

# Run
./open-telemorph-prime-linux-amd64

# Open your browser
open http://localhost:8080
```

### Option 3: Build from Source

```bash
# Prerequisites: Go 1.24+, Node.js 18+, npm
git clone https://github.com/your-org/open-telemorph-prime.git
cd open-telemorph-prime

# Install dependencies (Go and React UI)
make deps

# Build (builds React UI first, then embeds it in Go binary)
make build

# Run
./open-telemorph-prime
```

**Note:** The frontend is embedded into the Go binary during build, so you only need to deploy a single executable file.

## 📊 Features

- **Single Binary**: One executable with embedded frontend, zero configuration
- **Minimal Resource Usage**: Runs on any modern machine (<2GB RAM)
- **OTLP Support**: Ingest traces, metrics, and logs via HTTP/gRPC
- **Web UI**: Simple, responsive interface for data exploration
- **SQLite Storage**: Lightweight, file-based storage
- **REST API**: Query your data programmatically
- **Health Checks**: Built-in monitoring endpoints

## 🔧 Configuration

Open-Telemorph-Prime uses a simple YAML configuration file:

```yaml
server:
  port: 8080
  environment: "development"

storage:
  type: "sqlite"
  path: "./data/telemorph.db"
  retention_days: 30

ingestion:
  grpc_port: 4317
  http_port: 4318

web:
  enabled: true
  title: "Open-Telemorph-Prime"
```

## 📡 Sending Data

Open-Telemorph-Prime uses standard OpenTelemetry Collector ports:
- **Port 4317**: OTLP gRPC endpoint (for traces, metrics, logs)
- **Port 4318**: OTLP HTTP endpoint (for traces, metrics, logs)  
- **Port 8080**: Web UI and REST API

### HTTP Endpoint

```bash
# Send traces
curl -X POST http://localhost:4318/v1/traces \
  -H "Content-Type: application/json" \
  -d '{"resourceSpans": [...]}'

# Send metrics
curl -X POST http://localhost:4318/v1/metrics \
  -H "Content-Type: application/json" \
  -d '{"resourceMetrics": [...]}'

# Send logs
curl -X POST http://localhost:4318/v1/logs \
  -H "Content-Type: application/json" \
  -d '{"resourceLogs": [...]}'
```

### OpenTelemetry SDK Integration

```go
// Go example - HTTP endpoint
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/http"
)

exporter, err := otlptracehttp.New(
    context.Background(),
    otlptracehttp.WithEndpoint("http://localhost:4318"),
    otlptracehttp.WithInsecure(),
)
```

```go
// Go example - gRPC endpoint
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace/grpc"
)

exporter, err := otlptracegrpc.New(
    context.Background(),
    otlptracegrpc.WithEndpoint("localhost:4317"),
    otlptracegrpc.WithInsecure(),
)
```

## 🔍 API Endpoints

### Health
- `GET /health` - Health check
- `GET /ready` - Readiness check

### Data
- `GET /api/v1/metrics` - List metrics
- `GET /api/v1/traces` - List traces
- `GET /api/v1/logs` - List logs
- `GET /api/v1/services` - List services
- `POST /api/v1/query` - Generic query endpoint

### Web UI
- `GET /` - Home page
- `GET /dashboard` - Dashboard
- `GET /metrics` - Metrics explorer
- `GET /traces` - Traces explorer
- `GET /logs` - Logs viewer

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Data Sources                                  │
│  Applications with OTEL SDKs → Direct Ingestion                 │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│              Unified Service (Single Binary)                    │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │ gRPC/HTTP    │  │   Storage    │  │   Query      │          │
│  │ Receivers    │  │   Engine     │  │   Engine     │          │
│  │ (OTLP)       │  │ (SQLite/DB)  │  │ (Built-in)   │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          │
│  │   Web UI     │  │   REST API   │  │   Health     │          │
│  │  (Embedded)  │  │  Endpoints   │  │   Checks     │          │
│  └──────────────┘  └──────────────┘  └──────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## 🛠️ Development

### Prerequisites
- Go 1.24+
- Node.js 18+ and npm (for building the React UI)
- Docker (optional)
- SQLite3 (for local development)

### Building

```bash
# Install all dependencies (Go and React UI)
make deps
# Or manually:
# cd backend && go mod tidy
# cd frontend && npm install

# Build React UI and Go binary (frontend is embedded in binary)
make build
# Or manually:
# cd frontend && npm run build
# cd backend && go build -o ../open-telemorph-prime .

# Build only the React UI (for UI development)
make build-ui

# Run tests
make test

# Run in development mode
make dev

# For React UI development (hot reload)
cd frontend && npm run dev
```

### Project Structure

```
open-telemorph-prime/
├── backend/               # Go backend source code
│   ├── main.go           # Entry point
│   ├── internal/        # Internal packages
│   │   ├── config/      # Configuration management
│   │   ├── ingestion/   # OTLP receivers
│   │   ├── storage/     # SQLite storage
│   │   └── web/         # Web API handlers
│   ├── go.mod            # Go dependencies
│   └── go.sum            # Go dependencies checksum
├── frontend/              # React frontend source code
│   ├── public/          # Public static files
│   │   └── index.html   # HTML entry point
│   ├── src/             # React source code
│   │   ├── components/  # React components
│   │   │   ├── pages/   # Page components
│   │   │   └── ui/      # UI component library
│   │   ├── assets/      # Images and static assets
│   │   └── styles/      # Global styles
│   ├── dist/            # Built React app (generated, embedded in Go binary)
│   ├── package.json     # Node.js dependencies
│   ├── vite.config.ts   # Vite build configuration
│   ├── tailwind.config.js # Tailwind CSS configuration
│   └── tsconfig.json    # TypeScript configuration
├── config.yaml           # Default configuration
├── Dockerfile            # Docker build configuration
└── docker-compose.yml    # Docker Compose configuration
```

## 📈 Performance

### Simple Mode (Default)
- **Ingestion**: 1K+ spans/sec, 10K+ metrics/sec
- **Query Latency**: <100ms for simple queries
- **Storage**: <1GB for 30 days of data
- **Memory Usage**: <512MB RAM
- **Startup Time**: <5 seconds

## 🔒 Security

- CORS enabled for cross-origin requests
- Input validation on all endpoints
- SQL injection protection via prepared statements
- Rate limiting (configurable)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [OpenTelemetry](https://opentelemetry.io/) for the telemetry standards
- [Gin](https://gin-gonic.com/) for the web framework
- [SQLite](https://sqlite.org/) for the embedded database

---

**Open-Telemorph-Prime**: Observability made simple. 🚀

*Perfect for home labs, development environments, and anyone who wants observability without the complexity.*