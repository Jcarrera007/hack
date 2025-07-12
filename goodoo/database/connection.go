package database

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"gorm.io/gorm"
)

// Connection represents a database connection similar to Odoo's Connection class
type Connection struct {
	db     *gorm.DB
	config *ConnectionConfig
	pool   *ConnectionPool
	key    string
	mutex  sync.Mutex
}

// DB returns the underlying GORM database instance
func (c *Connection) DB() *gorm.DB {
	return c.db
}

// Config returns the connection configuration
func (c *Connection) Config() *ConnectionConfig {
	return c.config
}

// Close returns the connection to the pool
func (c *Connection) Close() {
	if c.pool != nil {
		c.pool.Return(c.key)
	}
}

// Cursor creates a new cursor for database operations
func (c *Connection) Cursor() *Cursor {
	return &Cursor{
		db:         c.db,
		connection: c,
	}
}

// Transaction executes a function within a database transaction
func (c *Connection) Transaction(fn func(*gorm.DB) error) error {
	return c.db.Transaction(fn)
}

// Ping tests the database connection
func (c *Connection) Ping() error {
	if sqlDB, err := c.db.DB(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return sqlDB.PingContext(ctx)
	}
	return fmt.Errorf("failed to get underlying SQL DB")
}

// Cursor represents a database cursor similar to Odoo's Cursor class
type Cursor struct {
	db         *gorm.DB
	connection *Connection
	savepoints []string
	mutex      sync.Mutex
}

// Execute executes a raw SQL query
func (c *Cursor) Execute(query string, args ...interface{}) error {
	return c.db.Exec(query, args...).Error
}

// Query executes a query and returns results
func (c *Cursor) Query(dest interface{}, query string, args ...interface{}) error {
	return c.db.Raw(query, args...).Scan(dest).Error
}

// Begin starts a new transaction
func (c *Cursor) Begin() error {
	c.db = c.db.Begin()
	return c.db.Error
}

// Commit commits the current transaction
func (c *Cursor) Commit() error {
	return c.db.Commit().Error
}

// Rollback rolls back the current transaction
func (c *Cursor) Rollback() error {
	return c.db.Rollback().Error
}

// Savepoint creates a new savepoint similar to Odoo's Savepoint
func (c *Cursor) Savepoint() (*Savepoint, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	sp := NewSavepoint(c)
	if err := sp.Create(); err != nil {
		return nil, err
	}
	
	c.savepoints = append(c.savepoints, sp.Name())
	return sp, nil
}

// Connection returns the associated connection
func (c *Cursor) Connection() *Connection {
	return c.connection
}

// Savepoint represents a database savepoint
type Savepoint struct {
	name   string
	cursor *Cursor
	closed bool
	mutex  sync.Mutex
}

// NewSavepoint creates a new savepoint
func NewSavepoint(cursor *Cursor) *Savepoint {
	return &Savepoint{
		name:   fmt.Sprintf("sp_%d", time.Now().UnixNano()),
		cursor: cursor,
	}
}

// Name returns the savepoint name
func (s *Savepoint) Name() string {
	return s.name
}

// Create creates the savepoint in the database
func (s *Savepoint) Create() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.closed {
		return fmt.Errorf("savepoint already closed")
	}
	
	return s.cursor.Execute(fmt.Sprintf("SAVEPOINT %s", s.name))
}

// Rollback rolls back to this savepoint
func (s *Savepoint) Rollback() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.closed {
		return fmt.Errorf("savepoint already closed")
	}
	
	return s.cursor.Execute(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", s.name))
}

// Release releases the savepoint
func (s *Savepoint) Release() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.closed {
		return nil
	}
	
	s.closed = true
	return s.cursor.Execute(fmt.Sprintf("RELEASE SAVEPOINT %s", s.name))
}

// Close closes the savepoint (rollback by default)
func (s *Savepoint) Close(rollback bool) error {
	if rollback {
		if err := s.Rollback(); err != nil {
			return err
		}
	}
	return s.Release()
}

// Global connection pool instance
var globalPool *ConnectionPool
var poolOnce sync.Once

// GetPool returns the global connection pool
func GetPool() *ConnectionPool {
	poolOnce.Do(func() {
		globalPool = NewConnectionPool(64) // Default max connections
	})
	return globalPool
}

// SetPool sets the global connection pool
func SetPool(pool *ConnectionPool) {
	globalPool = pool
}

// Connect creates a new database connection
func Connect(dbOrURI string, allowURI bool) (*Connection, error) {
	dbName, config, err := ParseConnectionInfo(dbOrURI)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection info: %w", err)
	}
	
	if !allowURI && dbName != dbOrURI {
		return nil, fmt.Errorf("URI connections not allowed")
	}
	
	pool := GetPool()
	return pool.Borrow(config)
}

// CloseDB closes all connections for a specific database
func CloseDB(dbName string) {
	_, config, err := ParseConnectionInfo(dbName)
	if err != nil {
		return
	}
	
	if globalPool != nil {
		globalPool.CloseAll(config)
	}
}

// CloseAll closes all database connections
func CloseAll() {
	if globalPool != nil {
		globalPool.CloseAllConnections()
	}
}

// Stats returns global pool statistics
func Stats() PoolStats {
	if globalPool != nil {
		return globalPool.Stats()
	}
	return PoolStats{}
}