package fields

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

// BooleanField represents a boolean field (like Odoo's Boolean field)
type BooleanField struct {
	*BaseField
}

// NewBooleanField creates a new boolean field
func NewBooleanField(attrs FieldAttribute) Field {
	if attrs.Default == nil {
		attrs.Default = false
	}
	
	field := &BooleanField{
		BaseField: NewBaseField(BooleanType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *BooleanField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return false, nil
	}
	
	converted, err := ConvertToBool(value)
	if err != nil {
		return false, fmt.Errorf("boolean field '%s': %w", f.Name, err)
	}
	
	return converted, nil
}

// ConvertToColumn converts value for database column
func (f *BooleanField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *BooleanField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *BooleanField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToDisplay converts value to display string
func (f *BooleanField) ConvertToDisplay(value interface{}, record interface{}) (string, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return "", err
	}
	
	if converted.(bool) {
		return "True", nil
	}
	return "False", nil
}

// Validate validates the boolean value
func (f *BooleanField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *BooleanField) GetColumnType() (string, string) {
	return "boolean", "bool"
}

// IntegerField represents an integer field (like Odoo's Integer field)
type IntegerField struct {
	*BaseField
}

// NewIntegerField creates a new integer field
func NewIntegerField(attrs FieldAttribute) Field {
	if attrs.Default == nil {
		attrs.Default = 0
	}
	
	field := &IntegerField{
		BaseField: NewBaseField(IntegerType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *IntegerField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return 0, nil
	}
	
	converted, err := ConvertToInt(value)
	if err != nil {
		return 0, fmt.Errorf("integer field '%s': %w", f.Name, err)
	}
	
	return converted, nil
}

// ConvertToColumn converts value for database column
func (f *IntegerField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *IntegerField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return nil, err
	}
	
	// Return 0 for nil to match Odoo behavior
	if converted == nil {
		return 0, nil
	}
	
	return converted, nil
}

// ConvertToExport converts value for export
func (f *IntegerField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return "", err
	}
	
	if converted.(int) == 0 && value == nil {
		return "", nil
	}
	
	return converted, nil
}

// Validate validates the integer value
func (f *IntegerField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *IntegerField) GetColumnType() (string, string) {
	return "integer", "int"
}

// FloatField represents a float field (like Odoo's Float field)
type FloatField struct {
	*BaseField
	Digits *FloatDigits `json:"digits,omitempty"` // Precision settings
}

// FloatDigits represents float precision settings
type FloatDigits struct {
	Total   int `json:"total"`   // Total number of digits
	Decimal int `json:"decimal"` // Number of decimal places
}

// NewFloatField creates a new float field
func NewFloatField(attrs FieldAttribute) Field {
	if attrs.Default == nil {
		attrs.Default = 0.0
	}
	
	field := &FloatField{
		BaseField: NewBaseField(FloatType, attrs),
	}
	
	return field
}

// SetDigits sets the precision digits for the float field
func (f *FloatField) SetDigits(total, decimal int) {
	f.Digits = &FloatDigits{
		Total:   total,
		Decimal: decimal,
	}
}

// ConvertToCache converts value for caching
func (f *FloatField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return 0.0, nil
	}
	
	converted, err := ConvertToFloat(value)
	if err != nil {
		return 0.0, fmt.Errorf("float field '%s': %w", f.Name, err)
	}
	
	// Apply precision if configured
	if f.Digits != nil {
		precision := 1.0
		for i := 0; i < f.Digits.Decimal; i++ {
			precision *= 10
		}
		converted = float64(int(converted*precision+0.5)) / precision
	}
	
	return converted, nil
}

// ConvertToColumn converts value for database column
func (f *FloatField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *FloatField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *FloatField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// Validate validates the float value
func (f *FloatField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *FloatField) GetColumnType() (string, string) {
	if f.Digits != nil {
		return fmt.Sprintf("numeric(%d,%d)", f.Digits.Total, f.Digits.Decimal), "float64"
	}
	return "double precision", "float64"
}

// StringField represents a string/char field (like Odoo's Char field)
type StringField struct {
	*BaseField
	Size int `json:"size,omitempty"` // Maximum length
}

// NewStringField creates a new string field
func NewStringField(attrs FieldAttribute) Field {
	field := &StringField{
		BaseField: NewBaseField(StringType, attrs),
		Size:      255, // Default size
	}
	
	return field
}

// SetSize sets the maximum size for the string field
func (f *StringField) SetSize(size int) {
	f.Size = size
}

// ConvertToCache converts value for caching
func (f *StringField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return "", nil
	}
	
	converted := ConvertToString(value)
	
	// Truncate if too long
	if f.Size > 0 && len(converted) > f.Size {
		converted = converted[:f.Size]
	}
	
	return converted, nil
}

// ConvertToColumn converts value for database column
func (f *StringField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *StringField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *StringField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// Validate validates the string value
func (f *StringField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return err
	}
	
	str := converted.(string)
	if f.Size > 0 && len(str) > f.Size {
		return fmt.Errorf("field '%s' exceeds maximum length of %d characters", f.Name, f.Size)
	}
	
	return nil
}

// GetColumnType returns the PostgreSQL column type
func (f *StringField) GetColumnType() (string, string) {
	if f.Size > 0 {
		return fmt.Sprintf("varchar(%d)", f.Size), "string"
	}
	return "text", "string"
}

// TextField represents a text field (like Odoo's Text field)
type TextField struct {
	*BaseField
}

// NewTextField creates a new text field
func NewTextField(attrs FieldAttribute) Field {
	field := &TextField{
		BaseField: NewBaseField(TextType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *TextField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return "", nil
	}
	
	return ConvertToString(value), nil
}

// ConvertToColumn converts value for database column
func (f *TextField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *TextField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *TextField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// Validate validates the text value
func (f *TextField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *TextField) GetColumnType() (string, string) {
	return "text", "string"
}

// DateField represents a date field (like Odoo's Date field)
type DateField struct {
	*BaseField
}

// NewDateField creates a new date field
func NewDateField(attrs FieldAttribute) Field {
	field := &DateField{
		BaseField: NewBaseField(DateType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *DateField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case time.Time:
		// Store as date only (remove time component)
		return time.Date(v.Year(), v.Month(), v.Day(), 0, 0, 0, 0, time.UTC), nil
	case string:
		// Parse date string
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			// Try datetime format and extract date
			parsed, err = time.Parse("2006-01-02 15:04:05", v)
			if err != nil {
				return nil, fmt.Errorf("invalid date format for field '%s': %s", f.Name, v)
			}
		}
		return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
	default:
		return nil, fmt.Errorf("cannot convert %T to date for field '%s'", value, f.Name)
	}
}

// ConvertToColumn converts value for database column
func (f *DateField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *DateField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *DateField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return nil, err
	}
	
	if converted == nil {
		return "", nil
	}
	
	date := converted.(time.Time)
	return date.Format("2006-01-02"), nil
}

// ConvertToDisplay converts value to display string
func (f *DateField) ConvertToDisplay(value interface{}, record interface{}) (string, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return "", err
	}
	
	if converted == nil {
		return "", nil
	}
	
	date := converted.(time.Time)
	return date.Format("2006-01-02"), nil
}

// Validate validates the date value
func (f *DateField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *DateField) GetColumnType() (string, string) {
	return "date", "time.Time"
}

// DatetimeField represents a datetime field (like Odoo's Datetime field)
type DatetimeField struct {
	*BaseField
}

// NewDatetimeField creates a new datetime field
func NewDatetimeField(attrs FieldAttribute) Field {
	field := &DatetimeField{
		BaseField: NewBaseField(DatetimeType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *DatetimeField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case time.Time:
		return v.UTC(), nil
	case string:
		// Try different datetime formats
		formats := []string{
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000Z",
			time.RFC3339,
		}
		
		for _, format := range formats {
			if parsed, err := time.Parse(format, v); err == nil {
				return parsed.UTC(), nil
			}
		}
		
		return nil, fmt.Errorf("invalid datetime format for field '%s': %s", f.Name, v)
	default:
		return nil, fmt.Errorf("cannot convert %T to datetime for field '%s'", value, f.Name)
	}
}

// ConvertToColumn converts value for database column
func (f *DatetimeField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *DatetimeField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *DatetimeField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return nil, err
	}
	
	if converted == nil {
		return "", nil
	}
	
	datetime := converted.(time.Time)
	return datetime.Format("2006-01-02 15:04:05"), nil
}

// ConvertToDisplay converts value to display string
func (f *DatetimeField) ConvertToDisplay(value interface{}, record interface{}) (string, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return "", err
	}
	
	if converted == nil {
		return "", nil
	}
	
	datetime := converted.(time.Time)
	return datetime.Format("2006-01-02 15:04:05"), nil
}

// Validate validates the datetime value
func (f *DatetimeField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *DatetimeField) GetColumnType() (string, string) {
	return "timestamp", "time.Time"
}

// SelectionField represents a selection field (like Odoo's Selection field)
type SelectionField struct {
	*BaseField
	Selection []SelectionOption `json:"selection"`
}

// SelectionOption represents a selection option
type SelectionOption struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// NewSelectionField creates a new selection field
func NewSelectionField(attrs FieldAttribute) Field {
	field := &SelectionField{
		BaseField: NewBaseField(SelectionType, attrs),
		Selection: []SelectionOption{},
	}
	
	return field
}

// SetSelection sets the selection options
func (f *SelectionField) SetSelection(options []SelectionOption) {
	f.Selection = options
}

// AddOption adds a selection option
func (f *SelectionField) AddOption(value, label string) {
	f.Selection = append(f.Selection, SelectionOption{
		Value: value,
		Label: label,
	})
}

// ConvertToCache converts value for caching
func (f *SelectionField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	strValue := ConvertToString(value)
	
	// Validate against selection options
	for _, option := range f.Selection {
		if option.Value == strValue {
			return strValue, nil
		}
	}
	
	return nil, fmt.Errorf("invalid selection value '%s' for field '%s'", strValue, f.Name)
}

// ConvertToColumn converts value for database column
func (f *SelectionField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *SelectionField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *SelectionField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToDisplay converts value to display string
func (f *SelectionField) ConvertToDisplay(value interface{}, record interface{}) (string, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return "", err
	}
	
	if converted == nil {
		return "", nil
	}
	
	strValue := converted.(string)
	
	// Find label for value
	for _, option := range f.Selection {
		if option.Value == strValue {
			return option.Label, nil
		}
	}
	
	return strValue, nil
}

// Validate validates the selection value
func (f *SelectionField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *SelectionField) GetColumnType() (string, string) {
	return "varchar(255)", "string"
}

// BinaryField represents a binary field (like Odoo's Binary field)
type BinaryField struct {
	*BaseField
}

// NewBinaryField creates a new binary field
func NewBinaryField(attrs FieldAttribute) Field {
	field := &BinaryField{
		BaseField: NewBaseField(BinaryType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *BinaryField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case []byte:
		return v, nil
	case string:
		// Assume base64 encoded
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("invalid base64 data for binary field '%s': %w", f.Name, err)
		}
		return decoded, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to binary for field '%s'", value, f.Name)
	}
}

// ConvertToColumn converts value for database column
func (f *BinaryField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToRecord converts value for record
func (f *BinaryField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *BinaryField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return nil, err
	}
	
	if converted == nil {
		return "", nil
	}
	
	bytes := converted.([]byte)
	return base64.StdEncoding.EncodeToString(bytes), nil
}

// Validate validates the binary value
func (f *BinaryField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *BinaryField) GetColumnType() (string, string) {
	return "bytea", "[]byte"
}

// JsonField represents a JSON field (like Odoo's Json field)
type JsonField struct {
	*BaseField
}

// NewJsonField creates a new JSON field
func NewJsonField(attrs FieldAttribute) Field {
	field := &JsonField{
		BaseField: NewBaseField(JsonType, attrs),
	}
	
	return field
}

// ConvertToCache converts value for caching
func (f *JsonField) ConvertToCache(value interface{}, record interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case string:
		// Parse JSON string
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, fmt.Errorf("invalid JSON for field '%s': %w", f.Name, err)
		}
		return parsed, nil
	case map[string]interface{}, []interface{}:
		// Already in correct format
		return v, nil
	default:
		// Try to marshal and unmarshal to validate JSON
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return nil, fmt.Errorf("cannot convert to JSON for field '%s': %w", f.Name, err)
		}
		
		var parsed interface{}
		if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
			return nil, fmt.Errorf("invalid JSON conversion for field '%s': %w", f.Name, err)
		}
		
		return parsed, nil
	}
}

// ConvertToColumn converts value for database column
func (f *JsonField) ConvertToColumn(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return nil, err
	}
	
	if converted == nil {
		return nil, nil
	}
	
	// Convert to JSON string for database storage
	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON for field '%s': %w", f.Name, err)
	}
	
	return string(jsonBytes), nil
}

// ConvertToRecord converts value for record
func (f *JsonField) ConvertToRecord(value interface{}, record interface{}) (interface{}, error) {
	return f.ConvertToCache(value, record)
}

// ConvertToExport converts value for export
func (f *JsonField) ConvertToExport(value interface{}, record interface{}) (interface{}, error) {
	converted, err := f.ConvertToCache(value, record)
	if err != nil {
		return nil, err
	}
	
	if converted == nil {
		return "", nil
	}
	
	jsonBytes, err := json.Marshal(converted)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON for export field '%s': %w", f.Name, err)
	}
	
	return string(jsonBytes), nil
}

// Validate validates the JSON value
func (f *JsonField) Validate(value interface{}, record interface{}) error {
	if err := f.ValidateRequired(value); err != nil {
		return err
	}
	
	_, err := f.ConvertToCache(value, record)
	return err
}

// GetColumnType returns the PostgreSQL column type
func (f *JsonField) GetColumnType() (string, string) {
	return "jsonb", "interface{}"
}