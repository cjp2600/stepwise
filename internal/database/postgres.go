package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cjp2600/stepwise/internal/logger"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// PostgresProvider implements the Provider interface for PostgreSQL
type PostgresProvider struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewPostgresProvider creates a new PostgreSQL provider
func NewPostgresProvider(log *logger.Logger) *PostgresProvider {
	return &PostgresProvider{
		logger: log,
	}
}

// Connect establishes a connection to PostgreSQL
func (p *PostgresProvider) Connect(config *Config) error {
	var connStr string

	// If DSN is provided, use it directly; otherwise build from individual parameters
	if config.DSN != "" {
		connStr = config.DSN
		p.logger.Debug("Connecting to PostgreSQL using DSN")
	} else {
		// Build connection string from individual parameters
		connStr = p.buildConnectionString(config)
		p.logger.Debug("Connecting to PostgreSQL",
			"host", config.Host,
			"port", config.Port,
			"database", config.Database)
	}

	var err error
	p.db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Set connection pool settings
	p.db.SetMaxOpenConns(25)
	p.db.SetMaxIdleConns(5)
	p.db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if config.Timeout > 0 {
		p.db.SetConnMaxLifetime(config.Timeout)
	}

	if err := p.db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.logger.Debug("Successfully connected to PostgreSQL")
	return nil
}

// ExecuteQuery executes a SQL query and returns results as JSON
func (p *PostgresProvider) ExecuteQuery(query string) (interface{}, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection is not established")
	}

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare slice for results
	var results []map[string]interface{}

	// Scan rows
	for rows.Next() {
		// Create a slice of interface{} to hold column values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row into value pointers
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Handle different types
			switch v := val.(type) {
			case []byte:
				// Try to parse as JSON if it looks like JSON
				var jsonVal interface{}
				if err := json.Unmarshal(v, &jsonVal); err == nil {
					rowMap[col] = jsonVal
				} else {
					rowMap[col] = string(v)
				}
			case time.Time:
				// Format time as RFC3339 string
				rowMap[col] = v.Format(time.RFC3339)
			case nil:
				rowMap[col] = nil
			default:
				rowMap[col] = v
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// If only one row, return it directly; otherwise return array
	if len(results) == 1 {
		return results[0], nil
	}

	if len(results) == 0 {
		// Return empty array for no results
		return []interface{}{}, nil
	}

	return results, nil
}

// Close closes the database connection
func (p *PostgresProvider) Close() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// IsConnected returns true if the connection is active
func (p *PostgresProvider) IsConnected() bool {
	if p.db == nil {
		return false
	}
	return p.db.Ping() == nil
}

// buildConnectionString builds a PostgreSQL connection string
func (p *PostgresProvider) buildConnectionString(config *Config) string {
	// Default port
	port := config.Port
	if port == 0 {
		port = 5432
	}

	// Build connection string
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host, port, config.Username, config.Password, config.Database)

	// Add SSL mode
	if config.SSLMode != "" {
		connStr += " sslmode=" + config.SSLMode
	} else {
		connStr += " sslmode=disable"
	}

	// Add additional options
	for key, value := range config.Options {
		connStr += fmt.Sprintf(" %s=%s", key, value)
	}

	return connStr
}

