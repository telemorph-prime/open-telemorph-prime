package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"open-telemorph-prime/internal/config"
	otlpgrpc "open-telemorph-prime/internal/grpc"
	"open-telemorph-prime/internal/storage"

	"github.com/gin-gonic/gin"
)

type Service struct {
	storage    storage.Storage
	config     config.IngestionConfig
	httpServer *http.Server
	grpcServer *otlpgrpc.Server
}

func NewService(storage storage.Storage, config config.IngestionConfig) *Service {
	return &Service{
		storage: storage,
		config:  config,
	}
}

func (s *Service) Start() error {
	// Start HTTP server for OTLP HTTP endpoints if enabled
	if s.config.HTTPEnabled {
		go s.startHTTPServer()
		log.Printf("OTLP HTTP server enabled on port %d", s.config.HTTPPort)
	} else {
		log.Printf("OTLP HTTP server disabled")
	}

	// Start gRPC server for OTLP gRPC endpoints if enabled
	if s.config.GRPCEnabled {
		go s.startGRPCServer()
		log.Printf("OTLP gRPC server enabled on port %d", s.config.GRPCPort)
	} else {
		log.Printf("OTLP gRPC server disabled")
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

	log.Printf("Starting OTLP HTTP server on port %d", s.config.HTTPPort)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Failed to start OTLP HTTP server: %v", err)
	}
}

func (s *Service) startGRPCServer() {
	// Create our custom gRPC server with all OTLP services registered
	s.grpcServer = otlpgrpc.NewServer(s.storage, s.config.GRPCPort)

	// Start the server
	if err := s.grpcServer.Start(); err != nil {
		log.Printf("Failed to start gRPC server: %v", err)
	}
}

func (s *Service) Stop(ctx context.Context) error {
	log.Println("Stopping ingestion service...")

	// Shutdown HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down OTLP HTTP server: %v", err)
		}
	}

	// Shutdown gRPC server
	if s.grpcServer != nil {
		s.grpcServer.Stop()
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
					log.Printf("Failed to insert trace: %v", err)
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
						log.Printf("Failed to insert metric: %v", err)
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
						log.Printf("Failed to insert metric: %v", err)
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
					log.Printf("Failed to insert log: %v", err)
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
	for _, attr := range resource.Attributes {
		if attr.Key == "service.name" {
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
