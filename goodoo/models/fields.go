package models

import (
	"time"
	"gorm.io/gorm"
)

// Field types representing Odoo field equivalents

// CharField represents a text field with limited length
type CharField struct {
	Value    string `gorm:"type:varchar(255)" json:"value"`
	Required bool   `json:"-"`
	Size     int    `json:"-"`
}

// TextField represents a text field with unlimited length
type TextField struct {
	Value    string `gorm:"type:text" json:"value"`
	Required bool   `json:"-"`
}

// IntegerField represents an integer field
type IntegerField struct {
	Value    int  `json:"value"`
	Required bool `json:"-"`
}

// FloatField represents a floating point field
type FloatField struct {
	Value     float64 `json:"value"`
	Required  bool    `json:"-"`
	Precision int     `json:"-"`
	Scale     int     `json:"-"`
}

// BooleanField represents a boolean field
type BooleanField struct {
	Value   bool `json:"value"`
	Default bool `json:"-"`
}

// DateField represents a date field
type DateField struct {
	Value    *time.Time `gorm:"type:date" json:"value"`
	Required bool       `json:"-"`
}

// DateTimeField represents a datetime field
type DateTimeField struct {
	Value    *time.Time `gorm:"type:timestamp" json:"value"`
	Required bool       `json:"-"`
}

// SelectionField represents a selection field (enum)
type SelectionField struct {
	Value     string            `json:"value"`
	Selection map[string]string `json:"-"`
	Required  bool              `json:"-"`
}

// Many2OneField represents a many-to-one relationship
type Many2OneField struct {
	ID       *uint  `json:"id"`
	Model    string `json:"-"`
	Required bool   `json:"-"`
}

// One2ManyField represents a one-to-many relationship
type One2ManyField struct {
	IDs        []uint `gorm:"-" json:"ids"`
	Model      string `json:"-"`
	ForeignKey string `json:"-"`
}

// Many2ManyField represents a many-to-many relationship
type Many2ManyField struct {
	IDs        []uint `gorm:"many2many" json:"ids"`
	Model      string `json:"-"`
	JoinTable  string `json:"-"`
}

// BinaryField represents a binary/file field
type BinaryField struct {
	Value    []byte `gorm:"type:bytea" json:"-"`
	Filename string `json:"filename"`
	Mimetype string `json:"mimetype"`
}

// MonetaryField represents a monetary field
type MonetaryField struct {
	Value      float64 `gorm:"type:decimal(16,2)" json:"value"`
	Currency   string  `gorm:"size:3" json:"currency"`
	Required   bool    `json:"-"`
}

// HTMLField represents an HTML field
type HTMLField struct {
	Value    string `gorm:"type:text" json:"value"`
	Required bool   `json:"-"`
	Sanitize bool   `json:"-"`
}

// JSONField represents a JSON field
type JSONField struct {
	Value    interface{} `gorm:"type:jsonb" json:"value"`
	Required bool        `json:"-"`
}

// Validation interface for field validation
type Validator interface {
	Validate() error
}

// Implement Validator for each field type as needed

// BeforeCreate hook for validation
func (c *CharField) BeforeCreate(tx *gorm.DB) error {
	if c.Required && c.Value == "" {
		return gorm.ErrInvalidValue
	}
	if c.Size > 0 && len(c.Value) > c.Size {
		return gorm.ErrInvalidValue
	}
	return nil
}

func (t *TextField) BeforeCreate(tx *gorm.DB) error {
	if t.Required && t.Value == "" {
		return gorm.ErrInvalidValue
	}
	return nil
}

func (i *IntegerField) BeforeCreate(tx *gorm.DB) error {
	if i.Required && i.Value == 0 {
		return gorm.ErrInvalidValue
	}
	return nil
}

func (s *SelectionField) BeforeCreate(tx *gorm.DB) error {
	if s.Required && s.Value == "" {
		return gorm.ErrInvalidValue
	}
	if s.Selection != nil {
		if _, exists := s.Selection[s.Value]; !exists {
			return gorm.ErrInvalidValue
		}
	}
	return nil
}

func (m *Many2OneField) BeforeCreate(tx *gorm.DB) error {
	if m.Required && m.ID == nil {
		return gorm.ErrInvalidValue
	}
	return nil
}