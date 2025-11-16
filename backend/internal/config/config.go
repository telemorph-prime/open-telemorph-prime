package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Storage   StorageConfig   `yaml:"storage"`
	Ingestion IngestionConfig `yaml:"ingestion"`
	Web       WebConfig       `yaml:"web"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
	Port         int           `yaml:"port"`
	Environment  string        `yaml:"environment"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type StorageConfig struct {
	Type           string `yaml:"type"`
	Path           string `yaml:"path"`
	RetentionDays  int    `yaml:"retention_days"`
	MaxConnections int    `yaml:"max_connections"`
}

type IngestionConfig struct {
	GRPCPort      int           `yaml:"grpc_port"`
	HTTPPort      int           `yaml:"http_port"`
	GRPCEnabled   bool          `yaml:"grpc_enabled"`
	HTTPEnabled   bool          `yaml:"http_enabled"`
	BatchSize     int           `yaml:"batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
}

type WebConfig struct {
	Enabled bool   `yaml:"enabled"`
	Title   string `yaml:"title"`
	Theme   string `yaml:"theme"`
	Dogfood bool   `yaml:"dogfood"`
}

type LoggingConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	FilePath string `yaml:"file_path"`
}

func Load(path string) (*Config, error) {
	// Check if config file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config if it doesn't exist
		cfg := DefaultConfig()
		if err := cfg.Save(path); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return cfg, nil
	}

	// Read existing config file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults for any missing values
	cfg.setDefaults()

	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) setDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.Environment == "" {
		c.Server.Environment = "development"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30 * time.Second
	}

	if c.Storage.Type == "" {
		c.Storage.Type = "sqlite"
	}
	if c.Storage.Path == "" {
		c.Storage.Path = "./data/telemorph.db"
	}
	if c.Storage.RetentionDays == 0 {
		c.Storage.RetentionDays = 30
	}
	if c.Storage.MaxConnections == 0 {
		c.Storage.MaxConnections = 10
	}

	if c.Ingestion.GRPCPort == 0 {
		c.Ingestion.GRPCPort = 4317
	}
	if c.Ingestion.HTTPPort == 0 {
		c.Ingestion.HTTPPort = 4318
	}
	if !c.Ingestion.GRPCEnabled && !c.Ingestion.HTTPEnabled {
		// If neither is explicitly set, enable both by default
		c.Ingestion.GRPCEnabled = true
		c.Ingestion.HTTPEnabled = true
	}
	if c.Ingestion.BatchSize == 0 {
		c.Ingestion.BatchSize = 1000
	}
	if c.Ingestion.FlushInterval == 0 {
		c.Ingestion.FlushInterval = 5 * time.Second
	}

	if c.Web.Title == "" {
		c.Web.Title = "Open-Telemorph-Prime"
	}
	if c.Web.Theme == "" {
		c.Web.Theme = "light"
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
}

func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			Environment:  "development",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		},
		Storage: StorageConfig{
			Type:           "sqlite",
			Path:           "./data/telemorph.db",
			RetentionDays:  30,
			MaxConnections: 10,
		},
		Ingestion: IngestionConfig{
			GRPCPort:      4317,
			HTTPPort:      4318,
			GRPCEnabled:   true,
			HTTPEnabled:   true,
			BatchSize:     1000,
			FlushInterval: 5 * time.Second,
		},
		Web: WebConfig{
			Enabled: true,
			Title:   "Open-Telemorph-Prime",
			Theme:   "light",
			Dogfood: false,
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
		},
	}
}

