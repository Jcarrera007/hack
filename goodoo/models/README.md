# Goodoo Models

This package provides a GORM-based implementation of Odoo's ORM functionality in Go. It recreates the core concepts from Odoo's `models.py` using Go structs and GORM.

## Core Concepts

### BaseModel
All models inherit from `BaseModel` which provides:
- `ID`: Primary key (auto-increment)
- `CreateUID`, `WriteUID`: User tracking
- `CreateDate`, `WriteDate`: Timestamp tracking
- `DeletedAt`: Soft delete support

### RecordSet
The `RecordSet[T]` type represents a collection of records and provides Odoo-like operations:
- `Search(domain, offset, limit, order)`: Find records
- `Create(records)`: Create new records
- `Read(fields)`: Read specific fields
- `Write(values)`: Update records
- `Unlink()`: Delete records
- `Count(domain)`: Count matching records

### Environment
The `Environment` provides execution context similar to Odoo's `env`:
- Database connection
- Current user information

### Model Registry
The `ModelRegistry` manages all registered models and provides:
- Model registration
- Model instantiation
- RecordSet creation

## Field Types

The package provides Go equivalents of Odoo field types:

- `CharField`: Limited text field
- `TextField`: Unlimited text field
- `IntegerField`: Integer values
- `FloatField`: Floating point values
- `BooleanField`: Boolean values
- `DateField`: Date values
- `DateTimeField`: DateTime values
- `SelectionField`: Enumeration field
- `Many2OneField`: Foreign key relationship
- `One2ManyField`: Reverse foreign key relationship
- `Many2ManyField`: Many-to-many relationship
- `BinaryField`: File/binary data
- `MonetaryField`: Currency values
- `HTMLField`: HTML content
- `JSONField`: JSON data

## Relationships

### Many2One
```go
type User struct {
    BaseModel
    PartnerID *uint   `gorm:"index"`
    Partner   Partner `gorm:"foreignKey:PartnerID"`
}
```

### One2Many
```go
type Partner struct {
    BaseModel
    Users []User `gorm:"foreignKey:PartnerID"`
}
```

### Many2Many
```go
type Product struct {
    BaseModel
    Suppliers []Partner `gorm:"many2many:product_supplierinfo;"`
}
```

## Usage Examples

### Basic CRUD Operations

```go
// Setup
db := // ... initialize GORM DB
env := NewEnvironment(db, userID)
registry := GetRegistry()
registry.SetEnvironment(env)

// Create a new user
userRS := Model(env, User{})
newUsers, err := userRS.Create([]User{
    {Name: "John Doe", Email: "john@example.com", Login: "john"},
})

// Search for users
domain := Domain{
    []interface{}{"active", "=", true},
    []interface{}{"email", "like", "%@example.com"},
}
users, err := userRS.Search(domain, 0, 10, "name")

// Read specific fields
userData, err := users.Read([]string{"name", "email"})

// Update records
err = users.Write(map[string]interface{}{
    "active": false,
})

// Delete records
err = users.Unlink()
```

### Working with Relationships

```go
// Create a partner with users
partner := Partner{
    Name:  "Acme Corp",
    Email: "info@acme.com",
}

partnerRS := Model(env, Partner{})
createdPartners, err := partnerRS.Create([]Partner{partner})

// Add users to the partner
userRS := Model(env, User{})
users, err := userRS.Create([]User{
    {Name: "John Doe", Email: "john@acme.com", PartnerID: &createdPartners.records[0].ID},
    {Name: "Jane Doe", Email: "jane@acme.com", PartnerID: &createdPartners.records[0].ID},
})
```

### Domain Queries

Domains use the same format as Odoo:
```go
domain := Domain{
    []interface{}{"field_name", "operator", "value"},
    []interface{}{"other_field", ">=", 100},
}
```

Supported operators:
- `=`, `!=`: Equality/inequality
- `>`, `>=`, `<`, `<=`: Comparison
- `like`, `ilike`: Pattern matching
- `in`, `not in`: List membership

## Model Registration

Register your models with the registry:

```go
func init() {
    registry := GetRegistry()
    registry.Register("res.users", User{})
    registry.Register("res.partner", Partner{})
    // ... other models
}
```

## Database Migration

Use GORM's AutoMigrate for database schema creation:

```go
err := db.AutoMigrate(
    &User{},
    &Partner{},
    &Product{},
    &ProductCategory{},
    &SaleOrder{},
    &SaleOrderLine{},
)
```

## Validation

Fields support automatic validation through GORM hooks:
- Required field validation
- Size constraints
- Selection value validation
- Custom validation logic

## Thread Safety

The model registry is thread-safe and can be used concurrently. Individual RecordSets are not thread-safe and should not be shared between goroutines.

## Performance Considerations

- Use `Select()` to limit loaded fields
- Use pagination with `offset` and `limit`
- Leverage GORM's preloading for relationships
- Use indexes on frequently queried fields

## Differences from Odoo

- Go's type system requires explicit struct definitions
- Relationships are handled through GORM conventions
- No computed fields (implement as methods)
- No onchange methods (implement in business logic)
- No access control (implement in application layer)