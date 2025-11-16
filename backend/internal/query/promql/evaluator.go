package promql

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"sort"
	"time"
)

// MetricPoint represents a single data point
type MetricPoint struct {
	Timestamp time.Time
	Value     float64
	Labels    map[string]string
}

// MetricSeries represents a time series of metric points
type MetricSeries struct {
	MetricName string
	Labels     map[string]string
	Points     []MetricPoint
}

// QueryResult represents the result of a PromQL query
type QueryResult struct {
	Series []MetricSeries
	Type   string // "vector", "matrix", "scalar"
}

// Evaluator handles PromQL query evaluation
type Evaluator struct {
	db *sql.DB
}

// NewEvaluator creates a new PromQL evaluator
func NewEvaluator(db *sql.DB) *Evaluator {
	return &Evaluator{db: db}
}

// Evaluate executes a parsed PromQL query
func (e *Evaluator) Evaluate(ctx context.Context, query *Query, startTime, endTime time.Time) (*QueryResult, error) {
	// Get base metric data
	series, err := e.getMetricSeries(ctx, query, startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to get metric series: %w", err)
	}

	// Apply function if specified
	if query.Function != "" {
		series, err = e.applyFunction(series, query.Function, query.Range)
		if err != nil {
			return nil, fmt.Errorf("failed to apply function %s: %w", query.Function, err)
		}
	}

	// Apply aggregation if specified
	if query.Aggregation != nil {
		series, err = e.applyAggregation(series, query.Aggregation)
		if err != nil {
			return nil, fmt.Errorf("failed to apply aggregation: %w", err)
		}
	}

	return &QueryResult{
		Series: series,
		Type:   "vector",
	}, nil
}

// getMetricSeries retrieves metric data from the database
func (e *Evaluator) getMetricSeries(ctx context.Context, query *Query, startTime, endTime time.Time) ([]MetricSeries, error) {
	// Build SQL query
	sqlQuery := `
		SELECT timestamp, value, labels, service_name
		FROM metrics 
		WHERE metric_name = ? 
		AND timestamp >= ? 
		AND timestamp <= ?
	`

	args := []interface{}{query.MetricName, startTime.Unix(), endTime.Unix()}

	// Add label filters
	for key, value := range query.Labels {
		sqlQuery += fmt.Sprintf(" AND JSON_EXTRACT(labels, '$.%s') = ?", key)
		args = append(args, value)
	}

	sqlQuery += " ORDER BY timestamp ASC"

	rows, err := e.db.QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("database query failed: %w", err)
	}
	defer rows.Close()

	// Group by labels to create series
	seriesMap := make(map[string]*MetricSeries)

	for rows.Next() {
		var timestamp int64
		var value float64
		var labelsJSON string
		var serviceName string

		if err := rows.Scan(&timestamp, &value, &labelsJSON, &serviceName); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Parse labels JSON (simplified - in real implementation, use proper JSON parsing)
		labels := map[string]string{
			"service": serviceName,
		}

		// Create series key for grouping
		seriesKey := e.createSeriesKey(labels)

		// Get or create series
		series, exists := seriesMap[seriesKey]
		if !exists {
			series = &MetricSeries{
				MetricName: query.MetricName,
				Labels:     labels,
				Points:     []MetricPoint{},
			}
			seriesMap[seriesKey] = series
		}

		// Add point to series
		series.Points = append(series.Points, MetricPoint{
			Timestamp: time.Unix(timestamp, 0),
			Value:     value,
			Labels:    labels,
		})
	}

	// Convert map to slice
	var result []MetricSeries
	for _, series := range seriesMap {
		result = append(result, *series)
	}

	return result, nil
}

// createSeriesKey creates a unique key for grouping series by labels
func (e *Evaluator) createSeriesKey(labels map[string]string) string {
	var keys []string
	for k, v := range labels {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(keys)
	return fmt.Sprintf("%v", keys)
}

// applyFunction applies PromQL functions to the series
func (e *Evaluator) applyFunction(series []MetricSeries, function string, rangeDuration time.Duration) ([]MetricSeries, error) {
	switch function {
	case "rate":
		return e.applyRate(series, rangeDuration)
	case "increase":
		return e.applyIncrease(series, rangeDuration)
	case "sum":
		return e.applySum(series)
	case "avg":
		return e.applyAvg(series)
	case "count":
		return e.applyCount(series)
	case "min":
		return e.applyMin(series)
	case "max":
		return e.applyMax(series)
	default:
		return nil, fmt.Errorf("unsupported function: %s", function)
	}
}

// applyRate calculates the per-second rate of increase
func (e *Evaluator) applyRate(series []MetricSeries, rangeDuration time.Duration) ([]MetricSeries, error) {
	result := make([]MetricSeries, len(series))

	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     []MetricPoint{},
		}

		// Calculate rate for each point
		for j, point := range s.Points {
			if j == 0 {
				continue // Skip first point
			}

			// Find previous point within range
			rangeStart := point.Timestamp.Add(-rangeDuration)
			var prevPoint *MetricPoint

			for k := j - 1; k >= 0; k-- {
				if s.Points[k].Timestamp.After(rangeStart) {
					prevPoint = &s.Points[k]
					break
				}
			}

			if prevPoint != nil {
				// Calculate rate
				timeDiff := point.Timestamp.Sub(prevPoint.Timestamp).Seconds()
				valueDiff := point.Value - prevPoint.Value
				rate := valueDiff / timeDiff

				result[i].Points = append(result[i].Points, MetricPoint{
					Timestamp: point.Timestamp,
					Value:     rate,
					Labels:    point.Labels,
				})
			}
		}
	}

	return result, nil
}

// applyIncrease calculates the increase over the range
func (e *Evaluator) applyIncrease(series []MetricSeries, rangeDuration time.Duration) ([]MetricSeries, error) {
	result := make([]MetricSeries, len(series))

	for i, s := range series {
		result[i] = MetricSeries{
			MetricName: s.MetricName,
			Labels:     s.Labels,
			Points:     []MetricPoint{},
		}

		// Calculate increase for each point
		for j, point := range s.Points {
			rangeStart := point.Timestamp.Add(-rangeDuration)
			var startPoint *MetricPoint

			// Find start point within range
			for k := j; k >= 0; k-- {
				if s.Points[k].Timestamp.Before(rangeStart) || s.Points[k].Timestamp.Equal(rangeStart) {
					startPoint = &s.Points[k]
					break
				}
			}

			if startPoint != nil {
				increase := point.Value - startPoint.Value
				result[i].Points = append(result[i].Points, MetricPoint{
					Timestamp: point.Timestamp,
					Value:     increase,
					Labels:    point.Labels,
				})
			}
		}
	}

	return result, nil
}

// applySum sums all series values
func (e *Evaluator) applySum(series []MetricSeries) ([]MetricSeries, error) {
	if len(series) == 0 {
		return series, nil
	}

	// Get all unique timestamps
	timestampMap := make(map[time.Time]bool)
	for _, s := range series {
		for _, point := range s.Points {
			timestampMap[point.Timestamp] = true
		}
	}

	var timestamps []time.Time
	for ts := range timestampMap {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	// Calculate sum for each timestamp
	var points []MetricPoint
	for _, ts := range timestamps {
		sum := 0.0
		for _, s := range series {
			for _, point := range s.Points {
				if point.Timestamp.Equal(ts) {
					sum += point.Value
					break
				}
			}
		}

		points = append(points, MetricPoint{
			Timestamp: ts,
			Value:     sum,
			Labels:    map[string]string{},
		})
	}

	return []MetricSeries{{
		MetricName: series[0].MetricName,
		Labels:     map[string]string{},
		Points:     points,
	}}, nil
}

// applyAvg calculates the average of all series values
func (e *Evaluator) applyAvg(series []MetricSeries) ([]MetricSeries, error) {
	if len(series) == 0 {
		return series, nil
	}

	// Get all unique timestamps
	timestampMap := make(map[time.Time]bool)
	for _, s := range series {
		for _, point := range s.Points {
			timestampMap[point.Timestamp] = true
		}
	}

	var timestamps []time.Time
	for ts := range timestampMap {
		timestamps = append(timestamps, ts)
	}
	sort.Slice(timestamps, func(i, j int) bool {
		return timestamps[i].Before(timestamps[j])
	})

	// Calculate average for each timestamp
	var points []MetricPoint
	for _, ts := range timestamps {
		sum := 0.0
		count := 0
		for _, s := range series {
			for _, point := range s.Points {
				if point.Timestamp.Equal(ts) {
					sum += point.Value
					count++
					break
				}
			}
		}

		avg := 0.0
		if count > 0 {
			avg = sum / float64(count)
		}

		points = append(points, MetricPoint{
			Timestamp: ts,
			Value:     avg,
			Labels:    map[string]string{},
		})
	}

	return []MetricSeries{{
		MetricName: series[0].MetricName,
		Labels:     map[string]string{},
		Points:     points,
	}}, nil
}

// applyCount counts the number of series
func (e *Evaluator) applyCount(series []MetricSeries) ([]MetricSeries, error) {
	count := float64(len(series))
	return []MetricSeries{{
		MetricName: "count",
		Labels:     map[string]string{},
		Points:     []MetricPoint{{Timestamp: time.Now(), Value: count}},
	}}, nil
}

// applyMin finds the minimum value across all series
func (e *Evaluator) applyMin(series []MetricSeries) ([]MetricSeries, error) {
	if len(series) == 0 {
		return series, nil
	}

	min := math.Inf(1)
	for _, s := range series {
		for _, point := range s.Points {
			if point.Value < min {
				min = point.Value
			}
		}
	}

	return []MetricSeries{{
		MetricName: series[0].MetricName,
		Labels:     map[string]string{},
		Points:     []MetricPoint{{Timestamp: time.Now(), Value: min}},
	}}, nil
}

// applyMax finds the maximum value across all series
func (e *Evaluator) applyMax(series []MetricSeries) ([]MetricSeries, error) {
	if len(series) == 0 {
		return series, nil
	}

	max := math.Inf(-1)
	for _, s := range series {
		for _, point := range s.Points {
			if point.Value > max {
				max = point.Value
			}
		}
	}

	return []MetricSeries{{
		MetricName: series[0].MetricName,
		Labels:     map[string]string{},
		Points:     []MetricPoint{{Timestamp: time.Now(), Value: max}},
	}}, nil
}

// applyAggregation applies aggregation operations
func (e *Evaluator) applyAggregation(series []MetricSeries, agg *Aggregation) ([]MetricSeries, error) {
	switch agg.Operation {
	case "sum":
		return e.applySum(series)
	case "avg":
		return e.applyAvg(series)
	case "count":
		return e.applyCount(series)
	case "min":
		return e.applyMin(series)
	case "max":
		return e.applyMax(series)
	default:
		return nil, fmt.Errorf("unsupported aggregation: %s", agg.Operation)
	}
}

