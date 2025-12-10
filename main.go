package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"open-telemorph-prime/internal/config"
	"open-telemorph-prime/internal/ingestion"
	"open-telemorph-prime/internal/logger"
	"open-telemorph-prime/internal/storage"
	"open-telemorph-prime/internal/web"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
	version    = "0.1.0"
)

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		// Use standard log for fatal errors before logger is initialized
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger with configuration
	if err := logger.Init(cfg.Logging.Level, cfg.Logging.Format, cfg.Logging.FilePath); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	log := logger.Get()

	// Initialize storage
	storage, err := storage.NewSQLiteStorage(cfg.Storage)
	if err != nil {
		log.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer storage.Close()

	// Initialize ingestion service
	ingestionService := ingestion.NewService(storage, cfg.Ingestion)

	// Initialize web service
	webService := web.NewService(storage, cfg.Web)

	// Set up Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Load HTML templates
	router.LoadHTMLGlob("web/*.html")

	// Register routes
	registerRoutes(router, ingestionService, webService)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start ingestion service
	go func() {
		if err := ingestionService.Start(); err != nil {
			log.Fatal("Failed to start ingestion service", zap.Error(err))
		}
	}()

	// Start HTTP server
	go func() {
		log.Info("Starting Open-Telemorph-Prime server",
			zap.Int("port", cfg.Server.Port),
			zap.String("environment", cfg.Server.Environment),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Open-Telemorph-Prime...")

	// Shutdown ingestion service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := ingestionService.Stop(ctx); err != nil {
		log.Error("Error stopping ingestion service", zap.Error(err))
	}

	// Shutdown HTTP server
	if err := server.Shutdown(ctx); err != nil {
		log.Error("Error shutting down server", zap.Error(err))
	}

	log.Info("Open-Telemorph-Prime stopped")
}

func registerRoutes(router *gin.Engine, ingestionService *ingestion.Service, webService *web.Service) {
	// Health endpoints
	router.GET("/health", healthCheck)
	router.GET("/ready", readinessCheck)

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/metrics", webService.GetMetrics)
		api.GET("/traces", webService.GetTraces)
		api.GET("/logs", webService.GetLogs)
		api.GET("/services", webService.GetServices)
		api.POST("/query", webService.Query)
	}

	// Admin API routes
	admin := router.Group("/api/v1/admin")
	{
		admin.GET("/config", webService.GetConfig)
		admin.POST("/config", webService.SaveConfig)
		admin.GET("/status", webService.GetSystemStatus)
	}

	// OTLP endpoints are now served on dedicated ingestion ports (4317/4318)
	// These are handled by the ingestion service directly

	// Web UI
	router.Static("/static", "./web/static")
	router.GET("/", webService.Index)
	router.GET("/dashboard", webService.Dashboard)
	router.GET("/metrics", webService.MetricsPage)
	router.GET("/traces", webService.TracesPage)
	router.GET("/logs", webService.LogsPage)
	router.GET("/services", webService.ServicesPage)
	router.GET("/alerts", webService.AlertsPage)
	router.GET("/query", webService.QueryPage)
	router.GET("/admin", webService.AdminPage)
}

func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
		"version":   version,
	})
}

func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().Unix(),
		"version":   version,
	})
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
