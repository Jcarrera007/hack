# Goodoo Database Layer

This package provides a comprehensive database connectivity layer inspired by Odoo's `sql_db.py`, implemented using GORM for Go. It manages PostgreSQL connections, connection pooling, transactions, and database lifecycle.

## Architecture Overview

The database layer consists of several key components:

- **ConnectionConfig**: Database connection configuration management
- **ConnectionPool**: Connection pooling and lifecycle management  
- **Connection**: Individual database connection wrapper
- **Cursor**: Database operations and transaction management
- **DatabaseRegistry**: Multi-database management and registration
- **Savepoint**: Database savepoint support for nested transactions

## Key Features

### Connection Management
- **Connection Pooling**: Automatic connection reuse and cleanup
- **URI Support**: PostgreSQL connection strings and individual parameters
- **Environment Variables**: Configuration via environment variables
- **Health Monitoring**: Connection health checks and automatic recovery

### Transaction Support
- **GORM Transactions**: Full GORM transaction support
- **Savepoints**: Nested transaction savepoints
- **Automatic Rollback**: Context-based transaction management

### Multi-Database Support
- **Database Registry**: Register and manage multiple databases
- **Connection Lifecycle**: Automatic connection cleanup and management
- **Database-specific Environments**: Isolated database contexts

## Configuration

### Environment Variables

```bash
# Database connection
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=goodoo_dev
DB_SSLMODE=disable

# Connection pooling
DB_MAXCONN=64

# Application name (supports {pid} placeholder)
GOODOO_PGAPPNAME=goodoo-{pid}
```

### Programmatic Configuration

```go
config := database.DefaultConfig()
config.Host = "localhost"
config.Port = 5432
config.User = "postgres"
config.Password = "password"
config.Database = "goodoo_dev"
config.SSLMode = "disable"
config.MaxOpenConns = 32
config.MaxIdleConns = 8
```

## Usage Examples

### Quick Setup

```go
// Quick setup with auto-migration
err := database.QuickSetup("goodoo_dev",
    &models.User{},
    &models.Partner{},
    &models.Product{},
)
if err != nil {
    log.Fatal(err)
}
```

### Manual Setup

```go
// Initialize the database system
opts := database.DefaultInitOptions()
opts.MaxConnections = 32
opts.LogLevel = logger.Info

if err := database.Initialize(opts); err != nil {
    log.Fatal(err)
}

// Configure database
config := database.DefaultConfig()
config.LoadFromEnv()
config.Database = "goodoo_dev"

// Register database
if err := database.RegisterDatabase("goodoo_dev", config); err != nil {
    log.Fatal(err)
}
```

### Getting Connections

```go
// Get a connection from the pool
conn, err := database.GetDatabaseConnection("goodoo_dev")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Get GORM DB instance
db, err := database.GetDatabase("goodoo_dev")
if err != nil {
    log.Fatal(err)
}
```

### Transaction Management

```go
// Using GORM transactions
err := conn.Transaction(func(tx *gorm.DB) error {
    // Create records
    if err := tx.Create(&user).Error; err != nil {
        return err
    }
    
    // Update records
    if err := tx.Model(&partner).Updates(updates).Error; err != nil {
        return err
    }
    
    return nil // Commit
})
```

### Savepoints

```go
cursor := conn.Cursor()

sp, err := cursor.Savepoint()
if err != nil {
    log.Fatal(err)
}

// Do some operations
if err := cursor.Execute("INSERT INTO ..."); err != nil {
    sp.Rollback() // Rollback to savepoint
} else {
    sp.Release() // Release savepoint
}
```

### URI Connections

```go
// Connect using PostgreSQL URI
conn, err := database.Connect("postgresql://user:pass@localhost/dbname", true)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()
```

## Integration with Models

### Creating Environment

```go
// Create environment for specific database
env, err := models.NewEnvironmentForDB("goodoo_dev", userID)
if err != nil {
    log.Fatal(err)
}

// Use with RecordSets
userRS := models.Model(env, models.User{})
users, err := userRS.Search(domain, 0, 10, "name")
```

### Auto-Migration

```go
registry := database.GetRegistry()
err := registry.AutoMigrate("goodoo_dev",
    &models.User{},
    &models.Partner{},
    &models.Product{},
)
```

## Connection Pool Management

### Pool Statistics

```go
stats := database.Stats()
fmt.Printf("Pool: used=%d/idle=%d/total=%d/max=%d\n",
    stats.UsedConnections,
    stats.IdleConnections, 
    stats.TotalConnections,
    stats.MaxConnections)
```

### Registry Statistics

```go
registry := database.GetRegistry()
stats := registry.Stats()
fmt.Printf("Registry: total=%d/active=%d/inactive=%d\n",
    stats.TotalDatabases,
    stats.ActiveDatabases,
    stats.InactiveDatabases)
```

### Cleanup Operations

```go
// Close specific database
database.CloseDB("goodoo_dev")

// Close all connections
database.CloseAll()

// Cleanup inactive connections (30 min idle)
registry.CleanupInactive(30 * time.Minute)

// Health check all databases
results := database.HealthCheck()
for dbName, err := range results {
    if err != nil {
        fmt.Printf("Database %s: ERROR - %v\n", dbName, err)
    } else {
        fmt.Printf("Database %s: OK\n", dbName)
    }
}
```

## Configuration Options

### ConnectionConfig Fields

- `Host`: Database server hostname
- `Port`: Database server port
- `User`: Database username
- `Password`: Database password  
- `Database`: Database name
- `SSLMode`: SSL connection mode (`disable`, `require`, `prefer`)
- `MaxOpenConns`: Maximum open connections
- `MaxIdleConns`: Maximum idle connections
- `AppName`: Application name for PostgreSQL
- `DSN`: Direct connection string (optional)

### InitOptions Fields

- `MaxConnections`: Pool maximum connections
- `LogLevel`: GORM logging level
- `SlowThreshold`: Slow query threshold
- `AutoMigrate`: Enable auto-migration
- `Models`: Models to auto-migrate

## Error Handling

The database layer provides detailed error information:

```go
conn, err := database.Connect("invalid://uri", true)
if err != nil {
    log.Printf("Connection failed: %v", err)
    return
}

// Connection validation
if err := conn.Ping(); err != nil {
    log.Printf("Database unreachable: %v", err)
    return
}
```

## Thread Safety

- **ConnectionPool**: Thread-safe with mutex protection
- **DatabaseRegistry**: Thread-safe for concurrent access
- **Connection**: Individual connections are not thread-safe
- **Cursor**: Not thread-safe, use one per goroutine

## Performance Considerations

- **Connection Reuse**: Pool reuses connections automatically
- **Idle Cleanup**: Inactive connections are cleaned up periodically
- **Health Checks**: Failed connections are removed from pool
- **Prepared Statements**: GORM handles statement preparation
- **Connection Limits**: Configure based on database server capacity

## Monitoring and Debugging

### Logging

```go
// Set GORM logger level
registry := database.GetRegistry()
registry.SetLogger(logger.Default.LogMode(logger.Info))
```

### Statistics Monitoring

```go
// Regular statistics reporting
ticker := time.NewTicker(5 * time.Minute)
go func() {
    for range ticker.C {
        stats := database.Stats()
        log.Printf("Pool stats: %s", stats.String())
    }
}()
```

## Comparison with Odoo

| Odoo sql_db.py | Goodoo database |
|---------------|-----------------|
| `ConnectionPool` | `ConnectionPool` |
| `Connection` | `Connection` |
| `BaseCursor`/`Cursor` | `Cursor` |
| `Savepoint` | `Savepoint` |
| `db_connect()` | `Connect()` |
| `connection_info_for()` | `ParseConnectionInfo()` |
| psycopg2 | GORM + pgx driver |
| Python dict config | Go struct config |

The Go implementation provides similar functionality with type safety and better performance characteristics.