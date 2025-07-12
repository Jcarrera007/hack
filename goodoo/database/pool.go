package database

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ConnectionPool manages database connections similar to Odoo's ConnectionPool
type ConnectionPool struct {
	connections map[string]*pooledConnection
	maxConns    int
	mutex       sync.RWMutex
	logger      logger.Interface
}

// pooledConnection represents a connection in the pool
type pooledConnection struct {
	db       *gorm.DB
	config   *ConnectionConfig
	used     bool
	lastUsed time.Time
	mutex    sync.Mutex
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxConns int) *ConnectionPool {
	if maxConns <= 0 {
		maxConns = 64
	}
	
	return &ConnectionPool{
		connections: make(map[string]*pooledConnection),
		maxConns:    maxConns,
		logger:      logger.Default.LogMode(logger.Info),
	}
}

// SetLogger sets the GORM logger for all connections
func (p *ConnectionPool) SetLogger(l logger.Interface) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.logger = l
}

// Borrow gets a connection from the pool or creates a new one
func (p *ConnectionPool) Borrow(config *ConnectionConfig) (*Connection, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	connKey := p.getConnectionKey(config)
	
	// Try to find an existing unused connection
	if pooledConn, exists := p.connections[connKey]; exists {
		pooledConn.mutex.Lock()
		if !pooledConn.used {
			// Test if connection is still alive
			if p.testConnection(pooledConn.db) {
				pooledConn.used = true
				pooledConn.lastUsed = time.Now()
				pooledConn.mutex.Unlock()
				return &Connection{
					db:     pooledConn.db,
					config: pooledConn.config,
					pool:   p,
					key:    connKey,
				}, nil
			}
		}
		pooledConn.mutex.Unlock()
	}
	
	// Create new connection if under limit
	if len(p.connections) >= p.maxConns {
		// Try to clean up unused connections
		p.cleanupUnusedConnections()
		
		if len(p.connections) >= p.maxConns {
			return nil, fmt.Errorf("connection pool exhausted (max %d connections)", p.maxConns)
		}
	}
	
	// Create new connection
	db, err := p.createConnection(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}
	
	pooledConn := &pooledConnection{
		db:       db,
		config:   config.Clone(),
		used:     true,
		lastUsed: time.Now(),
	}
	
	p.connections[connKey] = pooledConn
	
	return &Connection{
		db:     db,
		config: config,
		pool:   p,
		key:    connKey,
	}, nil
}

// Return returns a connection to the pool
func (p *ConnectionPool) Return(key string) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	if pooledConn, exists := p.connections[key]; exists {
		pooledConn.mutex.Lock()
		pooledConn.used = false
		pooledConn.lastUsed = time.Now()
		pooledConn.mutex.Unlock()
	}
}

// CloseAll closes all connections for a specific database
func (p *ConnectionPool) CloseAll(config *ConnectionConfig) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	connKey := p.getConnectionKey(config)
	
	if pooledConn, exists := p.connections[connKey]; exists {
		pooledConn.mutex.Lock()
		if sqlDB, err := pooledConn.db.DB(); err == nil {
			sqlDB.Close()
		}
		pooledConn.mutex.Unlock()
		delete(p.connections, connKey)
	}
}

// CloseAllConnections closes all connections in the pool
func (p *ConnectionPool) CloseAllConnections() {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	for key, pooledConn := range p.connections {
		pooledConn.mutex.Lock()
		if sqlDB, err := pooledConn.db.DB(); err == nil {
			sqlDB.Close()
		}
		pooledConn.mutex.Unlock()
		delete(p.connections, key)
	}
}

// Stats returns pool statistics
func (p *ConnectionPool) Stats() PoolStats {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	
	stats := PoolStats{
		TotalConnections: len(p.connections),
		MaxConnections:   p.maxConns,
	}
	
	for _, pooledConn := range p.connections {
		pooledConn.mutex.Lock()
		if pooledConn.used {
			stats.UsedConnections++
		} else {
			stats.IdleConnections++
		}
		pooledConn.mutex.Unlock()
	}
	
	return stats
}

// PoolStats represents connection pool statistics
type PoolStats struct {
	TotalConnections int
	UsedConnections  int
	IdleConnections  int
	MaxConnections   int
}

// String returns a string representation of pool stats
func (s PoolStats) String() string {
	return fmt.Sprintf("ConnectionPool(used=%d/idle=%d/total=%d/max=%d)",
		s.UsedConnections, s.IdleConnections, s.TotalConnections, s.MaxConnections)
}

// createConnection creates a new GORM database connection
func (p *ConnectionPool) createConnection(config *ConnectionConfig) (*gorm.DB, error) {
	dsn := config.BuildDSN()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: p.logger,
	})
	if err != nil {
		return nil, err
	}
	
	// Configure connection pool settings
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(config.MaxOpenConns)
		sqlDB.SetMaxIdleConns(config.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
	
	return db, nil
}

// testConnection tests if a connection is still alive
func (p *ConnectionPool) testConnection(db *gorm.DB) bool {
	if sqlDB, err := db.DB(); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return sqlDB.PingContext(ctx) == nil
	}
	return false
}

// getConnectionKey generates a unique key for a connection configuration
func (p *ConnectionPool) getConnectionKey(config *ConnectionConfig) string {
	if config.DSN != "" {
		return config.DSN
	}
	return fmt.Sprintf("%s:%d/%s@%s", config.Host, config.Port, config.Database, config.User)
}

// cleanupUnusedConnections removes old unused connections
func (p *ConnectionPool) cleanupUnusedConnections() {
	cutoff := time.Now().Add(-30 * time.Minute) // Remove connections unused for 30 minutes
	
	for key, pooledConn := range p.connections {
		pooledConn.mutex.Lock()
		if !pooledConn.used && pooledConn.lastUsed.Before(cutoff) {
			if sqlDB, err := pooledConn.db.DB(); err == nil {
				sqlDB.Close()
			}
			pooledConn.mutex.Unlock()
			delete(p.connections, key)
		} else {
			pooledConn.mutex.Unlock()
		}
	}
}