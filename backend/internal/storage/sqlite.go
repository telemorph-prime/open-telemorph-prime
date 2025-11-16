package storage

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"open-telemorph-prime/internal/config"

	_ "modernc.org/sqlite"
)

type SQLiteStorage struct {
	db     *sql.DB
	config config.StorageConfig
}

type Metric struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	MetricName  string    `json:"metric_name"`
	Value       float64   `json:"value"`
	Labels      string    `json:"labels"` // JSON string
	ServiceName string    `json:"service_name"`
	CreatedAt   time.Time `json:"created_at"`
}

type Trace struct {
	ID            int64     `json:"id"`
	TraceID       string    `json:"trace_id"`
	SpanID        string    `json:"span_id"`
	ParentSpanID  *string   `json:"parent_span_id"`
	ServiceName   string    `json:"service_name"`
	OperationName string    `json:"operation_name"`
	StartTime     time.Time `json:"start_time"`
	DurationNanos int64     `json:"duration_nanos"`
	Attributes    string    `json:"attributes"` // JSON string
	StatusCode    string    `json:"status_code"`
	CreatedAt     time.Time `json:"created_at"`
}

type Log struct {
	ID          int64     `json:"id"`
	Timestamp   time.Time `json:"timestamp"`
	ServiceName string    `json:"service_name"`
	Level       string    `json:"level"`
	Message     string    `json:"message"`
	Attributes  string    `json:"attributes"` // JSON string
	TraceID     *string   `json:"trace_id"`
	SpanID      *string   `json:"span_id"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewSQLiteStorage(cfg config.StorageConfig) (*SQLiteStorage, error) {
	// Create data directory if it doesn't exist
	if err := createDataDir(cfg.Path); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite", cfg.Path+"?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{
		db:     db,
		config: cfg,
	}

	// Create tables
	if err := storage.createTables(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return storage, nil
}

func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

func (s *SQLiteStorage) createTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp INTEGER NOT NULL,
			metric_name TEXT NOT NULL,
			value REAL NOT NULL,
			labels TEXT,
			service_name TEXT,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS traces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trace_id TEXT NOT NULL,
			span_id TEXT NOT NULL,
			parent_span_id TEXT,
			service_name TEXT,
			operation_name TEXT,
			start_time INTEGER NOT NULL,
			duration_nanos INTEGER NOT NULL,
			attributes TEXT,
			status_code TEXT,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		)`,
		`CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp INTEGER NOT NULL,
			service_name TEXT,
			level TEXT,
			message TEXT,
			attributes TEXT,
			trace_id TEXT,
			span_id TEXT,
			created_at INTEGER DEFAULT (strftime('%s', 'now'))
		)`,
		// Indexes for performance
		`CREATE INDEX IF NOT EXISTS idx_metrics_timestamp ON metrics(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_service ON metrics(service_name)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(metric_name)`,
		`CREATE INDEX IF NOT EXISTS idx_traces_trace_id ON traces(trace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_traces_service ON traces(service_name)`,
		`CREATE INDEX IF NOT EXISTS idx_traces_start_time ON traces(start_time)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_service ON logs(service_name)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query %s: %w", query, err)
		}
	}

	return nil
}

func createDataDir(path string) error {
	// Extract directory from path
	dir := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			dir = path[:i]
			break
		}
	}

	if dir == "" {
		return nil // No directory to create
	}

	return os.MkdirAll(dir, 0755)
}

// Metric methods
func (s *SQLiteStorage) InsertMetric(metric *Metric) error {
	query := `INSERT INTO metrics (timestamp, metric_name, value, labels, service_name) 
			  VALUES (?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		metric.Timestamp.UnixNano(),
		metric.MetricName,
		metric.Value,
		metric.Labels,
		metric.ServiceName,
	)
	return err
}

func (s *SQLiteStorage) GetMetrics(limit int, offset int) ([]*Metric, error) {
	query := `SELECT id, timestamp, metric_name, value, labels, service_name, created_at 
			  FROM metrics 
			  ORDER BY timestamp DESC 
			  LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*Metric
	for rows.Next() {
		var m Metric
		var timestamp, createdAt int64

		err := rows.Scan(&m.ID, &timestamp, &m.MetricName, &m.Value, &m.Labels, &m.ServiceName, &createdAt)
		if err != nil {
			return nil, err
		}

		m.Timestamp = time.Unix(0, timestamp)
		m.CreatedAt = time.Unix(createdAt, 0)
		metrics = append(metrics, &m)
	}

	return metrics, nil
}

// Trace methods
func (s *SQLiteStorage) InsertTrace(trace *Trace) error {
	query := `INSERT INTO traces (trace_id, span_id, parent_span_id, service_name, operation_name, 
			  start_time, duration_nanos, attributes, status_code) 
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		trace.TraceID,
		trace.SpanID,
		trace.ParentSpanID,
		trace.ServiceName,
		trace.OperationName,
		trace.StartTime.UnixNano(),
		trace.DurationNanos,
		trace.Attributes,
		trace.StatusCode,
	)
	return err
}

func (s *SQLiteStorage) GetTraces(limit int, offset int) ([]*Trace, error) {
	query := `SELECT id, trace_id, span_id, parent_span_id, service_name, operation_name, 
			  start_time, duration_nanos, attributes, status_code, created_at 
			  FROM traces 
			  ORDER BY start_time DESC 
			  LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var traces []*Trace
	for rows.Next() {
		var t Trace
		var startTime, createdAt int64

		err := rows.Scan(&t.ID, &t.TraceID, &t.SpanID, &t.ParentSpanID, &t.ServiceName,
			&t.OperationName, &startTime, &t.DurationNanos, &t.Attributes, &t.StatusCode, &createdAt)
		if err != nil {
			return nil, err
		}

		t.StartTime = time.Unix(0, startTime)
		t.CreatedAt = time.Unix(createdAt, 0)
		traces = append(traces, &t)
	}

	return traces, nil
}

// Log methods
func (s *SQLiteStorage) InsertLog(log *Log) error {
	query := `INSERT INTO logs (timestamp, service_name, level, message, attributes, trace_id, span_id) 
			  VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := s.db.Exec(query,
		log.Timestamp.UnixNano(),
		log.ServiceName,
		log.Level,
		log.Message,
		log.Attributes,
		log.TraceID,
		log.SpanID,
	)
	return err
}

func (s *SQLiteStorage) GetLogs(limit int, offset int) ([]*Log, error) {
	query := `SELECT id, timestamp, service_name, level, message, attributes, trace_id, span_id, created_at 
			  FROM logs 
			  ORDER BY timestamp DESC 
			  LIMIT ? OFFSET ?`

	rows, err := s.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*Log
	for rows.Next() {
		var l Log
		var timestamp, createdAt int64

		err := rows.Scan(&l.ID, &timestamp, &l.ServiceName, &l.Level, &l.Message,
			&l.Attributes, &l.TraceID, &l.SpanID, &createdAt)
		if err != nil {
			return nil, err
		}

		l.Timestamp = time.Unix(0, timestamp)
		l.CreatedAt = time.Unix(createdAt, 0)
		logs = append(logs, &l)
	}

	return logs, nil
}

// Service methods
func (s *SQLiteStorage) GetServices() ([]string, error) {
	query := `SELECT DISTINCT service_name FROM (
		SELECT service_name FROM metrics WHERE service_name IS NOT NULL AND service_name != ''
		UNION
		SELECT service_name FROM traces WHERE service_name IS NOT NULL AND service_name != ''
		UNION
		SELECT service_name FROM logs WHERE service_name IS NOT NULL AND service_name != ''
	) ORDER BY service_name`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []string
	for rows.Next() {
		var service string
		if err := rows.Scan(&service); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, nil
}

// Cleanup old data
func (s *SQLiteStorage) CleanupOldData() error {
	cutoff := time.Now().AddDate(0, 0, -s.config.RetentionDays).UnixNano()

	queries := []string{
		`DELETE FROM metrics WHERE timestamp < ?`,
		`DELETE FROM traces WHERE start_time < ?`,
		`DELETE FROM logs WHERE timestamp < ?`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query, cutoff); err != nil {
			return fmt.Errorf("failed to cleanup old data: %w", err)
		}
	}

	return nil
}

// GetDatabasePath returns the path to the database file
func (s *SQLiteStorage) GetDatabasePath() string {
	return s.config.Path
}

func (s *SQLiteStorage) GetDB() *sql.DB {
	return s.db
}
