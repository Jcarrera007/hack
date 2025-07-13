# Goodoo API System

A comprehensive API system based on Odoo's API decorators, providing method exposure, access control, and validation for the Goodoo framework.

## 🎯 Overview

The Goodoo API system brings Odoo's powerful API decorator patterns to Go, enabling:

- **Method Decorators** - Mark methods with behavior like `@api.model`, `@api.constrains`, `@api.onchange`
- **Access Control** - Role-based access control with user groups
- **Automatic Validation** - Built-in field validation using the field system
- **HTTP Integration** - RESTful API endpoints with Echo framework
- **Method Metadata** - Rich method information for API documentation

## 🏗️ Architecture

The API system consists of several components:

### Core Components

1. **APIRegistry** - Manages method registration and execution
2. **MethodBuilder** - Fluent interface for defining API methods with decorators
3. **APIHandler** - HTTP handlers for API endpoints
4. **Method Types** - Model, Record, Create, and Private method types

### Decorator System

```go
// Define API methods using decorators
api.NewMethod("partner", "search", searchHandler).
    Model().                              // @api.model
    Help("Search partners").
    Register()

api.NewMethod("partner", "validate_email", validateHandler).
    Model().
    Constrains("email").                  // @api.constrains
    Help("Validate email format").
    Register()

api.NewMethod("partner", "archive", archiveHandler).
    Help("Archive partners").
    Register()                            // Record method (default)
```

## 📋 Method Types

### Model Methods (`@api.model`)
Static methods that operate on the model class rather than specific records.

```go
api.NewMethod("partner", "search", func(ctx context.Context, model *models.ModelDefinition, domain []interface{}, limit int) ([]map[string]interface{}, error) {
    // Search implementation
    return results, nil
}).Model().Register()
```

**HTTP Usage:**
```bash
# Call via URL
GET /api/models/partner/search?limit=10&offset=0

# Call via JSON API
POST /api/call
{
    "model": "partner",
    "method": "search", 
    "args": [[], 10, 0]
}
```

### Record Methods (Default)
Methods that operate on specific record IDs.

```go
api.NewMethod("partner", "archive", func(ctx context.Context, ids []int) (bool, error) {
    // Archive specific records
    return true, nil
}).Register()
```

**HTTP Usage:**
```bash
# Archive records with IDs 1,2,3
POST /api/models/partner/1,2,3/archive

# Via JSON API
POST /api/call
{
    "model": "partner",
    "method": "archive",
    "ids": [1, 2, 3]
}
```

### Create Methods
Special methods for creating records with automatic validation.

```go
api.NewMethod("partner", "create", func(ctx context.Context, model *models.ModelDefinition, vals map[string]interface{}) (map[string]interface{}, error) {
    // Data is automatically validated before this handler is called
    return createdRecord, nil
}).Model().Register()
```

### Private Methods (`@api.private`)
Methods not accessible via RPC.

```go
api.NewMethod("partner", "internal_cleanup", func(ctx context.Context, model *models.ModelDefinition) error {
    // Internal method
    return nil
}).Model().Private().Register()
```

## 🎨 Decorators

### Constrains (`@api.constrains`)
Specify field dependencies for constraint validation.

```go
api.NewMethod("partner", "validate_data", handler).
    Constrains("email", "phone").
    Register()
```

### Depends (`@api.depends`)
Specify field dependencies for computed fields.

```go
api.NewMethod("partner", "compute_full_name", handler).
    Depends("first_name", "last_name").
    Register()
```

### OnChange (`@api.onchange`)
Define onchange methods that trigger when fields change.

```go
api.NewMethod("partner", "onchange_country", handler).
    OnChange("country_id").
    Register()
```

### Returns
Specify the model type returned by the method.

```go
api.NewMethod("partner", "get_invoices", handler).
    Returns("account.invoice").
    Register()
```

### Groups
Restrict access to specific user groups.

```go
api.NewMethod("partner", "delete_partner", handler).
    Groups("base.group_admin").
    Register()
```

### Context
Add context variables to method execution.

```go
api.NewMethod("partner", "special_search", handler).
    Context(map[string]interface{}{
        "active_test": false,
        "lang": "en_US",
    }).
    Register()
```

## 🌐 HTTP API Endpoints

### Generic API Call
**POST** `/api/call`

Execute any registered API method.

```json
{
    "model": "partner",
    "method": "search",
    "args": [[], 10, 0],
    "kwargs": {"order": "name"},
    "context": {"lang": "en_US"},
    "ids": [1, 2, 3]  // For record methods
}
```

**Response:**
```json
{
    "success": true,
    "result": [...],
    "error": null,
    "warning": null
}
```

### Model Method Calls
**GET/POST** `/api/models/{model}/{method}`

Call model-level methods directly via URL.

```bash
# GET with query parameters
GET /api/models/partner/search?limit=10&domain=[]

# POST with JSON body
POST /api/models/partner/create
{
    "args": [{"name": "John Doe", "email": "john@example.com"}]
}
```

### Record Method Calls  
**GET/POST** `/api/models/{model}/{ids}/{method}`

Call record-level methods on specific IDs.

```bash
# Archive partners 1, 2, and 3
POST /api/models/partner/1,2,3/archive

# Send email to partner 5
POST /api/models/partner/5/send_email
{
    "args": ["Subject", "Body content"]
}
```

### Method Information
**GET** `/api/models/{model}/methods`

List all public methods for a model.

**GET** `/api/models/{model}/methods/{method}`

Get detailed information about a specific method.

## 🛡️ Security Features

### Access Control
- **Public/Private Methods** - Control RPC accessibility
- **User Groups** - Role-based access control
- **Authentication** - Session-based authentication required

### Validation
- **Automatic Field Validation** - Create methods validate data using model fields
- **Parameter Validation** - Type checking and constraint validation
- **Context Security** - Secure context propagation

### Error Handling
- **Structured Errors** - Consistent error responses
- **Access Denied** - Clear 403 responses for unauthorized access
- **Method Not Found** - 404 responses for missing methods

## 📖 Examples

### Complete Partner API

```go
package main

import (
    "context"
    "goodoo/api"
    "goodoo/models"
)

func registerPartnerAPI() {
    // Search partners
    api.NewMethod("partner", "search", func(ctx context.Context, model *models.ModelDefinition, domain []interface{}, limit int, offset int) ([]map[string]interface{}, error) {
        // Implementation
        return results, nil
    }).Model().
        Help("Search partners with domain filters").
        Register()

    // Create partner  
    api.NewMethod("partner", "create", func(ctx context.Context, model *models.ModelDefinition, vals map[string]interface{}) (map[string]interface{}, error) {
        // Validation happens automatically
        return createdPartner, nil
    }).Model().
        Help("Create a new partner").
        Register()

    // Archive partners
    api.NewMethod("partner", "archive", func(ctx context.Context, ids []int) (bool, error) {
        // Archive specific records
        return true, nil
    }).Help("Archive selected partners").
        Register()

    // Email validation with constraints
    api.NewMethod("partner", "validate_email", func(ctx context.Context, model *models.ModelDefinition, email string) (bool, error) {
        // Validation logic
        return isValid, nil
    }).Model().
        Constrains("email").
        Help("Validate email address").
        Register()

    // OnChange method
    api.NewMethod("partner", "onchange_country", func(ctx context.Context, ids []int, countryID int) (map[string]interface{}, error) {
        return map[string]interface{}{
            "value": map[string]interface{}{
                "state_id": false,  // Clear state when country changes
            },
        }, nil
    }).OnChange("country_id").
        Help("Update state when country changes").
        Register()

    // Admin-only method
    api.NewMethod("partner", "merge_partners", func(ctx context.Context, ids []int, targetID int) (bool, error) {
        // Merge logic
        return true, nil
    }).Groups("base.group_admin").
        Help("Merge duplicate partners").
        Register()
}
```

### HTTP Server Setup

```go
package main

import (
    "github.com/labstack/echo/v4"
    "goodoo/api"
    "goodoo/http"
)

func main() {
    // Register API methods
    registerPartnerAPI()

    // Setup HTTP server
    e := echo.New()
    
    // Add Goodoo middleware
    e.Use(http.RequestMiddleware(config))
    
    // Register API routes
    api.RegisterAPIRoutes(e)
    
    // Start server
    e.Start(":8080")
}
```

### Usage Examples

```bash
# Get all partner methods
curl "http://localhost:8080/api/models/partner/methods"

# Search partners
curl "http://localhost:8080/api/models/partner/search?limit=5"

# Create partner
curl -X POST "http://localhost:8080/api/models/partner/create" \
     -H "Content-Type: application/json" \
     -d '{"args": [{"name": "John Doe", "email": "john@example.com"}]}'

# Archive partners
curl -X POST "http://localhost:8080/api/models/partner/1,2,3/archive"

# Generic API call
curl -X POST "http://localhost:8080/api/call" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "partner",
       "method": "search", 
       "args": [[], 10, 0],
       "kwargs": {"order": "name ASC"}
     }'
```

## 🚀 Integration with Goodoo Framework

The API system seamlessly integrates with other Goodoo components:

- **Models** - Automatic validation using field definitions
- **HTTP** - Session-based authentication and request context
- **Logging** - Comprehensive logging with performance tracking
- **Fields** - Type conversion and validation for method parameters

## 🔧 Advanced Features

### Custom Decorators
Extend the decorator system for custom behaviors:

```go
// Custom decorator for caching
func (b *MethodBuilder) Cached(duration time.Duration) *MethodBuilder {
    // Implementation
    return b
}

// Usage
api.NewMethod("partner", "expensive_computation", handler).
    Model().
    Cached(5*time.Minute).
    Register()
```

### Middleware Integration
Add custom middleware for API calls:

```go
// Custom API middleware
func customAPIMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
    return func(c echo.Context) error {
        // Custom logic
        return next(c)
    }
}
```

### Performance Monitoring
Built-in performance tracking for all API calls:

- Request duration
- Parameter validation time
- Method execution time
- Response serialization time

## 🎯 Best Practices

1. **Method Naming** - Use clear, descriptive method names
2. **Parameter Validation** - Validate all input parameters
3. **Error Handling** - Return structured errors with helpful messages
4. **Documentation** - Add help text to all public methods
5. **Security** - Use appropriate access controls and groups
6. **Performance** - Consider caching for expensive operations

This comprehensive API system provides a powerful, Odoo-compatible interface for building robust Go applications with rich method exposure and access control capabilities.