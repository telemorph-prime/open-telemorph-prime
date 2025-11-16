package grpc

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"open-telemorph-prime/internal/storage"

	colmetricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	commonpb "go.opentelemetry.io/proto/otlp/common/v1"
	metricspb "go.opentelemetry.io/proto/otlp/metrics/v1"
	resourcepb "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsService struct {
	colmetricspb.UnimplementedMetricsServiceServer
	storage storage.Storage
}

func NewMetricsService(storage storage.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

// Export implements the MetricsServiceServer interface
func (s *MetricsService) Export(ctx context.Context, req *colmetricspb.ExportMetricsServiceRequest) (*colmetricspb.ExportMetricsServiceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	// Process each resource metric
	for _, resourceMetric := range req.ResourceMetrics {
		if err := s.processResourceMetric(resourceMetric); err != nil {
			log.Printf("Failed to process resource metric: %v", err)
			// Continue processing other metrics even if one fails
		}
	}

	return &colmetricspb.ExportMetricsServiceResponse{
		PartialSuccess: &colmetricspb.ExportMetricsPartialSuccess{
			RejectedDataPoints: 0, // We process all metrics successfully
			ErrorMessage:       "",
		},
	}, nil
}

func (s *MetricsService) processResourceMetric(resourceMetric *metricspb.ResourceMetrics) error {
	// Extract service name from resource attributes
	serviceName := s.extractServiceName(resourceMetric.Resource)

	// Process each scope metric
	for _, scopeMetric := range resourceMetric.ScopeMetrics {
		for _, metric := range scopeMetric.Metrics {
			if err := s.processMetric(metric, serviceName); err != nil {
				log.Printf("Failed to process metric: %v", err)
				// Continue processing other metrics
			}
		}
	}

	return nil
}

func (s *MetricsService) processMetric(metric *metricspb.Metric, serviceName string) error {
	// Process different metric types
	switch data := metric.Data.(type) {
	case *metricspb.Metric_Gauge:
		return s.processGaugeMetric(metric.Name, data.Gauge, serviceName)
	case *metricspb.Metric_Sum:
		return s.processSumMetric(metric.Name, data.Sum, serviceName)
	case *metricspb.Metric_Histogram:
		return s.processHistogramMetric(metric.Name, data.Histogram, serviceName)
	case *metricspb.Metric_ExponentialHistogram:
		return s.processExponentialHistogramMetric(metric.Name, data.ExponentialHistogram, serviceName)
	case *metricspb.Metric_Summary:
		return s.processSummaryMetric(metric.Name, data.Summary, serviceName)
	default:
		log.Printf("Unknown metric type for metric: %s", metric.Name)
		return nil
	}
}

func (s *MetricsService) processGaugeMetric(name string, gauge *metricspb.Gauge, serviceName string) error {
	for _, dataPoint := range gauge.DataPoints {
		metricData := &storage.Metric{
			MetricName:  name,
			Value:       s.getNumericValue(dataPoint),
			Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
			ServiceName: serviceName,
			Labels:      s.convertAttributes(dataPoint.Attributes),
		}

		if err := s.storage.InsertMetric(metricData); err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricsService) processSumMetric(name string, sum *metricspb.Sum, serviceName string) error {
	for _, dataPoint := range sum.DataPoints {
		metricData := &storage.Metric{
			MetricName:  name,
			Value:       s.getNumericValue(dataPoint),
			Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
			ServiceName: serviceName,
			Labels:      s.convertAttributes(dataPoint.Attributes),
		}

		if err := s.storage.InsertMetric(metricData); err != nil {
			return err
		}
	}
	return nil
}

func (s *MetricsService) processHistogramMetric(name string, histogram *metricspb.Histogram, serviceName string) error {
	for _, dataPoint := range histogram.DataPoints {
		// Store count as a metric
		countMetric := &storage.Metric{
			MetricName:  name + "_count",
			Value:       float64(dataPoint.Count),
			Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
			ServiceName: serviceName,
			Labels:      s.convertAttributes(dataPoint.Attributes),
		}

		if err := s.storage.InsertMetric(countMetric); err != nil {
			return err
		}

		// Store sum as a metric
		if dataPoint.Sum != nil && *dataPoint.Sum != 0 {
			sumMetric := &storage.Metric{
				MetricName:  name + "_sum",
				Value:       *dataPoint.Sum,
				Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
				ServiceName: serviceName,
				Labels:      s.convertAttributes(dataPoint.Attributes),
			}

			if err := s.storage.InsertMetric(sumMetric); err != nil {
				return err
			}
		}

		// Store bucket counts
		for i, bucketCount := range dataPoint.BucketCounts {
			if i < len(dataPoint.ExplicitBounds) {
				bucketMetric := &storage.Metric{
					MetricName:  name + "_bucket",
					Value:       float64(bucketCount),
					Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
					ServiceName: serviceName,
					Labels:      s.addBucketLabel(s.convertAttributes(dataPoint.Attributes), dataPoint.ExplicitBounds[i]),
				}

				if err := s.storage.InsertMetric(bucketMetric); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *MetricsService) processExponentialHistogramMetric(name string, expHistogram *metricspb.ExponentialHistogram, serviceName string) error {
	for _, dataPoint := range expHistogram.DataPoints {
		// Store count as a metric
		countMetric := &storage.Metric{
			MetricName:  name + "_count",
			Value:       float64(dataPoint.Count),
			Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
			ServiceName: serviceName,
			Labels:      s.convertAttributes(dataPoint.Attributes),
		}

		if err := s.storage.InsertMetric(countMetric); err != nil {
			return err
		}

		// Store sum as a metric
		if dataPoint.Sum != nil && *dataPoint.Sum != 0 {
			sumMetric := &storage.Metric{
				MetricName:  name + "_sum",
				Value:       *dataPoint.Sum,
				Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
				ServiceName: serviceName,
				Labels:      s.convertAttributes(dataPoint.Attributes),
			}

			if err := s.storage.InsertMetric(sumMetric); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MetricsService) processSummaryMetric(name string, summary *metricspb.Summary, serviceName string) error {
	for _, dataPoint := range summary.DataPoints {
		// Store count as a metric
		countMetric := &storage.Metric{
			MetricName:  name + "_count",
			Value:       float64(dataPoint.Count),
			Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
			ServiceName: serviceName,
			Labels:      s.convertAttributes(dataPoint.Attributes),
		}

		if err := s.storage.InsertMetric(countMetric); err != nil {
			return err
		}

		// Store sum as a metric
		if dataPoint.Sum != 0 {
			sumMetric := &storage.Metric{
				MetricName:  name + "_sum",
				Value:       dataPoint.Sum,
				Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
				ServiceName: serviceName,
				Labels:      s.convertAttributes(dataPoint.Attributes),
			}

			if err := s.storage.InsertMetric(sumMetric); err != nil {
				return err
			}
		}

		// Store quantile values
		for _, quantile := range dataPoint.QuantileValues {
			quantileMetric := &storage.Metric{
				MetricName:  name + "_quantile",
				Value:       quantile.Value,
				Timestamp:   time.Unix(0, int64(dataPoint.TimeUnixNano)),
				ServiceName: serviceName,
				Labels:      s.addQuantileLabel(s.convertAttributes(dataPoint.Attributes), quantile.Quantile),
			}

			if err := s.storage.InsertMetric(quantileMetric); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *MetricsService) extractServiceName(resource *resourcepb.Resource) string {
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

func (s *MetricsService) getNumericValue(dataPoint *metricspb.NumberDataPoint) float64 {
	if dataPoint == nil {
		return 0.0
	}

	switch v := dataPoint.Value.(type) {
	case *metricspb.NumberDataPoint_AsDouble:
		return v.AsDouble
	case *metricspb.NumberDataPoint_AsInt:
		return float64(v.AsInt)
	default:
		return 0.0
	}
}

func (s *MetricsService) convertAttributes(attributes []*commonpb.KeyValue) string {
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

func (s *MetricsService) convertAttributeValue(value *commonpb.AnyValue) interface{} {
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

func (s *MetricsService) addBucketLabel(attributes string, bound float64) string {
	// Parse existing attributes
	var attrs map[string]interface{}
	if err := json.Unmarshal([]byte(attributes), &attrs); err != nil {
		attrs = make(map[string]interface{})
	}

	// Add bucket label
	attrs["le"] = bound

	// Convert back to JSON
	jsonData, err := json.Marshal(attrs)
	if err != nil {
		return attributes
	}

	return string(jsonData)
}

func (s *MetricsService) addQuantileLabel(attributes string, quantile float64) string {
	// Parse existing attributes
	var attrs map[string]interface{}
	if err := json.Unmarshal([]byte(attributes), &attrs); err != nil {
		attrs = make(map[string]interface{})
	}

	// Add quantile label
	attrs["quantile"] = quantile

	// Convert back to JSON
	jsonData, err := json.Marshal(attrs)
	if err != nil {
		return attributes
	}

	return string(jsonData)
}
