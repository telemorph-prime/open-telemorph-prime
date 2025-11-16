package web

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"open-telemorph-prime/internal/config"
	"open-telemorph-prime/internal/storage"

	"github.com/gin-gonic/gin"
)

type Service struct {
	storage   storage.Storage
	config    config.WebConfig
	version   string
	startTime time.Time
}

func NewService(storage storage.Storage, config config.WebConfig, version string) *Service {
	return &Service{
		storage:   storage,
		config:    config,
		version:   version,
		startTime: time.Now(),
	}
}

// API endpoints
func (s *Service) GetMetrics(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	metrics, err := s.storage.GetMetrics(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   metrics,
		"total":  len(metrics),
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Service) GetTraces(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	traces, err := s.storage.GetTraces(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   traces,
		"total":  len(traces),
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Service) GetLogs(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := s.storage.GetLogs(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":   logs,
		"total":  len(logs),
		"limit":  limit,
		"offset": offset,
	})
}

func (s *Service) GetServices(c *gin.Context) {
	services, err := s.storage.GetServices()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
	})
}

func (s *Service) Query(c *gin.Context) {
	var queryReq struct {
		Type      string `json:"type" binding:"required"`
		Query     string `json:"query" binding:"required"`
		Limit     int    `json:"limit"`
		Offset    int    `json:"offset"`
		TimeRange string `json:"timeRange"`
		Step      string `json:"step"`
	}

	if err := c.ShouldBindJSON(&queryReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if queryReq.Limit == 0 {
		queryReq.Limit = 100
	}

	// Convert to QueryRequest type
	req := QueryRequest{
		Type:      queryReq.Type,
		Query:     queryReq.Query,
		Limit:     queryReq.Limit,
		Offset:    queryReq.Offset,
		TimeRange: queryReq.TimeRange,
		Step:      queryReq.Step,
	}

	switch queryReq.Type {
	case "promql":
		// Route PromQL queries to the new query service
		s.handlePromQLQuery(c, req)
	case "logql":
		// Route LogQL queries to the new query service
		s.handleLogQLQuery(c, req)
	case "traceql":
		// Route TraceQL queries to the new query service
		s.handleTraceQLQuery(c, req)
	case "metrics":
		metrics, err := s.storage.GetMetrics(queryReq.Limit, queryReq.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": metrics})
	case "traces":
		traces, err := s.storage.GetTraces(queryReq.Limit, queryReq.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": traces})
	case "logs":
		logs, err := s.storage.GetLogs(queryReq.Limit, queryReq.Offset)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": logs})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query type"})
	}
}

// QueryRequest represents a query request
type QueryRequest struct {
	Type      string `json:"type"`
	Query     string `json:"query"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	TimeRange string `json:"timeRange"`
	Step      string `json:"step"`
}

// handlePromQLQuery handles PromQL queries by forwarding to the query service
func (s *Service) handlePromQLQuery(c *gin.Context, queryReq QueryRequest) {
	// For now, return a simple response indicating PromQL is not fully implemented
	// In a full implementation, this would forward to the query service
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"resultType": "vector",
			"result": []gin.H{
				{
					"metric": gin.H{
						"__name__": queryReq.Query,
					},
					"values": [][]interface{}{
						{float64(time.Now().Unix()), 0.0},
					},
				},
			},
		},
		"message": "PromQL query received - full implementation in progress",
	})
}

// handleLogQLQuery handles LogQL queries
func (s *Service) handleLogQLQuery(c *gin.Context, queryReq QueryRequest) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"message": "LogQL queries not yet implemented",
			"query":   queryReq.Query,
		},
	})
}

// handleTraceQLQuery handles TraceQL queries
func (s *Service) handleTraceQLQuery(c *gin.Context, queryReq QueryRequest) {
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data": gin.H{
			"message": "TraceQL queries not yet implemented",
			"query":   queryReq.Query,
		},
	})
}

// Web UI handlers
func (s *Service) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":   s.config.Title,
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":   s.config.Title + " - Dashboard",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) MetricsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "metrics.html", gin.H{
		"title":   s.config.Title + " - Metrics",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) TracesPage(c *gin.Context) {
	c.HTML(http.StatusOK, "traces.html", gin.H{
		"title":   s.config.Title + " - Traces",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) LogsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "logs.html", gin.H{
		"title":   s.config.Title + " - Logs",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) ServicesPage(c *gin.Context) {
	c.HTML(http.StatusOK, "services.html", gin.H{
		"title":   s.config.Title + " - Services",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) AlertsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "alerts.html", gin.H{
		"title":   s.config.Title + " - Alerts",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) QueryPage(c *gin.Context) {
	c.HTML(http.StatusOK, "query.html", gin.H{
		"title":   s.config.Title + " - Query Builder",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

func (s *Service) AdminPage(c *gin.Context) {
	c.HTML(http.StatusOK, "admin.html", gin.H{
		"title":   s.config.Title + " - Administration",
		"theme":   s.config.Theme,
		"version": s.version,
	})
}

// Admin API endpoints
func (s *Service) GetConfig(c *gin.Context) {
	// TODO: Implement config retrieval
	c.JSON(http.StatusOK, gin.H{
		"server": gin.H{
			"port":          8080,
			"read_timeout":  "5s",
			"write_timeout": "10s",
			"idle_timeout":  "120s",
		},
		"storage": gin.H{
			"type":           "sqlite",
			"path":           "./data/telemorph.db",
			"retention_days": 30,
		},
		"ingestion": gin.H{
			"api_endpoint":    "0.0.0.0:9013",
			"health_endpoint": "0.0.0.0:8080",
		},
		"web": gin.H{
			"port":      3000,
			"enable_ui": true,
			"theme":     "auto",
			"dogfood":   s.config.Dogfood,
		},
		"logging": gin.H{
			"level":       "info",
			"format":      "console",
			"development": true,
		},
	})
}

func (s *Service) SaveConfig(c *gin.Context) {
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement config saving
	c.JSON(http.StatusOK, gin.H{"message": "Configuration saved successfully"})
}

func (s *Service) GetSystemStatus(c *gin.Context) {
	// Calculate uptime
	uptime := time.Since(s.startTime)
	uptimeStr := formatDuration(uptime)

	// Get memory usage
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memoryUsage := formatBytes(m.Alloc)

	// Get storage usage
	storageUsed := s.getStorageUsage()

	c.JSON(http.StatusOK, gin.H{
		"uptime":       uptimeStr,
		"memory_usage": memoryUsage,
		"storage_used": storageUsed,
		"status":       "healthy",
	})
}

// formatDuration formats a duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
}

// formatBytes formats bytes into a human-readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// getStorageUsage calculates the storage usage of the database
func (s *Service) getStorageUsage() string {
	// Get the database path from storage
	dbPath := s.storage.GetDatabasePath()

	// Try to get file info for the database
	if fileInfo, err := os.Stat(dbPath); err == nil {
		return formatBytes(uint64(fileInfo.Size()))
	}

	// Fallback if we can't get the file size
	return "Unknown"
}
