package dogfood

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"open-telemorph-prime/internal/config"
	"open-telemorph-prime/internal/storage"
)

type Service struct {
	config     config.WebConfig
	storage    storage.Storage
	client     *http.Client
	enabled    bool
	serverPort int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewService(config config.WebConfig, storage storage.Storage, serverPort int) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	return &Service{
		config:     config,
		storage:    storage,
		client:     &http.Client{Timeout: 5 * time.Second},
		enabled:    config.Dogfood,
		serverPort: serverPort,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Service) Start(ctx context.Context) {
	// Always start the service, but only collect telemetry when enabled
	go s.runCollectionLoop()
}

func (s *Service) runCollectionLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if s.enabled {
				s.collectAndSendTelemetry()
			}
		}
	}
}

func (s *Service) SetEnabled(enabled bool) {
	s.enabled = enabled
	if enabled {
		log.Println("Dogfood monitoring enabled")
	} else {
		log.Println("Dogfood monitoring disabled")
	}
}

func (s *Service) IsEnabled() bool {
	return s.enabled
}

func (s *Service) collectAndSendTelemetry() {
	if !s.enabled {
		return
	}

	log.Println("Dogfood: Collecting telemetry data...")

	// Collect metrics
	metrics := s.collectMetrics()
	traces := s.collectTraces()
	logs := s.collectLogs()

	// Send to internal endpoints
	s.sendMetrics(metrics)
	s.sendTraces(traces)
	s.sendLogs(logs)

	log.Println("Dogfood: Telemetry collection completed")
}

func (s *Service) collectMetrics() []map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	now := time.Now()
	serviceName := "open-telemorph-prime"

	metrics := []map[string]interface{}{
		{
			"resource": map[string]interface{}{
				"attributes": []map[string]interface{}{
					{"key": "service.name", "value": map[string]interface{}{"stringValue": serviceName}},
					{"key": "service.version", "value": map[string]interface{}{"stringValue": "0.2.1"}},
					{"key": "service.instance.id", "value": map[string]interface{}{"stringValue": "telemorph-1"}},
				},
			},
			"scopeMetrics": []map[string]interface{}{
				{
					"scope": map[string]interface{}{
						"name":    "open-telemorph-prime",
						"version": "0.2.1",
					},
					"metrics": []map[string]interface{}{
						{
							"name": "memory.usage",
							"data": map[string]interface{}{
								"gauge": map[string]interface{}{
									"dataPoints": []map[string]interface{}{
										{
											"timeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
											"asDouble":     float64(m.Alloc),
										},
									},
								},
							},
						},
						{
							"name": "memory.heap.size",
							"data": map[string]interface{}{
								"gauge": map[string]interface{}{
									"dataPoints": []map[string]interface{}{
										{
											"timeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
											"asDouble":     float64(m.HeapSys),
										},
									},
								},
							},
						},
						{
							"name": "gc.collections",
							"data": map[string]interface{}{
								"sum": map[string]interface{}{
									"dataPoints": []map[string]interface{}{
										{
											"timeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
											"asDouble":     float64(m.NumGC),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return metrics
}

func (s *Service) collectTraces() []map[string]interface{} {
	now := time.Now()
	serviceName := "open-telemorph-prime"
	traceID := fmt.Sprintf("%016x", now.UnixNano())
	spanID := fmt.Sprintf("%016x", now.UnixNano()+1)

	traces := []map[string]interface{}{
		{
			"resource": map[string]interface{}{
				"attributes": []map[string]interface{}{
					{"key": "service.name", "value": map[string]interface{}{"stringValue": serviceName}},
					{"key": "service.version", "value": map[string]interface{}{"stringValue": "0.2.1"}},
				},
			},
			"scopeSpans": []map[string]interface{}{
				{
					"scope": map[string]interface{}{
						"name":    "open-telemorph-prime",
						"version": "0.2.1",
					},
					"spans": []map[string]interface{}{
						{
							"traceId":           traceID,
							"spanId":            spanID,
							"name":              "self-monitoring.collect",
							"kind":              1, // INTERNAL
							"startTimeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
							"endTimeUnixNano":   fmt.Sprintf("%d", now.Add(10*time.Millisecond).UnixNano()),
							"status": map[string]interface{}{
								"code": "STATUS_CODE_OK",
							},
							"attributes": []map[string]interface{}{
								{"key": "component", "value": map[string]interface{}{"stringValue": "dogfood"}},
								{"key": "operation", "value": map[string]interface{}{"stringValue": "telemetry_collection"}},
							},
						},
					},
				},
			},
		},
	}

	return traces
}

func (s *Service) collectLogs() []map[string]interface{} {
	now := time.Now()
	serviceName := "open-telemorph-prime"

	logs := []map[string]interface{}{
		{
			"resource": map[string]interface{}{
				"attributes": []map[string]interface{}{
					{"key": "service.name", "value": map[string]interface{}{"stringValue": serviceName}},
					{"key": "service.version", "value": map[string]interface{}{"stringValue": "0.2.1"}},
				},
			},
			"scopeLogs": []map[string]interface{}{
				{
					"scope": map[string]interface{}{
						"name":    "open-telemorph-prime",
						"version": "0.2.1",
					},
					"logRecords": []map[string]interface{}{
						{
							"timeUnixNano": fmt.Sprintf("%d", now.UnixNano()),
							"severityText": "INFO",
							"body": map[string]interface{}{
								"stringValue": "Dogfood telemetry collection completed",
							},
							"attributes": []map[string]interface{}{
								{"key": "component", "value": map[string]interface{}{"stringValue": "dogfood"}},
								{"key": "operation", "value": map[string]interface{}{"stringValue": "telemetry_collection"}},
							},
						},
					},
				},
			},
		},
	}

	return logs
}

func (s *Service) sendMetrics(metrics []map[string]interface{}) {
	payload := map[string]interface{}{
		"resourceMetrics": metrics,
	}

	s.sendToEndpoint("/v1/metrics", payload)
}

func (s *Service) sendTraces(traces []map[string]interface{}) {
	payload := map[string]interface{}{
		"resourceSpans": traces,
	}

	s.sendToEndpoint("/v1/traces", payload)
}

func (s *Service) sendLogs(logs []map[string]interface{}) {
	payload := map[string]interface{}{
		"resourceLogs": logs,
	}

	s.sendToEndpoint("/v1/logs", payload)
}

func (s *Service) sendToEndpoint(endpoint string, payload interface{}) {
	// Send to the OTLP HTTP ingestion endpoint on port 4318
	url := fmt.Sprintf("http://localhost:4318%s", endpoint)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal dogfood payload: %v", err)
		return
	}

	resp, err := s.client.Post(url, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("Failed to send dogfood telemetry to %s: %v", endpoint, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Dogfood telemetry endpoint %s returned status %d", endpoint, resp.StatusCode)
	}
}
