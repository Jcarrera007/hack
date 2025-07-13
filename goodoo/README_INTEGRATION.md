# Goodoo - Odoo-Inspired Framework in Go

A comprehensive Go framework that integrates the core concepts from Odoo's netsvc.py, http.py, loglevels.py, and fields.py into a unified, production-ready system.

## 🏗️ Architecture Overview

Goodoo provides a complete web application framework with:

- **Advanced Logging System** - Multi-level, colored, performance-tracked logging
- **Session Management** - Persistent sessions with filesystem storage
- **HTTP Framework** - Request/response handling with middleware pipeline
- **Field System** - Type-safe field definitions with validation and conversion
- **Model System** - Database models with field definitions and ORM integration

## 📦 Components

### 1. Logging System (`logging/`)

Based on Odoo's `netsvc.py` and `loglevels.py`:

- **Colored Output** - Terminal color support with level-based coloring
- **Multiple Handlers** - Console, file, PostgreSQL, and syslog handlers
- **Performance Tracking** - Built-in query and request performance monitoring
- **Level Hierarchy** - DEBUG, INFO, WARNING, ERROR, CRITICAL with proper filtering
- **Context Integration** - Request context propagation through all log messages

#### Usage:
```go
logger := logging.GetLogger("myapp.module")
logger.InfoCtx(ctx, "User %s logged in", username)
logger.ErrorCtx(ctx, "Database error: %v", err)
```

#### Environment Configuration:
- `GOODOO_LOG_LEVEL` - debug, info, warn, error, critical
- `GOODOO_LOG_FILE` - file path for log output
- `GOODOO_LOG_DB` - PostgreSQL database for log storage
- `GOODOO_COLORS` - force colored output

### 2. HTTP Framework (`http/`)

Based on Odoo's `http.py`:

- **Session Management** - Persistent sessions across requests
- **Request Wrapper** - Enhanced Echo context with session utilities
- **Authentication** - Session-based user authentication
- **Middleware Pipeline** - Security, logging, performance, and error handling
- **Context Propagation** - Request metadata available throughout the request lifecycle

#### Key Components:
- `Session` - Persistent data storage with authentication support
- `Request` - Enhanced request wrapper with parameter parsing
- `RequestMiddleware` - Session initialization and management
- `AuthenticationMiddleware` - Route-level authentication control

#### API Endpoints:
```
Public Routes:
- GET  /health          - Basic health check
- POST /auth/login      - User authentication  
- GET  /db/list         - Available databases

Protected Routes:
- POST /auth/logout     - User logout
- GET  /auth/session    - Session information
- POST /db/set          - Change database
```

### 3. Field System (`fields/`)

Based on Odoo's `fields.py`:

- **Type Safety** - Strongly typed field definitions with validation
- **Data Conversion** - Automatic type conversion for cache, database, and export
- **Validation** - Built-in validation with custom rules support
- **SQL Generation** - Automatic PostgreSQL schema generation
- **Field Types** - Boolean, Integer, Float, String, Text, Date, Datetime, Selection, Binary, Json

#### Field Types:
```go
// Basic fields
boolField := fields.NewBooleanField(attrs)
intField := fields.NewIntegerField(attrs)
strField := fields.NewStringField(attrs)
dateField := fields.NewDateField(attrs)

// Selection field
selectionField := fields.NewSelectionField(attrs)
selectionField.SetSelection([]fields.SelectionOption{
    {Value: "draft", Label: "Draft"},
    {Value: "confirmed", Label: "Confirmed"},
})
```

### 4. Model System (`models/`)

Integration with existing GORM-based models plus enhanced field definitions:

- **ModelDefinition** - Rich model metadata with field definitions
- **Field Registry** - Centralized field type management
- **Data Validation** - Automatic validation using field definitions
- **Schema Generation** - SQL DDL generation from field definitions
- **API Integration** - Field metadata for API responses

#### Usage:
```go
// Create model definition
model := models.NewModelDefinition("partner", "partners")

// Add fields
nameField, _ := fields.CreateField(fields.StringType, fields.FieldAttribute{
    String:   "Name",
    Required: true,
    Size:     255,
})
model.AddField("name", nameField)

// Validate data
err := model.ValidateData(map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
})
```

## 🚀 Getting Started

### 1. Basic Application

```go
package main

import (
    "goodoo/logging"
    "goodoo/http"
    "github.com/labstack/echo/v4"
)

func main() {
    // Initialize logging
    logging.InitLogger()
    logger := logging.GetLogger("myapp")
    
    // Create session store
    sessionStore, _ := http.NewFilesystemSessionStore("./sessions", true)
    
    // Configure HTTP
    config := &http.RequestConfig{
        SessionStore:  sessionStore,
        DefaultDBName: "mydb",
        Logger:       logger,
    }
    
    e := echo.New()
    e.Use(http.RequestMiddleware(config))
    e.Use(logging.PerformanceMiddleware())
    
    e.GET("/", func(c echo.Context) error {
        req := http.GetGoodooRequest(c)
        req.Logger.InfoCtx(req.Context, "Hello world!")
        return c.JSON(200, map[string]string{"message": "Hello"})
    })
    
    logger.Info("Starting server on :8080")
    e.Start(":8080")
}
```

### 2. Model with Fields

```go
// Define model
model := models.NewModelDefinition("user", "users")

// Add fields
nameField, _ := fields.CreateField(fields.StringType, fields.FieldAttribute{
    String:   "Full Name",
    Required: true,
})
model.AddField("name", nameField)

emailField, _ := fields.CreateField(fields.StringType, fields.FieldAttribute{
    String: "Email",
    Help:   "User's email address",
})
model.AddField("email", emailField)

// Register model
models.RegisterFieldModel(model)

// Use in HTTP handler
func validateUserData(c echo.Context) error {
    req := http.GetGoodooRequest(c)
    
    var data map[string]interface{}
    c.Bind(&data)
    
    userModel, _ := models.GetFieldModel("user")
    if err := userModel.ValidateData(data); err != nil {
        return c.JSON(400, map[string]string{"error": err.Error()})
    }
    
    converted, _ := userModel.ConvertData(data, "cache")
    return c.JSON(200, converted)
}
```

## 🧪 Testing

Run the integration tests:

```bash
# Test field system
go run cmd/test_fields.go

# Test full integration
go run examples/integration_example.go
```

## 📊 Performance Features

- **Request Tracking** - Every request gets a unique ID
- **Performance Metrics** - Query count, query time, and processing time
- **Colored Performance** - Visual indicators for slow operations
- **Context Propagation** - Performance data available in all log messages

## 🔒 Security Features

- **Session Security** - HTTP-only cookies with secure flags
- **CSRF Protection** - Built-in CSRF protection middleware
- **Security Headers** - XSS, clickjacking, and content-type protection
- **Authentication** - Session-based authentication with middleware

## 🌍 Environment Configuration

```bash
# Logging
export GOODOO_LOG_LEVEL=debug
export GOODOO_LOG_FILE=/var/log/goodoo.log
export GOODOO_LOG_DB=goodoo_logs
export GOODOO_COLORS=1

# HTTP
export GOODOO_SESSION_DIR=/var/sessions
export GOODOO_DEFAULT_DB=production
export PORT=8080

# Database
export DATABASE_URL=postgres://user:pass@localhost/goodoo
```

## 🎯 Key Benefits

1. **Odoo Compatibility** - Familiar patterns for Odoo developers
2. **Type Safety** - Go's type system prevents runtime errors
3. **Performance** - Compiled Go performance with built-in monitoring
4. **Observability** - Comprehensive logging and request tracking
5. **Middleware Pipeline** - Extensible request processing pipeline
6. **Field Validation** - Automatic data validation and conversion
7. **Session Management** - Persistent user sessions across requests

## 🔧 Architecture Decisions

- **Echo Framework** - Lightweight, fast HTTP framework
- **GORM** - Go ORM for database operations
- **PostgreSQL** - Primary database with JSON support
- **Filesystem Sessions** - Simple, reliable session storage
- **Structured Logging** - JSON-compatible log format
- **Context Propagation** - Request context through all layers

## 📈 Scalability

- **Stateless Design** - Sessions stored externally for horizontal scaling
- **Database Connection Pooling** - Efficient database connections
- **Middleware Pipeline** - Easy to add caching, rate limiting, etc.
- **Performance Monitoring** - Built-in metrics for optimization
- **Field System** - Efficient data validation and conversion

## 🤝 Integration Points

The system is designed to integrate with:

- **Existing Go Applications** - Drop-in HTTP middleware
- **PostgreSQL Databases** - Native PostgreSQL support
- **Monitoring Systems** - Structured logs and metrics
- **Authentication Providers** - Pluggable authentication
- **Frontend Frameworks** - RESTful API with field metadata

This comprehensive framework brings Odoo's robust architecture patterns to the Go ecosystem while maintaining Go's performance and type safety advantages.