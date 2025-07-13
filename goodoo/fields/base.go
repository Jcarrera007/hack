package fields

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"goodoo/logging"
)

// FieldType represents the type of a field
type FieldType string

const (
	// Basic field types
	BooleanType   FieldType = "boolean"
	IntegerType   FieldType = "integer"
	FloatType     FieldType = "float"
	StringType    FieldType = "char"
	TextType      FieldType = "text"
	DateType      FieldType = "date"
	DatetimeType  FieldType = "datetime"
	BinaryType    FieldType = "binary"
	SelectionType FieldType = "selection"
	
	// Special types
	JsonType       FieldType = "json"
	MonetaryType   FieldType = "monetary"
	IdType         FieldType = "id"
	
	// Relational types
	Many2oneType  FieldType = "many2one"
	One2manyType  FieldType = "one2many"
	Many2manyType FieldType = "many2many"
)

// FieldAttribute represents field attributes/properties
type FieldAttribute struct {
	String       string                 `json:"string,omitempty"`        // Field label
	Help         string                 `json:"help,omitempty"`          // Tooltip text
	Required     bool                   `json:"required,omitempty"`      // Is field required
	Readonly     bool                   `json:"readonly,omitempty"`      // Is field readonly
	Invisible    bool                   `json:"invisible,omitempty"`     // Is field invisible
	Store        bool                   `json:"store"`                   // Is field stored in DB
	Copy         bool                   `json:"copy"`                    // Copy on duplicate
	Index        string                 `json:"index,omitempty"`         // Index type (btree, trigram, etc.)
	Default      interface{}            `json:"default,omitempty"`       // Default value
	Groups       []string               `json:"groups,omitempty"`        // Access groups
	States       map[string]interface{} `json:"states,omitempty"`        // State-based conditions
	Depends      []string               `json:"depends,omitempty"`       // Computed field dependencies
	Domain       interface{}            `json:"domain,omitempty"`        // Field domain
	Context      map[string]interface{} `json:"context,omitempty"`       // Field context
	Translate    bool                   `json:"translate,omitempty"`     // Is field translatable
}

// DefaultFieldAttributes returns default field attributes
func DefaultFieldAttributes() FieldAttribute {
	return FieldAttribute{
		Store:    true,
		Copy:     true,
		Required: false,
		Readonly: false,
		Groups:   []string{},
		States:   make(map[string]interface{}),
		Context:  make(map[string]interface{}),
	}
}

// Field interface defines the contract for all field types
type Field interface {
	// Basic field operations
	GetType() FieldType
	GetName() string
	GetAttributes() FieldAttribute
	SetName(name string)
	
	// Value conversion and validation
	ConvertToCache(value interface{}, record interface{}) (interface{}, error)
	ConvertToColumn(value interface{}, record interface{}) (interface{}, error)
	ConvertToRecord(value interface{}, record interface{}) (interface{}, error)
	ConvertToExport(value interface{}, record interface{}) (interface{}, error)
	ConvertToDisplay(value interface{}, record interface{}) (string, error)
	
	// Validation
	Validate(value interface{}, record interface{}) error
	
	// SQL operations
	GetColumnType() (string, string) // (postgres_type, go_type)
	GetSQLConstraints() []string
	
	// Metadata
	IsStored() bool
	IsRequired() bool
	IsReadonly() bool
	GetDefault() interface{}
}

// BaseField provides common functionality for all field types
type BaseField struct {
	Name       string         `json:"name"`
	Type       FieldType      `json:"type"`
	Attributes FieldAttribute `json:"attributes"`
	Logger     *logging.Logger
}

// NewBaseField creates a new base field
func NewBaseField(fieldType FieldType, attrs FieldAttribute) *BaseField {
	// Merge with defaults
	defaultAttrs := DefaultFieldAttributes()
	if attrs.String == "" {
		attrs.String = defaultAttrs.String
	}
	if attrs.Groups == nil {
		attrs.Groups = defaultAttrs.Groups
	}
	if attrs.States == nil {
		attrs.States = defaultAttrs.States
	}
	if attrs.Context == nil {
		attrs.Context = defaultAttrs.Context
	}
	
	return &BaseField{
		Type:       fieldType,
		Attributes: attrs,
		Logger:     logging.GetLogger("goodoo.fields"),
	}
}

// GetType returns the field type
func (f *BaseField) GetType() FieldType {
	return f.Type
}

// GetName returns the field name
func (f *BaseField) GetName() string {
	return f.Name
}

// SetName sets the field name
func (f *BaseField) SetName(name string) {
	f.Name = name
	if f.Attributes.String == "" {
		f.Attributes.String = name
	}
}

// GetAttributes returns field attributes
func (f *BaseField) GetAttributes() FieldAttribute {
	return f.Attributes
}

// IsStored returns whether the field is stored
func (f *BaseField) IsStored() bool {
	return f.Attributes.Store
}

// IsRequired returns whether the field is required
func (f *BaseField) IsRequired() bool {
	return f.Attributes.Required
}

// IsReadonly returns whether the field is readonly
func (f *BaseField) IsReadonly() bool {
	return f.Attributes.Readonly
}

// GetDefault returns the default value
func (f *BaseField) GetDefault() interface{} {
	return f.Attributes.Default
}

// GetSQLConstraints returns SQL constraints for the field
func (f *BaseField) GetSQLConstraints() []string {
	var constraints []string
	
	if f.IsRequired() {
		constraints = append(constraints, "NOT NULL")
	}
	
	return constraints
}

// ValidateRequired checks if required field has a value
func (f *BaseField) ValidateRequired(value interface{}) error {
	if !f.IsRequired() {
		return nil
	}
	
	if value == nil {
		return fmt.Errorf("field '%s' is required", f.Name)
	}
	
	// Check for empty values based on type
	switch v := value.(type) {
	case string:
		if v == "" {
			return fmt.Errorf("field '%s' is required", f.Name)
		}
	case []interface{}:
		if len(v) == 0 {
			return fmt.Errorf("field '%s' is required", f.Name)
		}
	}
	
	return nil
}

// ConvertToDisplay provides a default string representation
func (f *BaseField) ConvertToDisplay(value interface{}, record interface{}) (string, error) {
	if value == nil {
		return "", nil
	}
	return fmt.Sprintf("%v", value), nil
}

// FieldRegistry manages field type registration and creation
type FieldRegistry struct {
	fields map[FieldType]func(FieldAttribute) Field
	logger *logging.Logger
}

// NewFieldRegistry creates a new field registry
func NewFieldRegistry() *FieldRegistry {
	registry := &FieldRegistry{
		fields: make(map[FieldType]func(FieldAttribute) Field),
		logger: logging.GetLogger("goodoo.fields.registry"),
	}
	
	// Register built-in field types
	registry.registerBuiltinFields()
	
	return registry
}

// RegisterField registers a new field type
func (r *FieldRegistry) RegisterField(fieldType FieldType, factory func(FieldAttribute) Field) {
	r.fields[fieldType] = factory
	r.logger.Debug("Registered field type: %s", fieldType)
}

// CreateField creates a field of the specified type
func (r *FieldRegistry) CreateField(fieldType FieldType, attrs FieldAttribute) (Field, error) {
	factory, exists := r.fields[fieldType]
	if !exists {
		return nil, fmt.Errorf("unknown field type: %s", fieldType)
	}
	
	field := factory(attrs)
	r.logger.Debug("Created field of type: %s", fieldType)
	
	return field, nil
}

// GetAvailableTypes returns all registered field types
func (r *FieldRegistry) GetAvailableTypes() []FieldType {
	var types []FieldType
	for fieldType := range r.fields {
		types = append(types, fieldType)
	}
	return types
}

// registerBuiltinFields registers all built-in field types
func (r *FieldRegistry) registerBuiltinFields() {
	r.RegisterField(BooleanType, func(attrs FieldAttribute) Field {
		return NewBooleanField(attrs)
	})
	
	r.RegisterField(IntegerType, func(attrs FieldAttribute) Field {
		return NewIntegerField(attrs)
	})
	
	r.RegisterField(FloatType, func(attrs FieldAttribute) Field {
		return NewFloatField(attrs)
	})
	
	r.RegisterField(StringType, func(attrs FieldAttribute) Field {
		return NewStringField(attrs)
	})
	
	r.RegisterField(TextType, func(attrs FieldAttribute) Field {
		return NewTextField(attrs)
	})
	
	r.RegisterField(DateType, func(attrs FieldAttribute) Field {
		return NewDateField(attrs)
	})
	
	r.RegisterField(DatetimeType, func(attrs FieldAttribute) Field {
		return NewDatetimeField(attrs)
	})
	
	r.RegisterField(SelectionType, func(attrs FieldAttribute) Field {
		return NewSelectionField(attrs)
	})
	
	r.RegisterField(BinaryType, func(attrs FieldAttribute) Field {
		return NewBinaryField(attrs)
	})
	
	r.RegisterField(JsonType, func(attrs FieldAttribute) Field {
		return NewJsonField(attrs)
	})
}

// Global field registry instance
var DefaultFieldRegistry = NewFieldRegistry()

// CreateField creates a field using the default registry
func CreateField(fieldType FieldType, attrs FieldAttribute) (Field, error) {
	return DefaultFieldRegistry.CreateField(fieldType, attrs)
}

// Utility functions for type conversion

// ConvertToString safely converts any value to string
func ConvertToString(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	case time.Time:
		return v.Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ConvertToInt safely converts any value to int
func ConvertToInt(value interface{}) (int, error) {
	if value == nil {
		return 0, nil
	}
	
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("cannot convert %T to int", value)
	}
}

// ConvertToFloat safely converts any value to float64
func ConvertToFloat(value interface{}) (float64, error) {
	if value == nil {
		return 0.0, nil
	}
	
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0.0, fmt.Errorf("cannot convert %T to float64", value)
	}
}

// ConvertToBool safely converts any value to bool
func ConvertToBool(value interface{}) (bool, error) {
	if value == nil {
		return false, nil
	}
	
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		return strconv.ParseBool(v)
	case int, int8, int16, int32, int64:
		rv := reflect.ValueOf(v)
		return rv.Int() != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		rv := reflect.ValueOf(v)
		return rv.Uint() != 0, nil
	case float32, float64:
		rv := reflect.ValueOf(v)
		return rv.Float() != 0.0, nil
	default:
		return false, fmt.Errorf("cannot convert %T to bool", value)
	}
}