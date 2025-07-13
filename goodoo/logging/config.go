package logging

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// LogConfig holds logging configuration (similar to Odoo's tools.config)
type LogConfig struct {
	LogLevel    string
	LogFile     string
	LogDB       string
	LogDBLevel  string
	SysLog      bool
	LogHandler  []string
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() *LogConfig {
	return &LogConfig{
		LogLevel:   getEnv("GOODOO_LOG_LEVEL", "info"),
		LogFile:    getEnv("GOODOO_LOG_FILE", ""),
		LogDB:      getEnv("GOODOO_LOG_DB", ""),
		LogDBLevel: getEnv("GOODOO_LOG_DB_LEVEL", "warning"),
		SysLog:     getEnvBool("GOODOO_SYSLOG", false),
		LogHandler: getEnvSlice("GOODOO_LOG_HANDLER", []string{}),
	}
}

// getEnv gets environment variable with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets boolean environment variable
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// getEnvSlice gets slice from environment variable (comma-separated)
func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}

// Default log configuration mappings (based on netsvc.py)
var DefaultLogConfiguration = []string{
	"goodoo.http.rpc.request:INFO",
	"goodoo.http.rpc.response:INFO",
	":INFO",
}

// PseudoConfigMapper maps log level names to configurations
var PseudoConfigMapper = map[string][]string{
	"debug_rpc_answer": {"goodoo:DEBUG", "goodoo.sql_db:INFO", "goodoo.http.rpc:DEBUG"},
	"debug_rpc":        {"goodoo:DEBUG", "goodoo.sql_db:INFO", "goodoo.http.rpc.request:DEBUG"},
	"debug":            {"goodoo:DEBUG", "goodoo.sql_db:INFO"},
	"debug_sql":        {"goodoo.sql_db:DEBUG"},
	"info":             {},
	"warn":             {"goodoo:WARNING"},
	"error":            {"goodoo:ERROR"},
	"critical":         {"goodoo:CRITICAL"},
}

// ParseLogLevel converts string to LogLevel (delegates to levels.go)
func ParseLogLevel(level string) LogLevel {
	return ParseLogLevelString(level)
}

// GetLogConfigurations returns the complete log configuration
func (c *LogConfig) GetLogConfigurations() []string {
	// Start with default configuration
	configurations := make([]string, len(DefaultLogConfiguration))
	copy(configurations, DefaultLogConfiguration)
	
	// Add pseudo-config mappings
	if pseudoConfig, exists := PseudoConfigMapper[c.LogLevel]; exists {
		configurations = append(configurations, pseudoConfig...)
	}
	
	// Add custom log handlers
	configurations = append(configurations, c.LogHandler...)
	
	return configurations
}

// ParseLogConfiguration parses a log configuration string
func ParseLogConfiguration(config string) (logger string, level LogLevel, err error) {
	parts := strings.Split(strings.TrimSpace(config), ":")
	if len(parts) != 2 {
		return "", INFO, fmt.Errorf("invalid log configuration format: %s", config)
	}
	
	logger = parts[0]
	level = ParseLogLevel(parts[1])
	
	return logger, level, nil
}

// LoggerLevels holds the configured levels for different loggers
type LoggerLevels map[string]LogLevel

// BuildLoggerLevels builds a map of logger names to their configured levels
func (c *LogConfig) BuildLoggerLevels() LoggerLevels {
	levels := make(LoggerLevels)
	
	configurations := c.GetLogConfigurations()
	for _, config := range configurations {
		logger, level, err := ParseLogConfiguration(config)
		if err != nil {
			continue // Skip invalid configurations
		}
		levels[logger] = level
	}
	
	return levels
}

// GetLoggerLevel returns the configured level for a logger
func (ll LoggerLevels) GetLoggerLevel(name string) LogLevel {
	// Try exact match first
	if level, exists := ll[name]; exists {
		return level
	}
	
	// Try progressively shorter prefixes
	parts := strings.Split(name, ".")
	for i := len(parts) - 1; i > 0; i-- {
		prefix := strings.Join(parts[:i], ".")
		if level, exists := ll[prefix]; exists {
			return level
		}
	}
	
	// Try root logger
	if level, exists := ll[""]; exists {
		return level
	}
	
	// Default to INFO
	return INFO
}

// ShouldLog checks if a message at the given level should be logged
func (ll LoggerLevels) ShouldLog(loggerName string, level LogLevel) bool {
	configuredLevel := ll.GetLoggerLevel(loggerName)
	return CompareLogLevels(level, configuredLevel)
}