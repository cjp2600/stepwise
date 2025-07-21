package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds application configuration
type Config struct {
	LogLevel    string
	Timeout     time.Duration
	Parallel    int
	Environment string
	Output      string
	Verbose     bool
	Quiet       bool
	Watch       bool
}

// Load loads configuration from environment variables and defaults
func Load() (*Config, error) {
	cfg := &Config{
		LogLevel:    getEnv("STEPWISE_LOG_LEVEL", "info"),
		Timeout:     getEnvDuration("STEPWISE_TIMEOUT", 30*time.Second),
		Parallel:    getEnvInt("STEPWISE_PARALLEL", 1),
		Environment: getEnv("STEPWISE_ENV", "development"),
		Output:      getEnv("STEPWISE_OUTPUT", "console"),
		Verbose:     getEnvBool("STEPWISE_VERBOSE", false),
		Quiet:       getEnvBool("STEPWISE_QUIET", false),
		Watch:       getEnvBool("STEPWISE_WATCH", false),
	}

	return cfg, nil
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as integer with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as boolean with a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvDuration gets an environment variable as duration with a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
