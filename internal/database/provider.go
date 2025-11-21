package database

import (
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
)

// Provider defines the interface for database providers
type Provider interface {
	// Connect establishes a connection to the database
	Connect(config *Config) error

	// ExecuteQuery executes a SQL query and returns results as JSON
	ExecuteQuery(query string) (interface{}, error)

	// Close closes the database connection
	Close() error

	// IsConnected returns true if the connection is active
	IsConnected() bool
}

// Config holds database connection configuration
type Config struct {
	Type     string            `yaml:"type" json:"type"`         // postgres, mysql, sqlite
	DSN      string            `yaml:"dsn" json:"dsn"`           // Data Source Name - if provided, used directly
	Host     string            `yaml:"host" json:"host"`
	Port     int               `yaml:"port" json:"port"`
	Database string            `yaml:"database" json:"database"`
	Username string            `yaml:"username" json:"username"`
	Password string            `yaml:"password" json:"password"`
	SSLMode  string            `yaml:"ssl_mode" json:"ssl_mode"` // postgres: disable, require, verify-ca, verify-full
	Options  map[string]string `yaml:"options" json:"options"`   // Additional connection options
	Timeout  time.Duration     `yaml:"timeout" json:"timeout"`
}

// Response represents a database query response
type Response struct {
	Data     interface{}   `json:"data"`
	Duration time.Duration `json:"duration"`
	Error    error         `json:"error,omitempty"`
	Rows     int           `json:"rows,omitempty"` // Number of rows returned
}

// Client represents a database client
type Client struct {
	provider Provider
	logger   *logger.Logger
	config   *Config
}

// NewClient creates a new database client
func NewClient(config *Config, log *logger.Logger) (*Client, error) {
	var provider Provider

	switch config.Type {
	case "postgres", "postgresql":
		provider = NewPostgresProvider(log)
	default:
		return nil, &UnsupportedDatabaseError{Type: config.Type}
	}

	client := &Client{
		provider: provider,
		logger:   log,
		config:   config,
	}

	if err := client.provider.Connect(config); err != nil {
		return nil, err
	}

	return client, nil
}

// Execute executes a database query
func (c *Client) Execute(query string) (*Response, error) {
	start := time.Now()

	c.logger.Debug("Executing database query",
		"type", c.config.Type,
		"database", c.config.Database,
		"query", query)

	data, err := c.provider.ExecuteQuery(query)
	if err != nil {
		return &Response{
			Data:     nil,
			Duration: time.Since(start),
			Error:    err,
		}, err
	}

	// Count rows if data is an array
	rows := 0
	if arr, ok := data.([]interface{}); ok {
		rows = len(arr)
	} else if data != nil {
		rows = 1
	}

	duration := time.Since(start)

	c.logger.Debug("Database query completed",
		"type", c.config.Type,
		"duration", duration,
		"rows", rows)

	return &Response{
		Data:     data,
		Duration: duration,
		Rows:     rows,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	if c.provider != nil {
		return c.provider.Close()
	}
	return nil
}

// IsConnected returns true if the connection is active
func (c *Client) IsConnected() bool {
	if c.provider != nil {
		return c.provider.IsConnected()
	}
	return false
}

// UnsupportedDatabaseError represents an error for unsupported database types
type UnsupportedDatabaseError struct {
	Type string
}

func (e *UnsupportedDatabaseError) Error() string {
	return "unsupported database type: " + e.Type
}

