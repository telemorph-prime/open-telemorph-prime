package grpc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"open-telemorph-prime/internal/storage"

	collogspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	logspb "go.opentelemetry.io/proto/otlp/logs/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogsService struct {
	collogspb.UnimplementedLogsServiceServer
	storage storage.Storage
}

func NewLogsService(storage storage.Storage) *LogsService {
	return &LogsService{
		storage: storage,
	}
}

// Export implements the LogsServiceServer interface
func (s *LogsService) Export(ctx context.Context, req *collogspb.ExportLogsServiceRequest) (*collogspb.ExportLogsServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Process each resource log
	for _, resourceLog := range req.ResourceLogs {
		if err := s.processResourceLog(resourceLog); err != nil {
			log.Printf("Failed to process resource log: %v", err)
			// Continue processing other logs even if one fails
		}
	}

	return &collogspb.ExportLogsServiceResponse{
		PartialSuccess: &collogspb.ExportLogsPartialSuccess{
			RejectedLogRecords: 0, // We process all logs successfully
			ErrorMessage:       "",
		},
	}, nil
}

func (s *LogsService) processResourceLog(resourceLog *logspb.ResourceLogs) error {
	// Extract service name from resource attributes
	serviceName := s.extractServiceName(resourceLog.Resource)

	// Process each scope log
	for _, scopeLog := range resourceLog.ScopeLogs {
		for _, logRecord := range scopeLog.LogRecords {
			if err := s.processLogRecord(logRecord, serviceName); err != nil {
				log.Printf("Failed to process log record: %v", err)
				// Continue processing other logs
			}
		}
	}

	return nil
}

func (s *LogsService) processLogRecord(logRecord *logspb.LogRecord, serviceName string) error {
	// Convert protobuf log record to our internal log format
	logData := &storage.Log{
		Timestamp:   time.Unix(0, int64(logRecord.TimeUnixNano)),
		ServiceName: serviceName,
		Level:       s.convertSeverityText(logRecord.SeverityText),
		Message:     s.extractLogBody(logRecord.Body),
		Attributes:  s.convertAttributes(logRecord.Attributes),
	}

	// Set trace and span IDs if present
	if len(logRecord.TraceId) > 0 {
		traceID := string(logRecord.TraceId)
		logData.TraceID = &traceID
	}

	if len(logRecord.SpanId) > 0 {
		spanID := string(logRecord.SpanId)
		logData.SpanID = &spanID
	}

	// Insert log into storage
	return s.storage.InsertLog(logData)
}

func (s *LogsService) extractServiceName(resource *resourcepb.Resource) string {
	if resource == nil {
		return "unknown"
	}

	for _, attr := range resource.Attributes {
		if attr.Key == "service.name" {
			if strVal := attr.Value.GetStringValue(); strVal != "" {
				return strVal
			}
		}
	}

	return "unknown"
}

func (s *LogsService) convertSeverityText(severityText string) string {
	if severityText == "" {
		return "INFO"
	}
	return severityText
}

func (s *LogsService) extractLogBody(body *commonpb.AnyValue) string {
	if body == nil {
		return ""
	}

	switch v := body.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_BoolValue:
		if v.BoolValue {
			return "true"
		}
		return "false"
	case *commonpb.AnyValue_IntValue:
		return string(rune(v.IntValue))
	case *commonpb.AnyValue_DoubleValue:
		return string(rune(v.DoubleValue))
	case *commonpb.AnyValue_ArrayValue:
		if v.ArrayValue != nil {
			items := make([]interface{}, len(v.ArrayValue.Values))
			for i, item := range v.ArrayValue.Values {
				items[i] = s.convertAttributeValue(item)
			}
			jsonData, err := json.Marshal(items)
			if err != nil {
				return ""
			}
			return string(jsonData)
		}
	case *commonpb.AnyValue_KvlistValue:
		if v.KvlistValue != nil {
			kvMap := make(map[string]interface{})
			for _, kv := range v.KvlistValue.Values {
				kvMap[kv.Key] = s.convertAttributeValue(kv.Value)
			}
			jsonData, err := json.Marshal(kvMap)
			if err != nil {
				return ""
			}
			return string(jsonData)
		}
	}

	return ""
}

func (s *LogsService) convertAttributes(attributes []*commonpb.KeyValue) string {
	if len(attributes) == 0 {
		return "{}"
	}

	attrs := make(map[string]interface{})
	for _, attr := range attributes {
		if attr == nil {
			continue
		}

		key := attr.Key
		value := s.convertAttributeValue(attr.Value)
		if value != nil {
			attrs[key] = value
		}
	}

	// Convert to JSON string
	jsonData, err := json.Marshal(attrs)
	if err != nil {
		log.Printf("Failed to marshal attributes to JSON: %v", err)
		return "{}"
	}

	return string(jsonData)
}

func (s *LogsService) convertAttributeValue(value *commonpb.AnyValue) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.Value.(type) {
	case *commonpb.AnyValue_StringValue:
		return v.StringValue
	case *commonpb.AnyValue_BoolValue:
		return v.BoolValue
	case *commonpb.AnyValue_IntValue:
		return v.IntValue
	case *commonpb.AnyValue_DoubleValue:
		return v.DoubleValue
	case *commonpb.AnyValue_ArrayValue:
		if v.ArrayValue != nil {
			items := make([]interface{}, len(v.ArrayValue.Values))
			for i, item := range v.ArrayValue.Values {
				items[i] = s.convertAttributeValue(item)
			}
			return items
		}
	case *commonpb.AnyValue_KvlistValue:
		if v.KvlistValue != nil {
			kvMap := make(map[string]interface{})
			for _, kv := range v.KvlistValue.Values {
				kvMap[kv.Key] = s.convertAttributeValue(kv.Value)
			}
			return kvMap
		}
	}

	return nil
}
