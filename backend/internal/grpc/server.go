package grpc

import (
	"fmt"
	"log"
	"net"

	"open-telemorph-prime/internal/storage"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	grpcServer     *grpc.Server
	traceService   *TraceService
	metricsService *MetricsService
	logsService    *LogsService
	port           int
}

func NewServer(storage storage.Storage, port int) *Server {
	// Create gRPC server with options
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(4*1024*1024), // 4MB max message size
		grpc.MaxSendMsgSize(4*1024*1024), // 4MB max message size
	)

	// Create service instances
	traceService := NewTraceService(storage)
	metricsService := NewMetricsService(storage)
	logsService := NewLogsService(storage)

	// Register services with gRPC server
	coltracepb.RegisterTraceServiceServer(grpcServer, traceService)
	colmetricspb.RegisterMetricsServiceServer(grpcServer, metricsService)
	collogspb.RegisterLogsServiceServer(grpcServer, logsService)

	// Enable gRPC reflection for debugging
	reflection.Register(grpcServer)

	return &Server{
		grpcServer:     grpcServer,
		traceService:   traceService,
		metricsService: metricsService,
		logsService:    logsService,
		port:           port,
	}
}

func (s *Server) Start() error {
	// Create a listener on the specified port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	log.Printf("Starting OTLP gRPC server on port %d", s.port)
	log.Printf("Registered services:")
	log.Printf("  - TraceService (opentelemetry.proto.collector.trace.v1.TraceService)")
	log.Printf("  - MetricsService (opentelemetry.proto.collector.metrics.v1.MetricsService)")
	log.Printf("  - LogsService (opentelemetry.proto.collector.logs.v1.LogsService)")
	log.Printf("  - gRPC Reflection enabled")

	// Start serving
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	log.Println("Stopping OTLP gRPC server...")
	s.grpcServer.GracefulStop()
	log.Println("OTLP gRPC server stopped")
}

func (s *Server) GetServer() *grpc.Server {
	return s.grpcServer
}
