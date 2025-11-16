package promql

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Query represents a parsed PromQL query
type Query struct {
	MetricName  string
	Labels      map[string]string
	Function    string
	Range       time.Duration
	Offset      time.Duration
	Aggregation *Aggregation
}

// Aggregation represents aggregation operations
type Aggregation struct {
	Operation string   // sum, avg, count, min, max
	By        []string // grouping labels
	Without   []string // excluding labels
}

// Parser handles PromQL query parsing
type Parser struct{}

// NewParser creates a new PromQL parser
func NewParser() *Parser {
	return &Parser{}
}

// Parse parses a PromQL query string into a Query struct
func (p *Parser) Parse(query string) (*Query, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	// Handle function calls (e.g., rate(http_requests_total[5m]))
	if strings.Contains(query, "(") && strings.Contains(query, ")") {
		return p.parseFunction(query)
	}

	// Handle simple metric queries (e.g., http_requests_total)
	if !strings.Contains(query, "{") {
		return &Query{
			MetricName: query,
			Labels:     make(map[string]string),
		}, nil
	}

	// Handle metric with labels (e.g., http_requests_total{service="api"})
	return p.parseMetricWithLabels(query)
}

// parseFunction handles function calls like rate(http_requests_total[5m])
func (p *Parser) parseFunction(query string) (*Query, error) {
	// Find function name
	openParen := strings.Index(query, "(")
	if openParen == -1 {
		return nil, fmt.Errorf("invalid function syntax")
	}

	funcName := strings.TrimSpace(query[:openParen])

	// Find closing parenthesis
	closeParen := strings.LastIndex(query, ")")
	if closeParen == -1 || closeParen <= openParen {
		return nil, fmt.Errorf("missing closing parenthesis")
	}

	// Extract function argument
	arg := strings.TrimSpace(query[openParen+1 : closeParen])

	// Parse the argument (could be a metric with range)
	metricQuery, rangeDuration, err := p.parseMetricWithRange(arg)
	if err != nil {
		return nil, fmt.Errorf("invalid function argument: %w", err)
	}

	return &Query{
		MetricName: metricQuery.MetricName,
		Labels:     metricQuery.Labels,
		Function:   funcName,
		Range:      rangeDuration,
	}, nil
}

// parseMetricWithRange handles metrics with time ranges like http_requests_total[5m]
func (p *Parser) parseMetricWithRange(query string) (*Query, time.Duration, error) {
	// Check for range selector [duration]
	rangeStart := strings.Index(query, "[")
	if rangeStart == -1 {
		// No range, parse as regular metric
		q, err := p.parseMetricWithLabels(query)
		return q, 0, err
	}

	rangeEnd := strings.Index(query, "]")
	if rangeEnd == -1 {
		return nil, 0, fmt.Errorf("missing closing bracket in range selector")
	}

	metricPart := strings.TrimSpace(query[:rangeStart])
	rangePart := strings.TrimSpace(query[rangeStart+1 : rangeEnd])

	// Parse the metric part
	metricQuery, err := p.parseMetricWithLabels(metricPart)
	if err != nil {
		return nil, 0, err
	}

	// Parse the range duration
	duration, err := p.parseDuration(rangePart)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid range duration: %w", err)
	}

	return metricQuery, duration, nil
}

// parseMetricWithLabels handles metrics with label selectors
func (p *Parser) parseMetricWithLabels(query string) (*Query, error) {
	// Find label selector
	labelStart := strings.Index(query, "{")
	if labelStart == -1 {
		// No labels, just metric name
		return &Query{
			MetricName: strings.TrimSpace(query),
			Labels:     make(map[string]string),
		}, nil
	}

	labelEnd := strings.Index(query, "}")
	if labelEnd == -1 {
		return nil, fmt.Errorf("missing closing brace in label selector")
	}

	metricName := strings.TrimSpace(query[:labelStart])
	labelSelector := strings.TrimSpace(query[labelStart+1 : labelEnd])

	// Parse labels
	labels, err := p.parseLabels(labelSelector)
	if err != nil {
		return nil, err
	}

	return &Query{
		MetricName: metricName,
		Labels:     labels,
	}, nil
}

// parseLabels parses label selectors like service="api",method="GET"
func (p *Parser) parseLabels(selector string) (map[string]string, error) {
	labels := make(map[string]string)

	if selector == "" {
		return labels, nil
	}

	// Split by comma
	pairs := strings.Split(selector, ",")
	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		// Split by equals sign
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label selector: %s", pair)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}

		labels[key] = value
	}

	return labels, nil
}

// parseDuration parses duration strings like "5m", "1h", "30s"
func (p *Parser) parseDuration(duration string) (time.Duration, error) {
	duration = strings.TrimSpace(duration)
	if duration == "" {
		return 0, fmt.Errorf("empty duration")
	}

	// Handle common duration formats
	switch {
	case strings.HasSuffix(duration, "s"):
		val, err := strconv.Atoi(duration[:len(duration)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Second, nil
	case strings.HasSuffix(duration, "m"):
		val, err := strconv.Atoi(duration[:len(duration)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Minute, nil
	case strings.HasSuffix(duration, "h"):
		val, err := strconv.Atoi(duration[:len(duration)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * time.Hour, nil
	case strings.HasSuffix(duration, "d"):
		val, err := strconv.Atoi(duration[:len(duration)-1])
		if err != nil {
			return 0, err
		}
		return time.Duration(val) * 24 * time.Hour, nil
	default:
		// Try parsing as Go duration
		return time.ParseDuration(duration)
	}
}

// ParseAggregation parses aggregation queries like sum(http_requests_total) by (service)
func (p *Parser) ParseAggregation(query string) (*Query, error) {
	query = strings.TrimSpace(query)

	// Find aggregation function
	openParen := strings.Index(query, "(")
	if openParen == -1 {
		return nil, fmt.Errorf("invalid aggregation syntax")
	}

	funcName := strings.TrimSpace(query[:openParen])

	// Find closing parenthesis
	closeParen := strings.LastIndex(query, ")")
	if closeParen == -1 {
		return nil, fmt.Errorf("missing closing parenthesis")
	}

	// Extract the metric query
	metricQuery := strings.TrimSpace(query[openParen+1 : closeParen])

	// Parse the metric
	parsedQuery, err := p.Parse(metricQuery)
	if err != nil {
		return nil, err
	}

	// Check for "by" clause
	byClause := ""
	if closeParen < len(query)-1 {
		remaining := strings.TrimSpace(query[closeParen+1:])
		if strings.HasPrefix(remaining, "by") {
			byStart := strings.Index(remaining, "(")
			byEnd := strings.Index(remaining, ")")
			if byStart != -1 && byEnd != -1 {
				byClause = strings.TrimSpace(remaining[byStart+1 : byEnd])
			}
		}
	}

	// Parse grouping labels
	var byLabels []string
	if byClause != "" {
		byLabels = strings.Split(byClause, ",")
		for i, label := range byLabels {
			byLabels[i] = strings.TrimSpace(label)
		}
	}

	parsedQuery.Aggregation = &Aggregation{
		Operation: funcName,
		By:        byLabels,
	}

	return parsedQuery, nil
}
