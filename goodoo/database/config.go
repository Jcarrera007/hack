package database

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ConnectionConfig holds database connection configuration
type ConnectionConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Database     string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	AppName      string
	DSN          string // Direct DSN if provided
}

// DefaultConfig returns a default configuration
func DefaultConfig() *ConnectionConfig {
	return &ConnectionConfig{
		Host:         "localhost",
		Port:         5432,
		User:         "postgres",
		Password:     "",
		Database:     "",
		SSLMode:      "prefer",
		MaxOpenConns: 64,
		MaxIdleConns: 16,
		AppName:      fmt.Sprintf("goodoo-%d", os.Getpid()),
	}
}

// LoadFromEnv loads configuration from environment variables
func (c *ConnectionConfig) LoadFromEnv() {
	if host := os.Getenv("DB_HOST"); host != "" {
		c.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			c.Port = p
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		c.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		c.Password = password
	}
	if database := os.Getenv("DB_NAME"); database != "" {
		c.Database = database
	}
	if sslmode := os.Getenv("DB_SSLMODE"); sslmode != "" {
		c.SSLMode = sslmode
	}
	if maxConns := os.Getenv("DB_MAXCONN"); maxConns != "" {
		if mc, err := strconv.Atoi(maxConns); err == nil {
			c.MaxOpenConns = mc
		}
	}
	if appName := os.Getenv("GOODOO_PGAPPNAME"); appName != "" {
		// Support {pid} placeholder like Odoo
		c.AppName = strings.ReplaceAll(appName, "{pid}", strconv.Itoa(os.Getpid()))
		// Trim to PostgreSQL NAMEDATALEN limit
		if len(c.AppName) > 63 {
			c.AppName = c.AppName[:63]
		}
	}
}

// ParseConnectionInfo parses a database name or URI and returns database name and connection params
// Similar to Odoo's connection_info_for function
func ParseConnectionInfo(dbOrURI string) (string, *ConnectionConfig, error) {
	config := DefaultConfig()
	config.LoadFromEnv()
	
	// Check if it's a PostgreSQL URI
	if strings.HasPrefix(dbOrURI, "postgresql://") || strings.HasPrefix(dbOrURI, "postgres://") {
		return parseURI(dbOrURI, config)
	}
	
	// It's just a database name
	config.Database = dbOrURI
	return dbOrURI, config, nil
}

// parseURI parses a PostgreSQL URI and extracts connection info
func parseURI(uri string, config *ConnectionConfig) (string, *ConnectionConfig, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", nil, fmt.Errorf("invalid URI: %w", err)
	}
	
	// Extract database name
	dbName := ""
	if len(parsed.Path) > 1 {
		dbName = parsed.Path[1:] // Remove leading slash
	} else if parsed.User != nil {
		dbName = parsed.User.Username()
	} else {
		dbName = parsed.Hostname()
	}
	
	// Store the full DSN for direct use
	config.DSN = uri
	config.Database = dbName
	
	return dbName, config, nil
}

// BuildDSN builds a PostgreSQL DSN from the configuration
func (c *ConnectionConfig) BuildDSN() string {
	// If we have a direct DSN, use it
	if c.DSN != "" {
		return c.DSN
	}
	
	// Build DSN from individual components
	var parts []string
	
	if c.Host != "" {
		parts = append(parts, fmt.Sprintf("host=%s", c.Host))
	}
	if c.Port != 0 {
		parts = append(parts, fmt.Sprintf("port=%d", c.Port))
	}
	if c.User != "" {
		parts = append(parts, fmt.Sprintf("user=%s", c.User))
	}
	if c.Password != "" {
		parts = append(parts, fmt.Sprintf("password=%s", c.Password))
	}
	if c.Database != "" {
		parts = append(parts, fmt.Sprintf("dbname=%s", c.Database))
	}
	if c.SSLMode != "" {
		parts = append(parts, fmt.Sprintf("sslmode=%s", c.SSLMode))
	}
	if c.AppName != "" {
		parts = append(parts, fmt.Sprintf("application_name=%s", c.AppName))
	}
	
	return strings.Join(parts, " ")
}

// Validate checks if the configuration is valid
func (c *ConnectionConfig) Validate() error {
	if c.Database == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Host == "" && c.DSN == "" {
		return fmt.Errorf("host is required when DSN is not provided")
	}
	if c.Port <= 0 && c.DSN == "" {
		return fmt.Errorf("valid port is required when DSN is not provided")
	}
	return nil
}

// Clone creates a copy of the configuration
func (c *ConnectionConfig) Clone() *ConnectionConfig {
	return &ConnectionConfig{
		Host:         c.Host,
		Port:         c.Port,
		User:         c.User,
		Password:     c.Password,
		Database:     c.Database,
		SSLMode:      c.SSLMode,
		MaxOpenConns: c.MaxOpenConns,
		MaxIdleConns: c.MaxIdleConns,
		AppName:      c.AppName,
		DSN:          c.DSN,
	}
}