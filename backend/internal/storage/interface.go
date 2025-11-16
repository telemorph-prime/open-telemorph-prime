package storage

import "database/sql"

// Storage interface defines the contract for data storage
type Storage interface {
	// Metrics
	InsertMetric(metric *Metric) error
	GetMetrics(limit int, offset int) ([]*Metric, error)

	// Traces
	InsertTrace(trace *Trace) error
	GetTraces(limit int, offset int) ([]*Trace, error)

	// Logs
	InsertLog(log *Log) error
	GetLogs(limit int, offset int) ([]*Log, error)

	// Services
	GetServices() ([]string, error)

	// Cleanup
	CleanupOldData() error
	Close() error

	// System info
	GetDatabasePath() string

	// Database access for query service
	GetDB() *sql.DB
}
