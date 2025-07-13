package models

import (
	"fmt"
	"reflect"
	"strings"

	"goodoo/fields"
	"goodoo/logging"
	"gorm.io/gorm"
)

// ModelDefinition represents a Goodoo model with field definitions
type ModelDefinition struct {
	Name        string                     `json:"name"`
	TableName   string                     `json:"table_name"`
	Description string                     `json:"description"`
	Fields      map[string]fields.Field    `json:"fields"`
	Logger      *logging.Logger            `json:"-"`
	DB          *gorm.DB                   `json:"-"`
	
	// Model configuration
	AutoCreate  bool                       `json:"auto_create"`  // Auto-create table
	Transient   bool                       `json:"transient"`    // Don't persist to DB
	Abstract    bool                       `json:"abstract"`     // Abstract model
	Inherits    []string                   `json:"inherits"`     // Inherited models
}

// NewModelDefinition creates a new model definition
func NewModelDefinition(name, tableName string) *ModelDefinition {
	if tableName == "" {
		// Convert CamelCase to snake_case for table name
		tableName = toSnakeCase(name)
	}
	
	model := &ModelDefinition{
		Name:       name,
		TableName:  tableName,
		Fields:     make(map[string]fields.Field),
		Logger:     logging.GetLogger(fmt.Sprintf("goodoo.models.%s", name)),
		AutoCreate: true,
		Transient:  false,
		Abstract:   false,
		Inherits:   []string{},
	}
	
	// Add default fields (like Odoo's BaseModel)
	model.addDefaultFields()
	
	return model
}

// addDefaultFields adds the standard Odoo fields
func (m *ModelDefinition) addDefaultFields() {
	// ID field (not required for input as it's auto-generated)
	idField, _ := fields.CreateField(fields.IntegerType, fields.FieldAttribute{
		String:   "ID",
		Required: false, // Not required for user input
		Readonly: true,
		Store:    true,
		Copy:     false,
	})
	idField.SetName("id")
	m.Fields["id"] = idField
	
	// Create user field
	createUIDField, _ := fields.CreateField(fields.IntegerType, fields.FieldAttribute{
		String:   "Created by",
		Readonly: true,
		Store:    true,
		Copy:     false,
		Default:  1, // Default to admin user
	})
	createUIDField.SetName("create_uid")
	m.Fields["create_uid"] = createUIDField
	
	// Write user field
	writeUIDField, _ := fields.CreateField(fields.IntegerType, fields.FieldAttribute{
		String:   "Last Updated by",
		Readonly: true,
		Store:    true,
		Copy:     false,
		Default:  1,
	})
	writeUIDField.SetName("write_uid")
	m.Fields["write_uid"] = writeUIDField
	
	// Create date field
	createDateField, _ := fields.CreateField(fields.DatetimeType, fields.FieldAttribute{
		String:   "Created on",
		Readonly: true,
		Store:    true,
		Copy:     false,
	})
	createDateField.SetName("create_date")
	m.Fields["create_date"] = createDateField
	
	// Write date field
	writeDateField, _ := fields.CreateField(fields.DatetimeType, fields.FieldAttribute{
		String:   "Last Updated on",
		Readonly: true,
		Store:    true,
		Copy:     false,
	})
	writeDateField.SetName("write_date")
	m.Fields["write_date"] = writeDateField
}

// AddField adds a field to the model
func (m *ModelDefinition) AddField(name string, field fields.Field) {
	field.SetName(name)
	m.Fields[name] = field
	m.Logger.Debug("Added field %s to model %s", name, m.Name)
}

// GetField retrieves a field by name
func (m *ModelDefinition) GetField(name string) (fields.Field, bool) {
	field, exists := m.Fields[name]
	return field, exists
}

// GetFieldNames returns all field names
func (m *ModelDefinition) GetFieldNames() []string {
	names := make([]string, 0, len(m.Fields))
	for name := range m.Fields {
		names = append(names, name)
	}
	return names
}

// GetStoredFields returns only stored fields
func (m *ModelDefinition) GetStoredFields() map[string]fields.Field {
	stored := make(map[string]fields.Field)
	for name, field := range m.Fields {
		if field.IsStored() {
			stored[name] = field
		}
	}
	return stored
}

// ValidateData validates data against model fields
func (m *ModelDefinition) ValidateData(data map[string]interface{}) error {
	for fieldName, field := range m.Fields {
		value, exists := data[fieldName]
		
		// Check required fields
		if field.IsRequired() && (!exists || value == nil) {
			return fmt.Errorf("field '%s' is required", fieldName)
		}
		
		// Validate field value if present
		if exists {
			if err := field.Validate(value, nil); err != nil {
				return fmt.Errorf("validation error for field '%s': %w", fieldName, err)
			}
		}
	}
	
	return nil
}

// ConvertData converts data using field converters
func (m *ModelDefinition) ConvertData(data map[string]interface{}, conversionType string) (map[string]interface{}, error) {
	converted := make(map[string]interface{})
	
	for fieldName, value := range data {
		field, exists := m.GetField(fieldName)
		if !exists {
			// Skip unknown fields
			continue
		}
		
		var convertedValue interface{}
		var err error
		
		switch conversionType {
		case "cache":
			convertedValue, err = field.ConvertToCache(value, nil)
		case "column":
			convertedValue, err = field.ConvertToColumn(value, nil)
		case "record":
			convertedValue, err = field.ConvertToRecord(value, nil)
		case "export":
			convertedValue, err = field.ConvertToExport(value, nil)
		default:
			convertedValue = value
		}
		
		if err != nil {
			return nil, fmt.Errorf("conversion error for field '%s': %w", fieldName, err)
		}
		
		converted[fieldName] = convertedValue
	}
	
	return converted, nil
}

// GetCreateSchema returns SQL DDL for creating the table
func (m *ModelDefinition) GetCreateSchema() string {
	if m.Transient || m.Abstract {
		return ""
	}
	
	var columns []string
	
	// Add stored fields
	for name, field := range m.GetStoredFields() {
		pgType, _ := field.GetColumnType()
		column := fmt.Sprintf("%s %s", name, pgType)
		
		// Add constraints
		constraints := field.GetSQLConstraints()
		if len(constraints) > 0 {
			column += " " + strings.Join(constraints, " ")
		}
		
		columns = append(columns, column)
	}
	
	// Add primary key
	columns = append(columns, "PRIMARY KEY (id)")
	
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);",
		m.TableName,
		strings.Join(columns, ",\n  "))
}

// GetDefaultValues returns default values for all fields
func (m *ModelDefinition) GetDefaultValues() map[string]interface{} {
	defaults := make(map[string]interface{})
	
	for name, field := range m.Fields {
		if defaultValue := field.GetDefault(); defaultValue != nil {
			defaults[name] = defaultValue
		}
	}
	
	return defaults
}

// GetFieldsInfo returns field information for API responses
func (m *ModelDefinition) GetFieldsInfo() map[string]interface{} {
	fieldsInfo := make(map[string]interface{})
	
	for name, field := range m.Fields {
		attrs := field.GetAttributes()
		
		fieldInfo := map[string]interface{}{
			"type":        field.GetType(),
			"string":      attrs.String,
			"help":        attrs.Help,
			"required":    attrs.Required,
			"readonly":    attrs.Readonly,
			"store":       attrs.Store,
			"copy":        attrs.Copy,
			"default":     attrs.Default,
			"groups":      attrs.Groups,
			"states":      attrs.States,
			"domain":      attrs.Domain,
			"context":     attrs.Context,
			"translate":   attrs.Translate,
		}
		
		// Add field-specific information
		switch f := field.(type) {
		case *fields.StringField:
			fieldInfo["size"] = f.Size
		case *fields.FloatField:
			if f.Digits != nil {
				fieldInfo["digits"] = []int{f.Digits.Total, f.Digits.Decimal}
			}
		case *fields.SelectionField:
			fieldInfo["selection"] = f.Selection
		}
		
		fieldsInfo[name] = fieldInfo
	}
	
	return fieldsInfo
}

// FieldModelRegistry manages model registration and creation
type FieldModelRegistry struct {
	models map[string]*ModelDefinition
	logger *logging.Logger
}

// NewFieldModelRegistry creates a new field model registry
func NewFieldModelRegistry() *FieldModelRegistry {
	return &FieldModelRegistry{
		models: make(map[string]*ModelDefinition),
		logger: logging.GetLogger("goodoo.models.registry"),
	}
}

// RegisterModel registers a model in the registry
func (r *FieldModelRegistry) RegisterModel(model *ModelDefinition) {
	r.models[model.Name] = model
	r.logger.Info("Registered model: %s", model.Name)
}

// GetModel retrieves a model by name
func (r *FieldModelRegistry) GetModel(name string) (*ModelDefinition, bool) {
	model, exists := r.models[name]
	return model, exists
}

// GetAllModels returns all registered models
func (r *FieldModelRegistry) GetAllModels() map[string]*ModelDefinition {
	return r.models
}

// CreateTables creates database tables for all models
func (r *FieldModelRegistry) CreateTables(db *gorm.DB) error {
	for _, model := range r.models {
		if model.AutoCreate && !model.Transient && !model.Abstract {
			schema := model.GetCreateSchema()
			if schema != "" {
				if err := db.Exec(schema).Error; err != nil {
					r.logger.Error("Failed to create table for model %s: %v", model.Name, err)
					return err
				}
				r.logger.Info("Created table for model: %s", model.Name)
			}
		}
	}
	
	return nil
}

// Global field model registry
var DefaultFieldModelRegistry = NewFieldModelRegistry()

// RegisterFieldModel registers a model in the default registry
func RegisterFieldModel(model *ModelDefinition) {
	DefaultFieldModelRegistry.RegisterModel(model)
}

// GetFieldModel retrieves a model from the default registry
func GetFieldModel(name string) (*ModelDefinition, bool) {
	return DefaultFieldModelRegistry.GetModel(name)
}

// Utility functions

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// toCamelCase converts snake_case to CamelCase
func toCamelCase(str string) string {
	parts := strings.Split(str, "_")
	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			result.WriteString(strings.ToUpper(part[:1]))
			if len(part) > 1 {
				result.WriteString(part[1:])
			}
		}
	}
	return result.String()
}

// getStructFields extracts field definitions from a struct using reflection
func getStructFields(structType reflect.Type) map[string]fields.Field {
	fieldMap := make(map[string]fields.Field)
	
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Get field name from tag or use field name
		fieldName := field.Tag.Get("goodoo")
		if fieldName == "" {
			fieldName = toSnakeCase(field.Name)
		}
		
		// Skip ignored fields
		if fieldName == "-" {
			continue
		}
		
		// Create field based on Go type
		goodooField := createFieldFromGoType(field)
		if goodooField != nil {
			goodooField.SetName(fieldName)
			fieldMap[fieldName] = goodooField
		}
	}
	
	return fieldMap
}

// createFieldFromGoType creates a Goodoo field from a Go struct field
func createFieldFromGoType(structField reflect.StructField) fields.Field {
	attrs := fields.DefaultFieldAttributes()
	
	// Parse field attributes from tags
	if label := structField.Tag.Get("label"); label != "" {
		attrs.String = label
	}
	if help := structField.Tag.Get("help"); help != "" {
		attrs.Help = help
	}
	if required := structField.Tag.Get("required"); required == "true" {
		attrs.Required = true
	}
	if readonly := structField.Tag.Get("readonly"); readonly == "true" {
		attrs.Readonly = true
	}
	
	// Map Go types to Goodoo field types
	switch structField.Type.Kind() {
	case reflect.Bool:
		field, _ := fields.CreateField(fields.BooleanType, attrs)
		return field
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		 reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		field, _ := fields.CreateField(fields.IntegerType, attrs)
		return field
	case reflect.Float32, reflect.Float64:
		field, _ := fields.CreateField(fields.FloatType, attrs)
		return field
	case reflect.String:
		field, _ := fields.CreateField(fields.StringType, attrs)
		return field
	default:
		// Handle complex types
		if structField.Type.String() == "time.Time" {
			field, _ := fields.CreateField(fields.DatetimeType, attrs)
			return field
		}
		if structField.Type.Kind() == reflect.Slice && structField.Type.Elem().Kind() == reflect.Uint8 {
			// []byte
			field, _ := fields.CreateField(fields.BinaryType, attrs)
			return field
		}
	}
	
	return nil
}

// CreateModelFromStruct creates a model definition from a Go struct
func CreateModelFromStruct(name string, structType reflect.Type) *ModelDefinition {
	model := NewModelDefinition(name, "")
	
	// Extract fields from struct
	structFields := getStructFields(structType)
	
	// Add extracted fields to model
	for fieldName, field := range structFields {
		model.AddField(fieldName, field)
	}
	
	return model
}