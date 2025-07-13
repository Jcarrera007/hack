# Goodoo Framework

A comprehensive Go framework inspired by Odoo's architecture, providing logging, HTTP handling, field validation, model management, and API decorators.

## 🏗️ Project Structure

```
goodoo/
├── main.go                     # Main entry point
├── api/                        # API system (decorators, registry)
│   └── decorators.go
├── database/                   # Database connection and management
│   ├── config.go
│   ├── connection.go
│   ├── init.go
│   ├── pool.go
│   └── registry.go
├── fields/                     # Field system (validation, conversion)
│   ├── base.go
│   └── basic.go
├── handlers/                   # HTTP handlers
│   ├── api.go                  # API endpoint handlers
│   ├── database.go             # Database management handlers
│   ├── health.go               # Health check handlers
│   ├── http.go                 # Authentication handlers
│   ├── index.go                # Main page handler
│   └── session.go              # Session management handlers
├── http/                       # HTTP framework (middleware, sessions)
│   ├── middleware.go
│   ├── request.go
│   └── session.go
├── logging/                    # Logging system
│   ├── colors.go
│   ├── config.go
│   ├── formatter.go
│   ├── handlers.go
│   ├── levels.go
│   ├── logger.go
│   ├── performance.go
│   └── utils.go
├── models/                     # Model definitions and registry
│   ├── base.go
│   ├── fields.go
│   ├── model.go
│   ├── registry.go
│   └── relations.go
└── tests/                      # Tests and examples
    ├── api_examples.go         # API usage examples
    ├── database_usage.go       # Database usage examples
    ├── integration_example.go  # Full integration example
    ├── model_examples.go       # Model usage examples
    ├── test_api.go            # API system tests
    └── test_fields.go         # Field system tests
```

## 🚀 Quick Start

### 1. Run the Main Application

```bash
go run main.go
```

The server will start on port 8080 with all systems integrated.

### 2. Run Tests

```bash
# Test field system
go run tests/test_fields.go

# Test API system  
go run tests/test_api.go

# Run full integration example
go run tests/integration_example.go
```

### 3. Environment Configuration

```bash
# Logging
export GOODOO_LOG_LEVEL=debug
export GOODOO_LOG_FILE=/var/log/goodoo.log
export GOODOO_COLORS=1

# HTTP
export GOODOO_SESSION_DIR=/var/sessions
export GOODOO_DEFAULT_DB=production
export PORT=8080

# Database
export DATABASE_URL=postgres://user:pass@localhost/goodoo
```

## 📋 Available Endpoints

### Core Endpoints
- `GET /` - Welcome page
- `GET /health` - Basic health check
- `GET /health/detailed` - Detailed health with authentication

### Authentication
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout  
- `GET /auth/session` - Session information

### Database Management
- `GET /db/list` - List available databases
- `POST /db/set` - Set current database

### Session Management
- `GET /session` - Get session data
- `POST /session/clear` - Clear session
- `POST /session/set` - Set session data

### API Endpoints
- `POST /api/call` - Generic API method call
- `GET /api/models/:model/methods` - List model methods
- `GET /api/models/:model/methods/:method` - Method information
- `ANY /api/models/:model/:method` - Call model method
- `ANY /api/models/:model/:ids/:method` - Call record method

## 🎯 Core Components

### 1. Main Entry Point (`main.go`)

The single entry point that initializes all systems:

```go
func main() {
    // Initialize logging
    logging.InitLogger()
    
    // Setup session store
    sessionStore, _ := http.NewFilesystemSessionStore("./sessions", true)
    
    // Configure HTTP
    config := &http.RequestConfig{
        SessionStore:  sessionStore,
        DefaultDBName: "goodoo",
        Logger:       logger,
    }
    
    // Setup Echo server with middleware
    e := echo.New()
    e.Use(http.RequestMiddleware(config))
    e.Use(logging.PerformanceMiddleware())
    
    // Register handlers
    handlers.RegisterAPIRoutes(e)
    
    // Start server
    e.Start(":8080")
}
```

### 2. Handlers (`handlers/`)

All HTTP request handlers organized by functionality:

- **API Handlers** (`api.go`) - Handle API method calls
- **Auth Handlers** (`http.go`) - Handle authentication
- **Database Handlers** (`database.go`) - Database management
- **Health Handlers** (`health.go`) - System health checks
- **Session Handlers** (`session.go`) - Session management

### 3. Database (`database/`)

Database connection management and configuration:

```go
// Initialize database connection
db, err := database.Connect(config)

// Use connection pool
pool := database.NewPool(config)
```

### 4. Tests (`tests/`)

Comprehensive tests and examples:

- **Integration Example** - Full system demonstration
- **API Tests** - API decorator system testing
- **Field Tests** - Field validation and conversion
- **Usage Examples** - Component usage demonstrations

## 🎨 API System

Define API methods with decorators:

```go
import "goodoo/api"

// Model method
api.NewMethod("partner", "search", searchHandler).
    Model().
    Help("Search partners").
    Register()

// Record method with constraints  
api.NewMethod("partner", "validate_email", validateHandler).
    Constrains("email").
    Help("Validate email format").
    Register()

// Private method
api.NewMethod("partner", "internal_cleanup", cleanupHandler).
    Model().
    Private().
    Register()
```

### HTTP API Usage

```bash
# List partner methods
curl "http://localhost:8080/api/models/partner/methods"

# Search partners
curl "http://localhost:8080/api/models/partner/search?limit=10"

# Generic API call
curl -X POST "http://localhost:8080/api/call" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "partner",
       "method": "create",
       "args": [{"name": "John Doe", "email": "john@example.com"}]
     }'
```

## 🔧 Field System

Define typed fields with validation:

```go
import "goodoo/fields"

// String field with validation
nameField, _ := fields.CreateField(fields.StringType, fields.FieldAttribute{
    String:   "Name",
    Required: true,
    Size:     255,
})

// Selection field
statusField, _ := fields.CreateField(fields.SelectionType, fields.FieldAttribute{
    String:  "Status",
    Default: "draft",
})
statusField.(*fields.SelectionField).SetSelection([]fields.SelectionOption{
    {Value: "draft", Label: "Draft"},
    {Value: "active", Label: "Active"},
})
```

## 📊 Model System

Create models with field definitions:

```go
import "goodoo/models"

// Create model
model := models.NewModelDefinition("partner", "partners")
model.Description = "Business Partner"

// Add fields
model.AddField("name", nameField)
model.AddField("email", emailField)

// Register model
models.RegisterFieldModel(model)

// Validate data
err := model.ValidateData(map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com",
})
```

## 🔒 Security Features

- **Session-based Authentication** - Secure session management
- **Access Control** - Role-based permissions for API methods
- **Request Validation** - Automatic field validation
- **Security Headers** - XSS, CSRF, and clickjacking protection
- **Private Methods** - Methods not accessible via RPC

## 📈 Performance & Monitoring

- **Request Tracking** - Unique ID for every request
- **Performance Metrics** - Query time and count monitoring
- **Colored Logging** - Visual performance indicators
- **Health Checks** - System status and metrics
- **Memory Monitoring** - Runtime memory statistics

## 🎛️ Configuration

### Environment Variables

```bash
# Logging Configuration
GOODOO_LOG_LEVEL=debug|info|warn|error|critical
GOODOO_LOG_FILE=/path/to/logfile
GOODOO_LOG_DB=log_database_name
GOODOO_COLORS=0|1

# HTTP Configuration  
GOODOO_SESSION_DIR=/path/to/sessions
GOODOO_DEFAULT_DB=default_database
PORT=8080

# Database Configuration
DATABASE_URL=postgres://user:pass@host:port/dbname
DB_MAX_CONNECTIONS=10
DB_MAX_IDLE=5
```

### Programmatic Configuration

```go
// Logging configuration
config := &logging.Config{
    Level:      logging.INFO,
    Filename:   "/var/log/goodoo.log",
    MaxSize:    100,
    MaxBackups: 5,
    Colors:     true,
}

// HTTP configuration
httpConfig := &http.RequestConfig{
    SessionStore:      sessionStore,
    DefaultDBName:     "production", 
    SessionCookieName: "goodoo_session",
    Logger:           logger,
}
```

## 🚀 Development

### Adding New Handlers

1. Create handler file in `handlers/`
2. Implement handler struct with methods
3. Register routes in main.go or handler

### Adding New API Methods

1. Define method with `api.NewMethod()`
2. Add appropriate decorators
3. Register with `.Register()`

### Adding New Fields

1. Implement field interface in `fields/`
2. Add to field registry
3. Update model definitions

### Running Tests

```bash
# Individual tests
go run tests/test_fields.go
go run tests/test_api.go

# Integration test
go run tests/integration_example.go

# All tests
go test ./...
```

## 📚 Documentation

- `README_API.md` - Comprehensive API system documentation
- `README_INTEGRATION.md` - Integration guide and examples
- Code comments - Inline documentation for all components

## 🔗 Architecture Benefits

1. **Modular Design** - Clear separation of concerns
2. **Odoo Compatibility** - Familiar patterns for Odoo developers  
3. **Type Safety** - Go's compile-time type checking
4. **Performance** - Compiled Go performance with monitoring
5. **Scalability** - Stateless design with external session storage
6. **Maintainability** - Well-organized structure with clear interfaces

This reorganized structure provides a clean, maintainable codebase with clear separation between entry point, business logic, database operations, HTTP handling, and testing.