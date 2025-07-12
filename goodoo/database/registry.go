package database

import (
	"fmt"
	"sync"
	"time"
	
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseRegistry manages database connections and their lifecycle
// Similar to Odoo's database registry concept
type DatabaseRegistry struct {
	databases map[string]*DatabaseInfo
	mutex     sync.RWMutex
	pool      *ConnectionPool
}

// DatabaseInfo holds information about a registered database
type DatabaseInfo struct {
	Name         string
	Config       *ConnectionConfig
	Connection   *Connection
	LastAccessed time.Time
	Active       bool
	mutex        sync.RWMutex
}

// NewDatabaseRegistry creates a new database registry
func NewDatabaseRegistry() *DatabaseRegistry {
	return &DatabaseRegistry{
		databases: make(map[string]*DatabaseInfo),
		pool:      GetPool(),
	}
}

// Register registers a database with the registry
func (r *DatabaseRegistry) Register(dbName string, config *ConnectionConfig) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	if _, exists := r.databases[dbName]; exists {
		return fmt.Errorf("database %s already registered", dbName)
	}
	
	r.databases[dbName] = &DatabaseInfo{
		Name:         dbName,
		Config:       config.Clone(),
		LastAccessed: time.Now(),
		Active:       false,
	}
	
	return nil
}

// GetConnection gets or creates a connection for the specified database
func (r *DatabaseRegistry) GetConnection(dbName string) (*Connection, error) {
	r.mutex.RLock()
	dbInfo, exists := r.databases[dbName]
	r.mutex.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("database %s not registered", dbName)
	}
	
	dbInfo.mutex.Lock()
	defer dbInfo.mutex.Unlock()
	
	// Check if we have an active connection
	if dbInfo.Connection != nil && dbInfo.Active {
		if err := dbInfo.Connection.Ping(); err == nil {
			dbInfo.LastAccessed = time.Now()
			return dbInfo.Connection, nil
		}
		// Connection is dead, clean it up
		dbInfo.Connection.Close()
		dbInfo.Connection = nil
		dbInfo.Active = false
	}
	
	// Create new connection
	conn, err := r.pool.Borrow(dbInfo.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection for %s: %w", dbName, err)
	}
	
	dbInfo.Connection = conn
	dbInfo.Active = true
	dbInfo.LastAccessed = time.Now()
	
	return conn, nil
}

// GetDB gets the GORM database instance for the specified database
func (r *DatabaseRegistry) GetDB(dbName string) (*gorm.DB, error) {
	conn, err := r.GetConnection(dbName)
	if err != nil {
		return nil, err
	}
	return conn.DB(), nil
}

// CloseDatabase closes the connection for a specific database
func (r *DatabaseRegistry) CloseDatabase(dbName string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	dbInfo, exists := r.databases[dbName]
	if !exists {
		return fmt.Errorf("database %s not registered", dbName)
	}
	
	dbInfo.mutex.Lock()
	defer dbInfo.mutex.Unlock()
	
	if dbInfo.Connection != nil {
		dbInfo.Connection.Close()
		dbInfo.Connection = nil
	}
	dbInfo.Active = false
	
	return nil
}

// Unregister removes a database from the registry
func (r *DatabaseRegistry) Unregister(dbName string) error {
	if err := r.CloseDatabase(dbName); err != nil {
		return err
	}
	
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	delete(r.databases, dbName)
	return nil
}

// ListDatabases returns a list of registered database names
func (r *DatabaseRegistry) ListDatabases() []string {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	var names []string
	for name := range r.databases {
		names = append(names, name)
	}
	return names
}

// GetDatabaseInfo returns information about a registered database
func (r *DatabaseRegistry) GetDatabaseInfo(dbName string) (*DatabaseInfo, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	dbInfo, exists := r.databases[dbName]
	if !exists {
		return nil, fmt.Errorf("database %s not registered", dbName)
	}
	
	// Return a copy to avoid data races
	dbInfo.mutex.RLock()
	defer dbInfo.mutex.RUnlock()
	
	return &DatabaseInfo{
		Name:         dbInfo.Name,
		Config:       dbInfo.Config.Clone(),
		LastAccessed: dbInfo.LastAccessed,
		Active:       dbInfo.Active,
	}, nil
}

// CleanupInactive closes connections for databases that haven't been accessed recently
func (r *DatabaseRegistry) CleanupInactive(maxIdleTime time.Duration) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	cutoff := time.Now().Add(-maxIdleTime)
	
	for _, dbInfo := range r.databases {
		dbInfo.mutex.Lock()
		if dbInfo.Active && dbInfo.LastAccessed.Before(cutoff) {
			if dbInfo.Connection != nil {
				dbInfo.Connection.Close()
				dbInfo.Connection = nil
			}
			dbInfo.Active = false
		}
		dbInfo.mutex.Unlock()
	}
}

// CloseAll closes all database connections
func (r *DatabaseRegistry) CloseAll() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	
	for _, dbInfo := range r.databases {
		dbInfo.mutex.Lock()
		if dbInfo.Connection != nil {
			dbInfo.Connection.Close()
			dbInfo.Connection = nil
		}
		dbInfo.Active = false
		dbInfo.mutex.Unlock()
	}
}

// Stats returns registry statistics
func (r *DatabaseRegistry) Stats() RegistryStats {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	
	stats := RegistryStats{
		TotalDatabases: len(r.databases),
		PoolStats:      r.pool.Stats(),
	}
	
	for _, dbInfo := range r.databases {
		dbInfo.mutex.RLock()
		if dbInfo.Active {
			stats.ActiveDatabases++
		} else {
			stats.InactiveDatabases++
		}
		dbInfo.mutex.RUnlock()
	}
	
	return stats
}

// RegistryStats represents database registry statistics
type RegistryStats struct {
	TotalDatabases    int
	ActiveDatabases   int
	InactiveDatabases int
	PoolStats         PoolStats
}

// String returns a string representation of registry stats
func (s RegistryStats) String() string {
	return fmt.Sprintf("DatabaseRegistry(total=%d/active=%d/inactive=%d) %s",
		s.TotalDatabases, s.ActiveDatabases, s.InactiveDatabases, s.PoolStats.String())
}

// SetLogger sets the logger for all database connections
func (r *DatabaseRegistry) SetLogger(l logger.Interface) {
	r.pool.SetLogger(l)
}

// AutoMigrate runs auto-migration for all registered models on a database
func (r *DatabaseRegistry) AutoMigrate(dbName string, models ...interface{}) error {
	db, err := r.GetDB(dbName)
	if err != nil {
		return fmt.Errorf("failed to get database %s: %w", dbName, err)
	}
	
	return db.AutoMigrate(models...)
}

// Global database registry instance
var globalRegistry *DatabaseRegistry
var registryOnce sync.Once

// GetRegistry returns the global database registry
func GetRegistry() *DatabaseRegistry {
	registryOnce.Do(func() {
		globalRegistry = NewDatabaseRegistry()
	})
	return globalRegistry
}

// SetRegistry sets the global database registry
func SetRegistry(registry *DatabaseRegistry) {
	globalRegistry = registry
}

// RegisterDatabase registers a database with the global registry
func RegisterDatabase(dbName string, config *ConnectionConfig) error {
	return GetRegistry().Register(dbName, config)
}

// GetDatabaseConnection gets a connection from the global registry
func GetDatabaseConnection(dbName string) (*Connection, error) {
	return GetRegistry().GetConnection(dbName)
}

// GetDatabase gets a GORM DB instance from the global registry
func GetDatabase(dbName string) (*gorm.DB, error) {
	return GetRegistry().GetDB(dbName)
}