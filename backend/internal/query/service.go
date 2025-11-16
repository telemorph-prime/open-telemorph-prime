package query

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"open-telemorph-prime/internal/query/promql"

	"github.com/gin-gonic/gin"
)

// Service handles query operations
type Service struct {
	db           *sql.DB
	promqlParser *promql.Parser
	promqlEval   *promql.Evaluator
}

// NewService creates a new query service
func NewService(db *sql.DB) *Service {
	return &Service{
		db:           db,
		promqlParser: promql.NewParser(),
		promqlEval:   promql.NewEvaluator(db),
	}
}

// QueryRequest represents a query request
type QueryRequest struct {
	Query     string    `json:"query" binding:"required"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Step      string    `json:"step,omitempty"`
}

// QueryResponse represents a query response
type QueryResponse struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

// RegisterRoutes registers query API routes
func (s *Service) RegisterRoutes(router *gin.RouterGroup) {
	query := router.Group("/query")
	{
		query.POST("/metrics", s.HandleMetricsQuery)
		query.POST("/logs", s.HandleLogsQuery)
		query.POST("/traces", s.HandleTracesQuery)
		query.GET("/export", s.HandleExport)
	}
}

// HandleMetricsQuery handles PromQL metrics queries
func (s *Service) HandleMetricsQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, QueryResponse{
			Status: "error",
			Error:  fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Set default time range if not provided
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().Add(-1 * time.Hour)
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}

	// Parse PromQL query
	query, err := s.promqlParser.Parse(req.Query)
	if err != nil {
		c.JSON(http.StatusBadRequest, QueryResponse{
			Status: "error",
			Error:  fmt.Sprintf("Invalid PromQL query: %v", err),
		})
		return
	}

	// Evaluate query
	result, err := s.promqlEval.Evaluate(context.Background(), query, req.StartTime, req.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, QueryResponse{
			Status: "error",
			Error:  fmt.Sprintf("Query evaluation failed: %v", err),
		})
		return
	}

	// Convert result to Prometheus format
	promResult := s.convertToPrometheusFormat(result)

	c.JSON(http.StatusOK, QueryResponse{
		Status: "success",
		Data:   promResult,
	})
}

// HandleLogsQuery handles log queries (placeholder)
func (s *Service) HandleLogsQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, QueryResponse{
			Status: "error",
			Error:  fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// TODO: Implement log query parsing and evaluation
	c.JSON(http.StatusOK, QueryResponse{
		Status: "success",
		Data: map[string]interface{}{
			"message": "Log queries not yet implemented",
			"query":   req.Query,
		},
	})
}

// HandleTracesQuery handles trace queries (placeholder)
func (s *Service) HandleTracesQuery(c *gin.Context) {
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, QueryResponse{
			Status: "error",
			Error:  fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// TODO: Implement trace query parsing and evaluation
	c.JSON(http.StatusOK, QueryResponse{
		Status: "success",
		Data: map[string]interface{}{
			"message": "Trace queries not yet implemented",
			"query":   req.Query,
		},
	})
}

// HandleExport handles data export requests
func (s *Service) HandleExport(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	// TODO: Implement export functionality
	c.JSON(http.StatusOK, QueryResponse{
		Status: "success",
		Data: map[string]interface{}{
			"message": "Export functionality not yet implemented",
			"format":  format,
		},
	})
}

// convertToPrometheusFormat converts internal result to Prometheus API format
func (s *Service) convertToPrometheusFormat(result *promql.QueryResult) map[string]interface{} {
	var data []map[string]interface{}

	for _, series := range result.Series {
		// Convert points to Prometheus format
		var values [][]interface{}
		for _, point := range series.Points {
			values = append(values, []interface{}{
				float64(point.Timestamp.Unix()),
				point.Value,
			})
		}

		// Create metric object
		metric := map[string]interface{}{
			"__name__": series.MetricName,
		}
		for k, v := range series.Labels {
			metric[k] = v
		}

		data = append(data, map[string]interface{}{
			"metric": metric,
			"values": values,
		})
	}

	return map[string]interface{}{
		"resultType": result.Type,
		"result":     data,
	}
}

// GetAvailableMetrics returns a list of available metrics
func (s *Service) GetAvailableMetrics(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT DISTINCT metric_name FROM metrics ORDER BY metric_name")
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var metrics []string
	for rows.Next() {
		var metricName string
		if err := rows.Scan(&metricName); err != nil {
			return nil, fmt.Errorf("failed to scan metric name: %w", err)
		}
		metrics = append(metrics, metricName)
	}

	return metrics, nil
}

// GetMetricLabels returns available labels for a metric
func (s *Service) GetMetricLabels(ctx context.Context, metricName string) (map[string][]string, error) {
	rows, err := s.db.QueryContext(ctx,
		"SELECT DISTINCT labels FROM metrics WHERE metric_name = ? AND labels IS NOT NULL",
		metricName)
	if err != nil {
		return nil, fmt.Errorf("failed to query metric labels: %w", err)
	}
	defer rows.Close()

	labelValues := make(map[string]map[string]bool)

	for rows.Next() {
		var labelsJSON string
		if err := rows.Scan(&labelsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan labels: %w", err)
		}

		// Parse labels JSON (simplified)
		var labels map[string]string
		if err := json.Unmarshal([]byte(labelsJSON), &labels); err != nil {
			continue // Skip invalid JSON
		}

		for key, value := range labels {
			if labelValues[key] == nil {
				labelValues[key] = make(map[string]bool)
			}
			labelValues[key][value] = true
		}
	}

	// Convert to slice format
	result := make(map[string][]string)
	for key, values := range labelValues {
		var valueSlice []string
		for value := range values {
			valueSlice = append(valueSlice, value)
		}
		result[key] = valueSlice
	}

	return result, nil
}

