package grpc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"open-telemorph-prime/internal/storage"

	coltracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	tracepb "go.opentelemetry.io/proto/otlp/trace/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TraceService struct {
	coltracepb.UnimplementedTraceServiceServer
	storage storage.Storage
}

func NewTraceService(storage storage.Storage) *TraceService {
	return &TraceService{
		storage: storage,
	}
}

// Export implements the TraceServiceServer interface
func (s *TraceService) Export(ctx context.Context, req *coltracepb.ExportTraceServiceRequest) (*coltracepb.ExportTraceServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Process each resource span
	for _, resourceSpan := range req.ResourceSpans {
		if err := s.processResourceSpan(resourceSpan); err != nil {
			log.Printf("Failed to process resource span: %v", err)
			// Continue processing other spans even if one fails
		}
	}

	return &coltracepb.ExportTraceServiceResponse{
		PartialSuccess: &coltracepb.ExportTracePartialSuccess{
			RejectedSpans: 0, // We process all spans successfully
			ErrorMessage:  "",
		},
	}, nil
}

func (s *TraceService) processResourceSpan(resourceSpan *tracepb.ResourceSpans) error {
	// Extract service name from resource attributes
	serviceName := s.extractServiceName(resourceSpan.Resource)

	// Process each scope span
	for _, scopeSpan := range resourceSpan.ScopeSpans {
		for _, span := range scopeSpan.Spans {
			if err := s.processSpan(span, serviceName); err != nil {
				log.Printf("Failed to process span: %v", err)
				// Continue processing other spans
			}
		}
	}

	return nil
}

func (s *TraceService) processSpan(span *tracepb.Span, serviceName string) error {
	// Convert protobuf span to our internal trace format
	trace := &storage.Trace{
		TraceID:       string(span.TraceId),
		SpanID:        string(span.SpanId),
		ServiceName:   serviceName,
		OperationName: span.Name,
		StartTime:     time.Unix(0, int64(span.StartTimeUnixNano)),
		DurationNanos: int64(span.EndTimeUnixNano - span.StartTimeUnixNano),
		StatusCode:    s.convertStatusCode(span.Status),
		Attributes:    s.convertAttributes(span.Attributes),
	}

	// Set parent span ID if present
	if span.ParentSpanId != nil && len(span.ParentSpanId) > 0 {
		parentSpanID := string(span.ParentSpanId)
		trace.ParentSpanID = &parentSpanID
	}

	// Insert trace into storage
	return s.storage.InsertTrace(trace)
}

func (s *TraceService) extractServiceName(resource *resourcepb.Resource) string {
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

func (s *TraceService) convertStatusCode(status *tracepb.Status) string {
	if status == nil {
		return "UNSET"
	}

	switch status.Code {
	case tracepb.Status_STATUS_CODE_UNSET:
		return "UNSET"
	case tracepb.Status_STATUS_CODE_OK:
		return "OK"
	case tracepb.Status_STATUS_CODE_ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func (s *TraceService) convertAttributes(attributes []*commonpb.KeyValue) string {
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

func (s *TraceService) convertAttributeValue(value *commonpb.AnyValue) interface{} {
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
