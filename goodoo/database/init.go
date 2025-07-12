package database

import (
	"fmt"
	"time"
	
	"gorm.io/gorm/logger"
)

// InitOptions holds options for database initialization
type InitOptions struct {
	MaxConnections int
	LogLevel       logger.LogLevel
	SlowThreshold  time.Duration
	AutoMigrate    bool
	Models         []interface{}
}

// DefaultInitOptions returns default initialization options
func DefaultInitOptions() *InitOptions {
	return &InitOptions{
		MaxConnections: 64,
		LogLevel:       logger.Info,
		SlowThreshold:  200 * time.Millisecond,
		AutoMigrate:    false,
		Models:         []interface{}{},
	}
}

// Initialize sets up the database system with the given options
func Initialize(opts *InitOptions) error {
	if opts == nil {
		opts = DefaultInitOptions()
	}
	
	// Initialize connection pool
	pool := NewConnectionPool(opts.MaxConnections)
	
	// Set up logger
	customLogger := logger.New(
		nil, // Use default log writer
		logger.Config{
			SlowThreshold:             opts.SlowThreshold,
			LogLevel:                  opts.LogLevel,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)
	
	pool.SetLogger(customLogger)
	SetPool(pool)
	
	return nil
}

// SetupDatabase registers and initializes a specific database
func SetupDatabase(dbName string, config *ConnectionConfig, opts *InitOptions) error {
	if opts == nil {
		opts = DefaultInitOptions()
	}
	
	// Register the database
	registry := GetRegistry()
	if err := registry.Register(dbName, config); err != nil {
		return fmt.Errorf("failed to register database: %w", err)
	}
	
	// Auto-migrate if requested
	if opts.AutoMigrate && len(opts.Models) > 0 {
		if err := registry.AutoMigrate(dbName, opts.Models...); err != nil {
			return fmt.Errorf("failed to auto-migrate: %w", err)
		}
	}
	
	return nil
}

// ConnectWithEnv creates a database connection using environment variables
func ConnectWithEnv(dbName string) (*Connection, error) {
	config := DefaultConfig()
	config.LoadFromEnv()
	config.Database = dbName
	
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	pool := GetPool()
	return pool.Borrow(config)
}

// QuickSetup provides a quick way to set up a database with default settings
func QuickSetup(dbName string, models ...interface{}) error {
	config := DefaultConfig()
	config.LoadFromEnv()
	config.Database = dbName
	
	opts := DefaultInitOptions()
	opts.AutoMigrate = true
	opts.Models = models
	
	// Initialize the system if not already done
	if err := Initialize(opts); err != nil {
		return fmt.Errorf("failed to initialize database system: %w", err)
	}
	
	// Setup the specific database
	return SetupDatabase(dbName, config, opts)
}

// Cleanup performs cleanup operations
func Cleanup() {
	// Close all connections
	CloseAll()
	
	// Close registry connections
	if globalRegistry != nil {
		globalRegistry.CloseAll()
	}
}

// HealthCheck performs a health check on all registered databases
func HealthCheck() map[string]error {
	registry := GetRegistry()
	results := make(map[string]error)
	
	for _, dbName := range registry.ListDatabases() {
		conn, err := registry.GetConnection(dbName)
		if err != nil {
			results[dbName] = fmt.Errorf("failed to get connection: %w", err)
			continue
		}
		
		if err := conn.Ping(); err != nil {
			results[dbName] = fmt.Errorf("ping failed: %w", err)
		} else {
			results[dbName] = nil // Success
		}
	}
	
	return results
}