package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"open-telemorph-prime/internal/config"
	"open-telemorph-prime/internal/logger"
	"open-telemorph-prime/internal/storage"

	"github.com/gin-gonic/gin"
	semconv "go.opentelemetry.io/otel/semconv/v1.27.0"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Service struct {
	storage    storage.Storage
	config     config.IngestionConfig
	httpServer *http.Server
	grpcServer *grpc.Server
	logger     *zap.Logger
}

func NewService(storage storage.Storage, config config.IngestionConfig) *Service {
	return &Service{
		storage: storage,
		config:  config,
		logger:  logger.Get(),
	}
}

func (s *Service) Start() error {
	// Start HTTP server for OTLP HTTP endpoints if enabled
	if s.config.HTTPEnabled {
		go s.startHTTPServer()
		s.logger.Info("OTLP HTTP server enabled",
			zap.Int("port", s.config.HTTPPort),
			zap.String("protocol", "http"),
		)
	} else {
		s.logger.Info("OTLP HTTP server disabled")
	}

	// Start gRPC server for OTLP gRPC endpoints if enabled
	if s.config.GRPCEnabled {
		go s.startGRPCServer()
		s.logger.Info("OTLP gRPC server enabled",
			zap.Int("port", s.config.GRPCPort),
			zap.String("protocol", "grpc"),
		)
	} else {
		s.logger.Info("OTLP gRPC server disabled")
	}

	return nil
}

func (s *Service) startHTTPServer() {
	// Create Gin router for OTLP HTTP endpoints
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// OTLP HTTP endpoints
	otlp := router.Group("/v1")
	{
		otlp.POST("/traces", s.HandleTraces)
		otlp.POST("/metrics", s.HandleMetrics)
		otlp.POST("/logs", s.HandleLogs)
	}

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.HTTPPort),
		Handler: router,
	}

	s.logger.Info("Starting OTLP HTTP server",
		zap.Int("port", s.config.HTTPPort),
		zap.String("protocol", "http"),
	)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Error("Failed to start OTLP HTTP server",
			zap.Error(err),
			zap.Int("port", s.config.HTTPPort),
		)
	}
}

func (s *Service) startGRPCServer() {
	// Create a listener on the gRPC port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.GRPCPort))
	if err != nil {
		s.logger.Error("Failed to listen on gRPC port",
			zap.Error(err),
			zap.Int("port", s.config.GRPCPort),
		)
		return
	}

	// Create gRPC server
	s.grpcServer = grpc.NewServer()

	// TODO: Register OTLP gRPC services here
	// For now, we'll just start the server without any services
	// In a full implementation, you would register:
	// - traceservice.RegisterTraceServiceServer(s.grpcServer, s)
	// - metricsservice.RegisterMetricsServiceServer(s.grpcServer, s)
	// - logsservice.RegisterLogsServiceServer(s.grpcServer, s)

	s.logger.Info("Starting OTLP gRPC server",
		zap.Int("port", s.config.GRPCPort),
		zap.String("protocol", "grpc"),
	)
	if err := s.grpcServer.Serve(lis); err != nil {
		s.logger.Error("Failed to start gRPC server",
			zap.Error(err),
			zap.Int("port", s.config.GRPCPort),
		)
	}
}

func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping ingestion service...")

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.Error("Error shutting down OTLP HTTP server",
				zap.Error(err),
			)
		}
	}

	// Shutdown gRPC server
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}

	return nil
}

// HTTP handlers for OTLP endpoints
func (s *Service) HandleTraces(c *gin.Context) {
	var req struct {
		ResourceSpans []struct {
			Resource struct {
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"resource"`
			ScopeSpans []struct {
				Spans []struct {
					TraceId           string `json:"traceId"`
					SpanId            string `json:"spanId"`
					ParentSpanId      string `json:"parentSpanId"`
					Name              string `json:"name"`
					StartTimeUnixNano string `json:"startTimeUnixNano"`
					EndTimeUnixNano   string `json:"endTimeUnixNano"`
					Status            struct {
						Code string `json:"code"`
					} `json:"status"`
					Attributes []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes"`
				} `json:"spans"`
			} `json:"scopeSpans"`
		} `json:"resourceSpans"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process traces
	for _, resourceSpan := range req.ResourceSpans {
		serviceName := extractServiceNameFromResource(resourceSpan.Resource)

		for _, scopeSpan := range resourceSpan.ScopeSpans {
			for _, span := range scopeSpan.Spans {
				startTime, _ := time.Parse(time.RFC3339Nano, span.StartTimeUnixNano)
				endTime, _ := time.Parse(time.RFC3339Nano, span.EndTimeUnixNano)

				trace := &storage.Trace{
					TraceID:       span.TraceId,
					SpanID:        span.SpanId,
					ServiceName:   serviceName,
					OperationName: span.Name,
					StartTime:     startTime,
					DurationNanos: endTime.Sub(startTime).Nanoseconds(),
					StatusCode:    span.Status.Code,
					Attributes:    convertAttributesToJSON(span.Attributes),
				}

				if span.ParentSpanId != "" {
					trace.ParentSpanID = &span.ParentSpanId
				}

				if err := s.storage.InsertTrace(trace); err != nil {
					s.logger.Error("Failed to insert trace",
						zap.Error(err),
						zap.String("trace_id", trace.TraceID),
						zap.String("span_id", trace.SpanID),
					)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (s *Service) HandleMetrics(c *gin.Context) {
	var req struct {
		ResourceMetrics []struct {
			Resource struct {
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"resource"`
			ScopeMetrics []struct {
				Metrics []struct {
					Name string `json:"name"`
					Data struct {
						Gauge struct {
							DataPoints []struct {
								TimeUnixNano string  `json:"timeUnixNano"`
								AsDouble     float64 `json:"asDouble"`
								Attributes   []struct {
									Key   string `json:"key"`
									Value struct {
										StringValue string `json:"stringValue"`
									} `json:"value"`
								} `json:"attributes"`
							} `json:"dataPoints"`
						} `json:"gauge"`
						Sum struct {
							DataPoints []struct {
								TimeUnixNano string  `json:"timeUnixNano"`
								AsDouble     float64 `json:"asDouble"`
								Attributes   []struct {
									Key   string `json:"key"`
									Value struct {
										StringValue string `json:"stringValue"`
									} `json:"value"`
								} `json:"attributes"`
							} `json:"dataPoints"`
						} `json:"sum"`
					} `json:"data"`
				} `json:"metrics"`
			} `json:"scopeMetrics"`
		} `json:"resourceMetrics"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process metrics
	for _, resourceMetric := range req.ResourceMetrics {
		serviceName := extractServiceNameFromResource(resourceMetric.Resource)

		for _, scopeMetric := range resourceMetric.ScopeMetrics {
			for _, metric := range scopeMetric.Metrics {
				// Handle gauge metrics
				for _, dataPoint := range metric.Data.Gauge.DataPoints {
					timestamp, _ := time.Parse(time.RFC3339Nano, dataPoint.TimeUnixNano)
					metricData := &storage.Metric{
						MetricName:  metric.Name,
						Value:       dataPoint.AsDouble,
						Timestamp:   timestamp,
						ServiceName: serviceName,
						Labels:      convertAttributesToJSON(dataPoint.Attributes),
					}

					if err := s.storage.InsertMetric(metricData); err != nil {
						s.logger.Error("Failed to insert metric",
							zap.Error(err),
							zap.String("metric_name", metricData.MetricName),
							zap.String("service_name", metricData.ServiceName),
						)
					}
				}

				// Handle sum metrics
				for _, dataPoint := range metric.Data.Sum.DataPoints {
					timestamp, _ := time.Parse(time.RFC3339Nano, dataPoint.TimeUnixNano)
					metricData := &storage.Metric{
						MetricName:  metric.Name,
						Value:       dataPoint.AsDouble,
						Timestamp:   timestamp,
						ServiceName: serviceName,
						Labels:      convertAttributesToJSON(dataPoint.Attributes),
					}

					if err := s.storage.InsertMetric(metricData); err != nil {
						s.logger.Error("Failed to insert metric",
							zap.Error(err),
							zap.String("metric_name", metricData.MetricName),
							zap.String("service_name", metricData.ServiceName),
						)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (s *Service) HandleLogs(c *gin.Context) {
	var req struct {
		ResourceLogs []struct {
			Resource struct {
				Attributes []struct {
					Key   string `json:"key"`
					Value struct {
						StringValue string `json:"stringValue"`
					} `json:"value"`
				} `json:"attributes"`
			} `json:"resource"`
			ScopeLogs []struct {
				LogRecords []struct {
					TimeUnixNano string `json:"timeUnixNano"`
					SeverityText string `json:"severityText"`
					Body         struct {
						StringValue string `json:"stringValue"`
					} `json:"body"`
					Attributes []struct {
						Key   string `json:"key"`
						Value struct {
							StringValue string `json:"stringValue"`
						} `json:"value"`
					} `json:"attributes"`
					TraceId string `json:"traceId"`
					SpanId  string `json:"spanId"`
				} `json:"logRecords"`
			} `json:"scopeLogs"`
		} `json:"resourceLogs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process logs
	for _, resourceLog := range req.ResourceLogs {
		serviceName := extractServiceNameFromResource(resourceLog.Resource)

		for _, scopeLog := range resourceLog.ScopeLogs {
			for _, logRecord := range scopeLog.LogRecords {
				timestamp, _ := time.Parse(time.RFC3339Nano, logRecord.TimeUnixNano)

				logData := &storage.Log{
					Timestamp:   timestamp,
					ServiceName: serviceName,
					Level:       logRecord.SeverityText,
					Message:     logRecord.Body.StringValue,
					Attributes:  convertAttributesToJSON(logRecord.Attributes),
				}

				if logRecord.TraceId != "" {
					logData.TraceID = &logRecord.TraceId
				}
				if logRecord.SpanId != "" {
					logData.SpanID = &logRecord.SpanId
				}

				if err := s.storage.InsertLog(logData); err != nil {
					s.logger.Error("Failed to insert log",
						zap.Error(err),
						zap.String("service_name", logData.ServiceName),
						zap.String("level", logData.Level),
					)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// Helper functions
func extractServiceNameFromResource(resource struct {
	Attributes []struct {
		Key   string `json:"key"`
		Value struct {
			StringValue string `json:"stringValue"`
		} `json:"value"`
	} `json:"attributes"`
}) string {
	// Use OpenTelemetry semantic convention for service name
	// ServiceNameKey is an attribute.Key, convert to string for comparison
	serviceNameKey := string(semconv.ServiceNameKey)
	for _, attr := range resource.Attributes {
		if attr.Key == serviceNameKey {
			return attr.Value.StringValue
		}
	}

	return "unknown"
}

func convertAttributesToJSON(attributes []struct {
	Key   string `json:"key"`
	Value struct {
		StringValue string `json:"stringValue"`
	} `json:"value"`
}) string {
	if len(attributes) == 0 {
		return "{}"
	}

	attrs := make(map[string]interface{})
	for _, attr := range attributes {
		attrs[attr.Key] = attr.Value.StringValue
	}

	jsonData, err := json.Marshal(attrs)
	if err != nil {
		return "{}"
	}

	return string(jsonData)
}
